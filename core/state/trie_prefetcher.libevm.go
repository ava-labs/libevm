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
	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/libevm/libevm/options"
	"github.com/ava-labs/libevm/libevm/sync"
	"github.com/ava-labs/libevm/log"
)

// A PrefetcherOption configures behaviour of trie prefetching.
type PrefetcherOption = options.Option[prefetcherConfig]

type prefetcherConfig struct {
	newWorkerPool func() WorkerPool
}

// A WorkerPool executes functions asynchronously. Done() is called to signal
// that the pool is no longer needed and that Execute() is guaranteed to not be
// called again.
type WorkerPool interface {
	Execute(func())
	Done()
}

// WithWorkerPool configures trie prefetching to execute asynchronously. The
// provided constructor is called once per trie prefetcher (i.e. typically once
// per block). Done() is called on the pool once when the prefetcher is closed.
func WithWorkerPool(ctor func() WorkerPool) PrefetcherOption {
	return options.Func[prefetcherConfig](func(c *prefetcherConfig) {
		c.newWorkerPool = ctor
	})
}

// newWorkerPool returns a new [WorkerPool] if provided a [WithWorkerPool].
// Otherwise, it returns nil.
//
// The pool should be shared by all of the prefetcher's subfetchers and MUST NOT
// be propagated to copies made with [triePrefetcher.copy]. A copy MAY be closed
// while the original is still fetching.
func newWorkerPool(opts ...PrefetcherOption) WorkerPool {
	c := options.As(opts...)
	if c.newWorkerPool == nil {
		return nil
	}
	return c.newWorkerPool()
}

type subfetcherPool struct {
	workers WorkerPool
	tries   sync.Pool[Trie]
	wg      sync.WaitGroup
}

// A subfetcherPoolOption configures a [subfetcherPool].
type subfetcherPoolOption = options.Option[subfetcherPool]

// withWorkerPool configures a [subfetcherPool] to execute functions on the
// provided [WorkerPool], which MAY be nil, in which case they are executed
// synchronously.
func withWorkerPool(workers WorkerPool) subfetcherPoolOption {
	return options.Func[subfetcherPool](func(c *subfetcherPool) {
		c.workers = workers
	})
}

// initPool initialises the [subfetcher]'s pool, applying the provided options.
func (sf *subfetcher) initPool(opts ...subfetcherPoolOption) {
	sf.pool = &subfetcherPool{
		tries: sync.Pool[Trie]{
			// Although the workers are shared between all subfetchers, each
			// MUST have its own Trie pool.
			New: func() Trie {
				return sf.db.CopyTrie(sf.trie)
			},
		},
	}
	options.ApplyTo(sf.pool, opts...)
}

// releaseWorkerPool calls Done() on the shared [WorkerPool] if one was
// provided with [WithWorkerPool]. This MUST only be called after
// [subfetcher.abort] returns on ALL fetchers to guarantee that no further calls
// will be made to Execute() after calling Done().
func (p *triePrefetcher) releaseWorkerPool() {
	if w := p.workers; w != nil {
		w.Done()
	}
}

func (p *subfetcherPool) wait() {
	p.wg.Wait()
}

// execute runs the provided function with a copy of the subfetcher's Trie.
// Copies are stored in a [sync.Pool] to reduce creation overhead. If p was
// configured with a [WorkerPool] then it is used for function execution,
// otherwise `fn` is just called directly.
func (p *subfetcherPool) execute(fn func(Trie)) {
	p.wg.Add(1)
	do := func() {
		t := p.tries.Get()
		fn(t)
		p.tries.Put(t)
		p.wg.Done()
	}

	if w := p.workers; w != nil {
		w.Execute(do)
	} else {
		do()
	}
}

// GetAccount optimistically pre-fetches an account, dropping the returned value
// and logging errors. See [subfetcherPool.execute] re worker pools.
func (p *subfetcherPool) GetAccount(addr common.Address) {
	p.execute(func(t Trie) {
		if _, err := t.GetAccount(addr); err != nil {
			log.Error("account prefetching failed", "address", addr, "err", err)
		}
	})
}

// GetStorage is the storage equivalent of [subfetcherPool.GetAccount].
func (p *subfetcherPool) GetStorage(addr common.Address, key []byte) {
	p.execute(func(t Trie) {
		if _, err := t.GetStorage(addr, key); err != nil {
			log.Error("storage prefetching failed", "address", addr, "key", key, "err", err)
		}
	})
}
