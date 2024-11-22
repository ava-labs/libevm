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
	"testing"
	"time"

	"github.com/ava-labs/libevm/common"
)

type synchronisingWorkerPool struct {
	executed, unblock chan struct{}
}

var _ WorkerPool = (*synchronisingWorkerPool)(nil)

func (p *synchronisingWorkerPool) Execute(func()) {
	select {
	case <-p.executed:
	default:
		close(p.executed)
	}
}

func (p *synchronisingWorkerPool) Wait() {
	<-p.unblock
}

func TestStopPrefetcherWaitsOnWorkers(t *testing.T) {
	pool := &synchronisingWorkerPool{
		executed: make(chan struct{}),
		unblock:  make(chan struct{}),
	}
	opt := WithWorkerPools(func() WorkerPool { return pool })

	db := filledStateDB()
	db.prefetcher = newTriePrefetcher(db.db, db.originalRoot, "", opt)
	db.prefetcher.prefetch(common.Hash{}, common.Hash{}, common.Address{}, [][]byte{{}})

	go func() {
		<-pool.executed
		// Sleep otherwise there is a small chance that we close pool.unblock
		// between db.StopPrefetcher() returning and the select receiving on the
		// channel.
		time.Sleep(time.Second)
		close(pool.unblock)
	}()

	<-pool.executed
	db.StopPrefetcher()
	select {
	case <-pool.unblock:
		// The channel was closed, therefore pool.Wait() unblocked. This is a
		// necessary pre-condition for db.StopPrefetcher() unblocking, and the
		// purpose of this test.
	default:
		t.Errorf("%T.StopPrefetcher() returned before %T.Wait() unblocked", db, pool)
	}
}
