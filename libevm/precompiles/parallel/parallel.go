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

// Package parallel provides functionality for precompiled contracts that can
// pre-process their results in an embarrassingly parallel fashion.
package parallel

import (
	"errors"
	"fmt"
	"sync"

	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/libevm/core"
	"github.com/ava-labs/libevm/core/state"
	"github.com/ava-labs/libevm/core/types"
	"github.com/ava-labs/libevm/core/vm"
	"github.com/ava-labs/libevm/libevm"
	"github.com/ava-labs/libevm/libevm/stateconf"
	"github.com/ava-labs/libevm/params"
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
type Handler[Data, Result any] interface {
	BeforeBlock(libevm.StateReader, *types.Block)
	Gas(*types.Transaction) (gas uint64, process bool)
	Prefetch(sdb libevm.StateReader, index int, tx *types.Transaction) Data
	Process(sdb libevm.StateReader, index int, tx *types.Transaction, data Data) Result
	AfterBlock(StateDB, *types.Block, types.Receipts)
}

// StateDB is the subset of [state.StateDB] methods that MAY be called by
// [Handler.AfterBlock].
type StateDB interface {
	libevm.StateReader
	SetState(_ common.Address, key, val common.Hash, _ ...stateconf.StateDBStateOption)
}

// A Processor orchestrates dispatch and collection of results from a [Handler].
type Processor[D, R any] struct {
	handler           Handler[D, R]
	workers           sync.WaitGroup
	prefetch, process chan *job
	data              [](chan D)
	results           [](chan result[R])
	txGas             map[common.Hash]uint64
	stateShare        stateDBSharer
}

type job struct {
	index int
	tx    *types.Transaction
}

type result[T any] struct {
	tx  common.Hash
	val *T
}

// New constructs a new [Processor] with the specified number of concurrent
// workers. [Processor.Close] must be called after the final call to
// [Processor.FinishBlock] to avoid leaking goroutines.
func New[D, R any](h Handler[D, R], prefetchers, processors int) *Processor[D, R] {
	prefetchers = max(prefetchers, 1)
	processors = max(processors, 1)
	workers := prefetchers + processors

	p := &Processor[D, R]{
		handler:  h,
		prefetch: make(chan *job),
		process:  make(chan *job),
		txGas:    make(map[common.Hash]uint64),
		stateShare: stateDBSharer{
			workers:       workers,
			nextAvailable: make(chan struct{}),
		},
	}

	p.workers.Add(workers)       // for shutdown via [Processor.Close]
	p.stateShare.wg.Add(workers) // for readiness of [Processor.worker] loops
	for range prefetchers {
		go p.worker(p.prefetch, nil)
	}
	for range processors {
		go p.worker(nil, p.process)
	}
	p.stateShare.wg.Wait()

	return p
}

// A stateDBSharer allows concurrent workers to make copies of a primary
// database. When the `nextAvailable` channel is closed, all workers call
// [state.StateDB.Copy] then signal completion on the [sync.WaitGroup]. The
// channel is replaced for each round of distribution.
type stateDBSharer struct {
	nextAvailable chan struct{}
	primary       *state.StateDB
	mu            sync.Mutex
	workers       int
	wg            sync.WaitGroup
}

func (s *stateDBSharer) distribute(sdb *state.StateDB) {
	s.primary = sdb // no need to Copy() as each worker does it

	ch := s.nextAvailable                 // already copied by [Processor.worker], which is waiting for it to close
	s.nextAvailable = make(chan struct{}) // will be copied, ready for the next distribution

	s.wg.Add(s.workers)
	close(ch)
	s.wg.Wait()
}

func (p *Processor[D, R]) worker(prefetch, process chan *job) {
	defer p.workers.Done()

	var sdb *state.StateDB
	share := &p.stateShare
	stateAvailable := share.nextAvailable
	// Without this signal of readiness, a premature call to
	// [Processor.StartBlock] could replace `share.nextAvailable` before we've
	// copied it.
	share.wg.Done()

	for {
		select {
		case <-stateAvailable: // guaranteed at the beginning of each block
			share.mu.Lock()
			sdb = share.primary.Copy()
			share.mu.Unlock()

			stateAvailable = share.nextAvailable
			share.wg.Done()

		case job, ok := <-prefetch:
			if !ok {
				return
			}
			p.data[job.index] <- p.handler.Prefetch(sdb, job.index, job.tx)

		case job, ok := <-process:
			if !ok {
				return
			}

			r := p.handler.Process(sdb, job.index, job.tx, <-p.data[job.index])
			p.results[job.index] <- result[R]{
				tx:  job.tx.Hash(),
				val: &r,
			}
		}
	}
}

// Close shuts down the [Processor], after which it can no longer be used.
func (p *Processor[D, R]) Close() {
	close(p.prefetch)
	close(p.process)
	p.workers.Wait()
}

// StartBlock dispatches transactions to the [Handler] and returns immediately.
// It MUST be paired with a call to [Processor.FinishBlock], without overlap of
// blocks.
func (p *Processor[D, R]) StartBlock(sdb *state.StateDB, rules params.Rules, b *types.Block) error {
	// The distribution mechanism copies the StateDB so we don't need to do it
	// here, but the [Handler] is called directly so we do copy.
	p.stateShare.distribute(sdb)
	p.handler.BeforeBlock(
		sdb.Copy(),
		types.NewBlockWithHeader(
			b.Header(),
		).WithBody(
			*b.Body(),
		),
	)

	txs := b.Transactions()
	jobs := make([]*job, 0, len(txs))

	// We can reuse the channels already in the data and results slices because
	// they're emptied by [Processor.FinishBlock].
	for i, n := len(p.results), len(txs); i < n; i++ {
		p.data = append(p.data, make(chan D, 1))
		p.results = append(p.results, make(chan result[R], 1))
	}

	for i, tx := range txs {
		switch do, err := p.shouldProcess(tx, rules); {
		case err != nil:
			return err

		case do:
			jobs = append(jobs, &job{
				index: i,
				tx:    tx,
			})

		default:
			p.results[i] <- result[R]{
				tx:  tx.Hash(),
				val: nil,
			}
		}
	}

	// The first goroutine pipelines into the second, which has its results
	// emptied by [Processor.FinishBlock]. The return of said function therefore
	// guarantees that we haven't leaked either of these.
	go func() {
		for _, j := range jobs {
			p.prefetch <- j
		}
	}()
	go func() {
		for _, j := range jobs {
			p.process <- j
		}
	}()
	return nil
}

// FinishBlock returns the [Processor] to a state ready for the next block. A
// return from FinishBlock guarantees that all dispatched work from the
// respective call to [Processor.StartBlock] has been completed.
func (p *Processor[D, R]) FinishBlock(sdb vm.StateDB, b *types.Block, rs types.Receipts) {
	for i := range len(b.Transactions()) {
		// Every result channel is guaranteed to have some value in its buffer
		// because [Processor.BeforeBlock] either sends a nil *R or it
		// dispatches a job, which will send a non-nil *R.
		tx := (<-p.results[i]).tx
		delete(p.txGas, tx)
	}
	p.handler.AfterBlock(sdb, b, rs)
}

// Result blocks until the i'th transaction passed to [Processor.StartBlock] has
// had its result processed, and then returns the value returned by the
// [Handler]. The returned boolean will be false if no processing occurred,
// either because the [Handler] indicated as such or because the transaction
// supplied insufficient gas.
//
// Multiple calls to Result with the same argument are allowed. Callers MUST NOT
// charge the gas price for preprocessing as this is handled by
// [Processor.PreprocessingGasCharge] if registered as a [vm.Preprocessor].
// The same value will be returned by each call with the same argument, such
// that if R is a pointer then modifications will persist between calls.
func (p *Processor[D, R]) Result(i int) (R, bool) {
	ch := p.results[i]
	r := <-ch
	defer func() {
		ch <- r
	}()

	if r.val == nil {
		// TODO(arr4n) if we're here then the implementoor might have a bug in
		// their [Handler], so logging a warning is probably a good idea.
		var zero R
		return zero, false
	}
	return *r.val, true
}

func (p *Processor[R, D]) shouldProcess(tx *types.Transaction, rules params.Rules) (process bool, retErr error) {
	// An explicit 0 is necessary to avoid [Processor.PreprocessingGasCharge]
	// returning [ErrTxUnknown].
	p.txGas[tx.Hash()] = 0

	cost, ok := p.handler.Gas(tx)
	if !ok {
		return false, nil
	}
	defer func() {
		if process && retErr == nil {
			p.txGas[tx.Hash()] = cost
		}
	}()

	spent, err := txIntrinsicGas(tx, &rules)
	if err != nil {
		return false, fmt.Errorf("calculating intrinsic gas of %v: %v", tx.Hash(), err)
	}
	if spent > tx.Gas() {
		// If this happens then consensus has a bug because the tx shouldn't
		// have been included. We include the check, however, for completeness.
		return false, core.ErrIntrinsicGas
	}
	return tx.Gas()-spent >= cost, nil
}

func txIntrinsicGas(tx *types.Transaction, rules *params.Rules) (uint64, error) {
	return intrinsicGas(tx.Data(), tx.AccessList(), tx.To(), rules)
}

func intrinsicGas(data []byte, access types.AccessList, txTo *common.Address, rules *params.Rules) (uint64, error) {
	create := txTo == nil
	return core.IntrinsicGas(
		data,
		access,
		create,
		rules.IsHomestead,
		rules.IsIstanbul, // EIP-2028
		rules.IsShanghai, // EIP-3860
	)
}

// ErrTxUnknown is returned by [Processor.PreprocessingGasCharge] if it is
// called with a transaction hash that wasn't in the last block passed to
// [Processor.StartBlock].
var ErrTxUnknown = errors.New("transaction unknown by parallel preprocessor")

// PreprocessingGasCharge implements the [vm.Preprocessor] interface and MUST be
// registered via [vm.RegisterHooks] to ensure proper gas accounting.
func (p *Processor[R, D]) PreprocessingGasCharge(tx common.Hash) (uint64, error) {
	g, ok := p.txGas[tx]
	if !ok {
		return 0, fmt.Errorf("%w: %v", ErrTxUnknown, tx)
	}
	return g, nil
}

var _ vm.Preprocessor = (*Processor[struct{}, struct{}])(nil)
