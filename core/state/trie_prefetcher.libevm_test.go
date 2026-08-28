// Copyright 2024 the libevm authors.
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

package state

import (
	"fmt"
	"testing"
	"testing/synctest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ava-labs/libevm/common"
)

type synchronisingWorkerPool struct {
	t                             *testing.T
	executed, unblock             chan struct{}
	done                          bool
	preconditionsToStopPrefetcher int
}

var _ WorkerPool = (*synchronisingWorkerPool)(nil)

func (p *synchronisingWorkerPool) Execute(fn func()) {
	fn()
	select {
	case <-p.executed:
	default:
		close(p.executed)
	}

	<-p.unblock
	assert.False(p.t, p.done, "Done() called before Execute() returns")
	p.preconditionsToStopPrefetcher++
}

func (p *synchronisingWorkerPool) Done() {
	p.done = true
	p.preconditionsToStopPrefetcher++
}

// TestStopPrefetcherWaitsOnWorkers verifies that [triePrefetcher.StopPrefetcher]
// waits for all in-progress [WorkerPool.Execute] calls to return before it
// calls [WorkerPool.Done] and returns.
func TestStopPrefetcherWaitsOnWorkers(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		pool := &synchronisingWorkerPool{
			t:        t,
			executed: make(chan struct{}),
			unblock:  make(chan struct{}),
		}
		opt := WithWorkerPools(func() WorkerPool { return pool })

		db := filledStateDB()
		db.prefetcher = newTriePrefetcher(db.db, db.originalRoot, "", opt)
		db.prefetcher.prefetch(common.Hash{}, common.Hash{}, common.Address{}, [][]byte{{}})

		// The worker is now inside Execute(), blocked on pool.unblock.
		<-pool.executed

		stopped := make(chan struct{})
		go func() {
			db.StopPrefetcher()
			close(stopped)
		}()

		// Wait() returns once StopPrefetcher() is durably blocked on the
		// still-running Execute(), or has (incorrectly) returned.
		synctest.Wait()
		select {
		case <-stopped:
			t.Fatalf("%T.StopPrefetcher() returned while Execute() was still in progress", db)
		default:
		}

		close(pool.unblock)
		<-stopped
		// If this errors then either Execute() hadn't returned or Done() wasn't
		// called.
		assert.Equalf(t, 2, pool.preconditionsToStopPrefetcher, "%T.StopPrefetcher() returned early", db)
	})
}

type countingWorkerPool struct {
	dones int
}

var _ WorkerPool = (*countingWorkerPool)(nil)

func (p *countingWorkerPool) Execute(f func()) { f() }
func (p *countingWorkerPool) Done()            { p.dones++ }

// TestWorkerPoolSharedAcrossTries verifies that a [triePrefetcher] constructs
// exactly one [WorkerPool], shares it between all of its subfetchers, and calls
// Done() on it exactly once when closed.
func TestWorkerPoolSharedAcrossTries(t *testing.T) {
	for _, numFetchers := range []int{0, 1, 2, 5} {
		t.Run(fmt.Sprintf("%d_fetchers", numFetchers), func(t *testing.T) {
			var (
				ctorCalls int
				pool      = &countingWorkerPool{}
			)
			opt := WithWorkerPools(func() WorkerPool {
				ctorCalls++
				return pool
			})

			db := filledStateDB()
			prefetcher := newTriePrefetcher(db.db, db.originalRoot, "", opt)
			require.Equal(t, 1, ctorCalls, "pool-constructor should have been called in newTriePrefetcher()")

			for i := range numFetchers {
				owner := common.BytesToHash([]byte{byte(i)})
				addr := common.BytesToAddress([]byte{byte(i)})
				keys := [][]byte{common.HexToHash("0xdead").Bytes()}
				prefetcher.prefetch(owner, db.originalRoot, addr, keys)
			}
			require.Len(t, prefetcher.fetchers, numFetchers, "each call to prefetch() should have created a new subfetcher")
			for id, f := range prefetcher.fetchers {
				assert.Samef(t, pool, f.pool.workers, "subfetcher %x does not have the prefetcher's shared worker pool", id)
			}
			assert.Equal(t, 1, ctorCalls, "pool-constructor should not be called more than once")

			prefetcher.close()
			assert.Equal(t, 1, pool.dones, "should have called Done() once after close()")
		})
	}
}
