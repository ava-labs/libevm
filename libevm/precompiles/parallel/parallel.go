package parallel

import (
	"fmt"
	"sync"

	"github.com/ava-labs/libevm/core"
	"github.com/ava-labs/libevm/core/types"
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
type Handler[Result any] interface {
	Gas(*types.Transaction) (gas uint64, process bool)
	Process(index int, tx *types.Transaction) Result
}

// A Processor orchestrates dispatch and collection of results from a [Handler].
type Processor[R any] struct {
	handler Handler[R]
	workers sync.WaitGroup
	work    chan *job
	results [](chan *R)
}

type job struct {
	index int
	tx    *types.Transaction
}

// New constructs a new [Processor] with the specified number of concurrent
// workers. [Processor.Close] must be called after the final call to
// [Processor.FinishBlock] to avoid leaking goroutines.
func New[R any](h Handler[R], workers int) *Processor[R] {
	p := &Processor[R]{
		handler: h,
		work:    make(chan *job),
	}

	workers = max(workers, 1)
	p.workers.Add(workers)
	for range workers {
		go p.worker()
	}
	return p
}

func (p *Processor[R]) worker() {
	defer p.workers.Done()
	for {
		w, ok := <-p.work
		if !ok {
			return
		}

		r := p.handler.Process(w.index, w.tx)
		p.results[w.index] <- &r
	}
}

// Close shuts down the [Processor], after which it can no longer be used.
func (p *Processor[R]) Close() {
	close(p.work)
	p.workers.Wait()
}

// StartBlock dispatches transactions to the [Handler] and returns immediately.
// It MUST be paired with a call to [Processor.FinishBlock], without overlap of
// blocks.
func (p *Processor[R]) StartBlock(b *types.Block) error {
	txs := b.Transactions()
	jobs := make([]*job, 0, len(txs))

	// We can reuse the channels already in the results slice because they're
	// emptied by [Processor.FinishBlock].
	for i, n := len(p.results), len(txs); i < n; i++ {
		p.results = append(p.results, make(chan *R, 1))
	}

	for i, tx := range txs {
		switch do, err := p.shouldProcess(tx); {
		case err != nil:
			return err

		case do:
			jobs = append(jobs, &job{
				index: i,
				tx:    tx,
			})

		default:
			p.results[i] <- nil
		}
	}

	go func() {
		// This goroutine is guaranteed to have returned by the time
		// [Processor.FinishBlock] does.
		for _, j := range jobs {
			p.work <- j
		}
	}()
	return nil
}

// FinishBlock returns the [Processor] to a state ready for the next block. A
// return from FinishBlock guarantees that all dispatched work from the
// respective call to [Processor.StartBlock] has been completed.
func (p *Processor[R]) FinishBlock(b *types.Block) {
	for i := range len(b.Transactions()) {
		// Every result channel is guaranteed to have some value in its buffer
		// because [Processor.BeforeBlock] either sends a nil *R or it
		// dispatches a job that will send a non-nil *R.
		<-p.results[i]
	}
}

// Result blocks until the i'th transaction passed to [Processor.StartBlock] has
// had its result processed, and then returns the value returned by the
// [Handler]. The returned boolean will be false if no processing occurred,
// either because the [Handler] indicated as such or because the transaction
// supplied insufficient gas.
func (p *Processor[R]) Result(i int) (R, bool) {
	ch := p.results[i]
	r := <-ch
	defer func() {
		ch <- r
	}()

	if r == nil {
		// TODO(arr4n) if we're here then the implementoor might have a bug in
		// their [Handler], so logging a warning is probably a good idea.
		var zero R
		return zero, false
	}
	return *r, true
}

func (p *Processor[R]) shouldProcess(tx *types.Transaction) (bool, error) {
	cost, ok := p.handler.Gas(tx)
	if !ok {
		return false, nil
	}

	spent, err := core.IntrinsicGas(
		tx.Data(),
		tx.AccessList(),
		tx.To() == nil,
		true, // Homestead
		true, // EIP-2028 (Istanbul)
		true, // EIP-3860 (Shanghai)
	)
	if err != nil {
		return false, fmt.Errorf("calculating intrinsic gas of %v: %v", tx.Hash(), err)
	}

	// This could only overflow if the gas limit was insufficient to cover
	// the intrinsic cost, which would have invalidated it for inclusion.
	left := tx.Gas() - spent
	return left >= cost, nil
}
