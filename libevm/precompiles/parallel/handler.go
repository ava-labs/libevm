// Copyright 2025 the libevm authors.
//
// The libevm additions to go-ethereum are free software: you can redistribute
// them and/or modify them under the terms of the GNU Lesser General Public License
// as published by the Free Software Foundation, either version 3 of the License,
// or (at your option) any later version.
//
// The libevm additions are distributed in the hope that they will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU Lesser
// General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see
// <http://www.gnu.org/licenses/>.

package parallel

import (
	"sync"

	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/libevm/core/types"
	"github.com/ava-labs/libevm/core/vm"
	"github.com/ava-labs/libevm/libevm"
	"github.com/ava-labs/libevm/libevm/stateconf"
)

// A Handler is responsible for processing [types.Transactions] in an
// embarrassingly parallel fashion. It is the responsibility of the Handler to
// determine whether this is possible, typically only so if one of the following
// is true with respect to a precompile associated with the Handler:
//
// 1. The destination address is that of the precompile; or
//
// 2. At least one [types.AccessTuple] references the precompile's address.
//
// Scenario (2) allows precompile access to be determined through inspection of
// the [types.Transaction] alone, without the need for execution.
//
// All [libevm.StateReader] instances are opened to the state at the beginning
// of the block. The [StateDB] is the same one used to execute the block,
// before being committed, and MAY be written to.
//
// NOTE: other than [Handler.AfterBlock], all methods MAY be called concurrently
// with one another and with other [Handler] implementations, unless otherwise
// specified. AfterBlock() methods are called in the same order as they were
// registered with [AddHandler].
type Handler[CommonData, Data, Result, Aggregated any] interface {
	// Gas reports whether the [Handler] SHOULD receive the transaction for
	// processing and, if so, how much gas to charge. Processing is performed
	// i.f.f. the returned boolean is true and there is sufficient gas limit to
	// cover intrinsic gas and all [Handler]s that returned true. If there is
	// insufficient gas for processing then the transaction will result in
	// [vm.ErrOutOfGas] as long as the [Processor] is registered with
	// [vm.RegisterHooks] as a [vm.Preprocessor].
	Gas(*types.Transaction) (gas uint64, process bool)
	// BeforeBlock is called before all calls to Prefetch() on this [Handler],
	// all of which receive the returned value.
	BeforeBlock(libevm.StateReader, *types.Block) CommonData
	// Prefetch is called before the respective call to Process() on this
	// [Handler]. It MUST NOT perform any meaningful computation beyond what is
	// necessary to determine that necessary state to propagate to Process().
	Prefetch(libevm.StateReader, IndexedTx, CommonData) Data
	// Process is responsible for performing all meaningful computation. It
	// receives the common data returned by the single call to BeforeBlock() as
	// well as the data from the respective call to Prefetch(). The returned
	// result is propagated to PostProcess() and any calls to the function
	// returned by [AddHandler].
	//
	// NOTE: if the result is exposed to the EVM via a precompile then said
	// precompile will block until Process() returns. While this guarantees the
	// availability of pre-processed results, it is also the hot path for EVM
	// transactions.
	Process(libevm.StateReader, IndexedTx, CommonData, Data) Result
	// PostProcess is called concurrently with all calls to Process(). It allows
	// for online aggregation of results into a format ready for writing to
	// state.
	PostProcess(Results[Result]) Aggregated
	// AfterBlock is called after PostProcess() returns and all regular EVM
	// transaction processing is complete.
	AfterBlock(StateDB, Aggregated, *types.Block, types.Receipts)
}

// An IndexedTx couples a [types.Transaction] with its index in a block.
type IndexedTx struct {
	Index int
	*types.Transaction
}

// Results provides mechanisms for blocking on the output of [Handler.Process].
type Results[R any] struct {
	WaitForAll            func()
	TxOrder, ProcessOrder <-chan TxResult[R]
}

// A TxResult couples an [IndexedTx] with its respective result from
// [Handler.Process].
type TxResult[R any] struct {
	Tx     IndexedTx
	Result R
}

// StateDB is the subset of [state.StateDB] methods that MAY be called by
// [Handler.AfterBlock].
type StateDB interface {
	libevm.StateReader
	SetState(_ common.Address, key, val common.Hash, _ ...stateconf.StateDBStateOption)
}

var _ handler = (*wrapper[any, any, any, any])(nil)

// A wrapper exposes the generic functionality of a [Handler] in a non-generic
// manner, allowing [Processor] to be free of type parameters.
type wrapper[CD, D, R, A any] struct {
	Handler[CD, D, R, A]

	totalTxsInBLock   int
	txsBeingProcessed sync.WaitGroup

	common eventual[CD]
	data   []eventual[D]

	results       []eventual[result[R]]
	whenProcessed chan TxResult[R]

	aggregated eventual[A]
}

// AddHandler registers the [Handler] with the [Processor] and returns a
// function to fetch the [TxResult] for the i'th transaction passed to
// [Processor.StartBlock].
//
// The returned function until the respective transaction has had its result
// processed, and then returns the value returned by the [Handler]. The returned
// boolean will be false if no processing occurred, either because the [Handler]
// indicated as such or because the transaction supplied insufficient gas.
//
// Multiple calls to Result with the same argument are allowed. Callers MUST NOT
// charge the gas price for preprocessing as this is handled by
// [Processor.PreprocessingGasCharge] if registered as a [vm.Preprocessor].
//
// Within the scope of a given block, the same value will be returned by each
// call with the same argument, such that if R is a pointer then modifications
// will persist between calls. However, the caller does NOT have mutually
// exclusive access to the [TxResult] so SHOULD NOT modify it, especially since
// the result MAY also be accessed by [Handler.PostProcess], with no ordering
// guarantees.
func AddHandler[CD, D, R, A any](p *Processor, h Handler[CD, D, R, A]) func(txIndex int) (TxResult[R], bool) {
	w := &wrapper[CD, D, R, A]{
		Handler:    h,
		common:     eventually[CD](),
		aggregated: eventually[A](),
	}
	p.handlers = append(p.handlers, w)
	return w.result
}

func (w *wrapper[CD, D, R, A]) beforeBlock(sdb libevm.StateReader, b *types.Block) {
	w.totalTxsInBLock = len(b.Transactions())
	// We can reuse the channels already in the data and results slices because
	// they're emptied by [wrapper.process] and [wrapper.finishBlock]
	// respectively.
	for i := len(w.results); i < w.totalTxsInBLock; i++ {
		w.data = append(w.data, eventually[D]())
		w.results = append(w.results, eventually[result[R]]())
	}

	go func() {
		// goroutine guaranteed to have completed by the time a respective
		// getter unblocks (i.e. in any call to [wrapper.prefetch]).
		w.common.set(w.BeforeBlock(sdb, b))
	}()
}

func (w *wrapper[SD, D, R, A]) beforeWork(jobs int) {
	w.txsBeingProcessed.Add(jobs)
	w.whenProcessed = make(chan TxResult[R], jobs)
	go func() {
		w.txsBeingProcessed.Wait()
		close(w.whenProcessed)
	}()
}

func (w *wrapper[SD, D, R, A]) prefetch(sdb libevm.StateReader, job *job) {
	w.data[job.tx.Index].set(w.Prefetch(sdb, job.tx, w.common.getAndReplace()))
}

func (w *wrapper[SD, D, R, A]) process(sdb libevm.StateReader, job *job) {
	defer w.txsBeingProcessed.Done()

	idx := job.tx.Index
	val := w.Process(sdb, job.tx, w.common.getAndReplace(), w.data[idx].getAndKeep())
	r := result[R]{
		tx:  job.tx,
		val: &val,
	}
	w.results[idx].set(r)
	w.whenProcessed <- TxResult[R]{
		Tx:     job.tx,
		Result: val,
	}
}

func (w *wrapper[SD, D, R, A]) nullResult(job *job) {
	w.results[job.tx.Index].set(result[R]{
		tx:  job.tx,
		val: nil,
	})
}

func (w *wrapper[SD, D, R, A]) result(i int) (TxResult[R], bool) {
	r := w.results[i].getAndReplace()

	txr := TxResult[R]{
		Tx: r.tx,
	}
	if r.val == nil {
		// TODO(arr4n) if we're here then the implementoor might have a bug in
		// their [Handler], so logging a warning is probably a good idea.
		return txr, false
	}
	txr.Result = *r.val
	return txr, true
}

func (w *wrapper[SD, D, R, A]) postProcess() {
	txOrder := make(chan TxResult[R], w.totalTxsInBLock)
	go func() {
		defer close(txOrder)
		for i := range w.totalTxsInBLock {
			r, ok := w.result(i)
			if !ok {
				continue
			}
			txOrder <- r
		}
	}()

	w.aggregated.set(w.PostProcess(Results[R]{
		WaitForAll:   w.txsBeingProcessed.Wait,
		TxOrder:      txOrder,
		ProcessOrder: w.whenProcessed,
	}))
}

func (p *wrapper[SD, D, R, A]) finishBlock(sdb vm.StateDB, b *types.Block, rs types.Receipts) {
	p.AfterBlock(sdb, p.aggregated.getAndKeep(), b, rs)
	p.common.getAndKeep()
	for _, v := range p.results[:p.totalTxsInBLock] {
		// Every result channel is guaranteed to have some value in its buffer
		// because [Processor.BeforeBlock] either sends a nil *R or it
		// dispatches a job, which will send a non-nil *R.
		v.getAndKeep()
	}
}
