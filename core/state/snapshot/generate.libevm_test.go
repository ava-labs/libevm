// Copyright 2026 the libevm authors.
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

package snapshot

import (
	"bytes"
	"fmt"
	"math/rand"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/holiman/uint256"

	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/libevm/core/rawdb"
	"github.com/ava-labs/libevm/core/types"
	"github.com/ava-labs/libevm/ethdb"
	"github.com/ava-labs/libevm/rlp"
)

// countingDB counts the steps taken by its snapshot iterators and, when every
// is set, pauses on each multiple of it so a test can stop generation there.
type countingDB struct {
	ethdb.KeyValueStore
	every  atomic.Int64
	paused chan struct{}
	resume chan struct{}
	steps  atomic.Int64

	deletedPastMarker atomic.Int64 // deletions written past the next marker journalled
	midStorageMarkers atomic.Int64 // markers journalled part way through a contract's storage
	writes            atomic.Int64
	unjournalled      [][]byte // deletions written since the last marker was journalled
}

func newCountingDB(db ethdb.KeyValueStore, every int64) *countingDB {
	c := &countingDB{KeyValueStore: db, paused: make(chan struct{}), resume: make(chan struct{})}
	c.every.Store(every)
	return c
}

func (db *countingDB) NewIterator(prefix, start []byte) ethdb.Iterator {
	it := db.KeyValueStore.NewIterator(prefix, start)
	if !bytes.Equal(prefix, rawdb.SnapshotAccountPrefix) && !bytes.Equal(prefix, rawdb.SnapshotStoragePrefix) {
		return it
	}
	return &countingIterator{Iterator: it, db: db}
}

func (db *countingDB) NewBatch() ethdb.Batch {
	return &recordingBatch{Batch: db.KeyValueStore.NewBatch(), db: db}
}

// generatorKey is the key journalProgress saves the generator's marker under.
var generatorKey = []byte("SnapshotGenerator")

// recordingBatch reports to its countingDB what each write deleted and which
// generator marker, if any, it journalled.
type recordingBatch struct {
	ethdb.Batch
	db       *countingDB
	deleted  [][]byte
	journals bool
	marker   []byte
}

func (b *recordingBatch) Put(key, value []byte) error {
	if bytes.Equal(key, generatorKey) {
		var gen journalGenerator
		if err := rlp.DecodeBytes(value, &gen); err != nil {
			return err
		}
		b.journals, b.marker = true, gen.Marker
	}
	return b.Batch.Put(key, value)
}

func (b *recordingBatch) Delete(key []byte) error {
	b.deleted = append(b.deleted, common.CopyBytes(key))
	return b.Batch.Delete(key)
}

func (b *recordingBatch) Write() error {
	b.db.writes.Add(1)
	b.db.recordWrite(b.deleted, b.journals, b.marker)
	return b.Batch.Write()
}

func (b *recordingBatch) Reset() {
	b.deleted, b.journals, b.marker = nil, false, nil
	b.Batch.Reset()
}

// recordWrite compares deletions with the next marker journalled rather than
// one in the same write, as a stop seen by an iterator leaves the marker to the
// write that follows it.
func (db *countingDB) recordWrite(deleted [][]byte, journals bool, marker []byte) {
	db.unjournalled = append(db.unjournalled, deleted...)
	if !journals {
		return
	}
	if len(marker) > common.HashLength {
		db.midStorageMarkers.Add(1)
	}
	for _, key := range db.unjournalled {
		if len(marker) > 0 && bytes.Compare(key[1:], marker) > 0 { // key[1:] drops the snapshot prefix
			db.deletedPastMarker.Add(1)
		}
	}
	db.unjournalled = nil
}

type countingIterator struct {
	ethdb.Iterator
	db *countingDB
}

func (it *countingIterator) Next() bool {
	if n, every := it.db.steps.Add(1), it.db.every.Load(); every > 0 && n%every == 0 {
		it.db.paused <- struct{}{}
		<-it.db.resume
	}
	return it.Iterator.Next()
}

// putSkippedKeys stores n uniformly distributed keys of size
// [common.HashLength], each with the specified prefix. The snapshot iterators
// cover that range with a [rawdb.KeyLengthIterator] of different length, so
// they have to step over every one of them.
func putSkippedKeys(t *testing.T, db ethdb.KeyValueWriter, prefix []byte, n uint64) {
	t.Helper()

	var key common.Hash
	copy(key[:], prefix)
	rest := key[min(len(prefix), len(key)):]

	rng := rand.New(rand.NewSource(0)) //nolint:gosec // Seeded so every run stores the same keys
	for range n {
		rng.Read(rest) //nolint:gosec // Documented to always return a nil error
		if err := db.Put(key.Bytes(), []byte{1}); err != nil {
			t.Fatalf("%T.Put(%v, ...): %v", db, key, err)
		}
	}
}

type steppingIterator struct {
	ethdb.Iterator
	step <-chan struct{}
}

func (it *steppingIterator) Next() bool {
	<-it.step
	return it.Iterator.Next()
}

type steppingIterDB struct {
	ethdb.Database
	step <-chan struct{}
}

func (db *steppingIterDB) NewIterator(prefix []byte, start []byte) ethdb.Iterator {
	return &steppingIterator{
		db.Database.NewIterator(prefix, start),
		db.step,
	}
}

func TestGenerateStopsWhileSkippingKeys(t *testing.T) {
	synctest.Test(t, func(t *testing.T) { //nolint:thelper // False positive, fixed in thelper v0.7.1.
		helper := newHelper(rawdb.HashScheme)
		// The helper doesn't require any contents because we're only concerned
		// with our ability to abort the skipping of hashdb-originated keys.
		root := helper.Commit()
		putSkippedKeys(t, helper.diskdb, rawdb.SnapshotAccountPrefix, 10_000)

		step := make(chan struct{})
		db := &steppingIterDB{helper.diskdb, step}
		snap := generateSnapshot(db, helper.triedb, 16, root)

		synctest.Wait() // generator blocked in [rawdb.KeyLengthIterator.Next]
		for range 5 {
			step <- struct{}{} // skipping is actually happening
		}
		synctest.Wait() // generator blocked as before

		stopped := make(chan struct{})
		go func() {
			snap.stopGeneration()
			close(stopped)
		}()
		synctest.Wait()    // [diskLayer.stopGeneration] blocked on <-done,
		step <- struct{}{} // allow next [abortableIterator.Next] to register cancellation
		<-stopped

		// We require an explicit `[]byte{}` as it signals that generation started
		// but made no progress. A nil marker signals completion.
		if diff := cmp.Diff([]byte{}, snap.genMarker); diff != "" {
			t.Errorf("%T.genMarker diff after stopping mid-iteration:\n%s", snap, diff)
		}
	})
}

// restartUntilDone restarts generation on a fresh disk layer, as a node does on
// each block, every time db pauses, until the snapshot is complete.
func restartUntilDone(t *testing.T, db *countingDB, layer *diskLayer, root common.Hash, maxRestarts int) *diskLayer {
	t.Helper()
	for restarts := 0; ; restarts++ {
		select {
		case <-layer.genPending:
			if restarts == 0 {
				t.Fatal("BAD TEST SETUP: finished with zero restarts")
			}
			db.every.Store(0) // the checks that follow iterate the snapshot too
			t.Logf("finished after %d restarts and %d iteration steps", restarts, db.steps.Load())
			return layer
		case <-db.paused:
		case <-time.After(time.Minute):
			t.Fatalf("generation neither paused nor finished after %d restarts", restarts)
		}
		if restarts == maxRestarts {
			close(db.resume)
			t.Fatalf("generation still unfinished after %d restarts and %d iteration steps", restarts, db.steps.Load())
		}
		next := make(chan *diskLayer)
		go func(base *diskLayer) {
			next <- diffToDisk(newDiffLayer(base, root, nil, nil, nil))
		}(layer)
		<-layer.cancel
		db.resume <- struct{}{}
		layer = <-next
	}
}

// TestGenerateKeepsProgressWhenStoppedMidRange stops generation while it reads
// a range of existing snapshot entries, after it has finished earlier ranges.
func TestGenerateKeepsProgressWhenStoppedMidRange(t *testing.T) {
	helper := newHelper(rawdb.HashScheme)
	for i := uint64(0); i < 300; i++ {
		acc := &types.StateAccount{Balance: uint256.NewInt(i), Root: types.EmptyRootHash, CodeHash: types.EmptyCodeHash.Bytes()}
		helper.addAccount(fmt.Sprintf("acc-%d", i), acc)
	}
	root := helper.Commit()
	putSkippedKeys(t, helper.diskdb, rawdb.SnapshotAccountPrefix, 3_000)

	db := newCountingDB(helper.diskdb, 2_000)
	layer := restartUntilDone(t, db, generateSnapshot(db, helper.triedb, 16, root), root, 20)
	checkSnapRoot(t, layer, root)
}

// TestGenerateKeepsDeletionsWhenStoppedMidRange stops generation while its
// batch still holds deletions of stale entries past the last finished position,
// and part way through a contract's storage, then checks the finished snapshot.
func TestGenerateKeepsDeletionsWhenStoppedMidRange(t *testing.T) {
	helper := newHelper(rawdb.HashScheme)
	slots := func(prefix string, n int) (keys, vals []string) {
		for i := 0; i < n; i++ {
			keys = append(keys, fmt.Sprintf("%s-key-%d", prefix, i))
			vals = append(vals, fmt.Sprintf("%s-val-%d", prefix, i))
		}
		return keys, vals
	}
	addContract := func(name string, n int) *types.StateAccount {
		keys, vals := slots(name, n)
		root := helper.makeStorageTrie(hashData([]byte(name)), keys, vals, true)
		acc := &types.StateAccount{Balance: uint256.NewInt(1), Root: root, CodeHash: types.EmptyCodeHash.Bytes()}
		helper.addAccount(name, acc)
		helper.addSnapStorage(name, keys, vals)
		staleKeys, staleVals := slots(name+"-stale", 3)
		helper.addSnapStorage(name, staleKeys, staleVals)
		return acc
	}
	for i := 0; i < 200; i++ {
		acc := addContract(fmt.Sprintf("acc-%d", i), 10)

		gone := fmt.Sprintf("gone-%d", i) // in the snapshot only
		keys, vals := slots(gone, 3)
		helper.addSnapAccount(gone, acc)
		helper.addSnapStorage(gone, keys, vals)

		orphan := fmt.Sprintf("orphan-%d", i) // storage without an account
		keys, vals = slots(orphan, 3)
		helper.addSnapStorage(orphan, keys, vals)
	}
	addContract("big", 3*storageCheckRange)
	root := helper.Commit()

	db := newCountingDB(helper.diskdb, 1_500)
	layer := restartUntilDone(t, db, generateSnapshot(db, helper.triedb, 16, root), root, 50)
	checkSnapRoot(t, layer, root)

	if db.deletedPastMarker.Load() == 0 {
		t.Error("no stop saved deletions past its marker; the test no longer covers that case")
	}
	if db.midStorageMarkers.Load() == 0 {
		t.Error("no stop landed inside a contract's storage; the test no longer covers that case")
	}
	t.Logf("%d deletions saved past a marker, %d markers inside a contract's storage", db.deletedPastMarker.Load(), db.midStorageMarkers.Load())
}
