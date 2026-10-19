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
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/holiman/uint256"

	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/libevm/core/rawdb"
	"github.com/ava-labs/libevm/core/types"
	"github.com/ava-labs/libevm/ethdb/memorydb"
	"github.com/ava-labs/libevm/rlp"
	"github.com/ava-labs/libevm/trie"
	"github.com/ava-labs/libevm/trie/trienode"
)

// keyRangesDiff returns the [cmp.Diff] of want and got.
func keyRangesDiff(want, got keyRanges) string {
	return cmp.Diff(want, got, cmp.AllowUnexported(keyRange{}))
}

func TestSkippingIteratorStepsOverKnownEmptyStretch(t *testing.T) {
	prefix := rawdb.SnapshotAccountPrefix
	entry := func(b byte) []byte {
		return append(common.CopyBytes(prefix), bytes.Repeat([]byte{b}, common.HashLength)...)
	}
	db := memorydb.New()
	first, last := entry(0x01), entry(0xfe)
	for _, key := range [][]byte{first, last} {
		if err := db.Put(key, []byte{1}); err != nil {
			t.Fatal(err)
		}
	}
	putSkippedKeys(t, db, prefix, 1_000)

	read := func(known keyRanges) (entries [][]byte, steps int64) {
		counted := newCountingDB(db, 0)
		it := newSkippingIterator(counted, counted.NewIterator(prefix, nil), prefix, nil, 1+common.HashLength, known, nil)
		defer it.Release()
		for it.Next() {
			key := it.Key()
			for _, r := range known {
				if bytes.Compare(key, r.from) >= 0 && bytes.Compare(key, r.to) <= 0 {
					t.Errorf("read %x inside known stretch [%x, %x]", key, r.from, r.to)
				}
			}
			if len(key) == 1+common.HashLength {
				entries = append(entries, common.CopyBytes(key))
			}
		}
		if err := it.Error(); err != nil {
			t.Fatal(err)
		}
		return entries, counted.steps.Load()
	}

	entries, fullSteps := read(nil)
	if diff := cmp.Diff([][]byte{first, last}, entries); diff != "" {
		t.Fatalf("entries read diff (-want +got):\n%s", diff)
	}

	// Stop short of the last entry to learn the stretch between the 2 entries.
	stretch := keyRange{from: append(common.CopyBytes(first), 0), to: nil}
	probe := newSkippingIterator(db, db.NewIterator(prefix, nil), prefix, nil, 1+common.HashLength, nil, nil)
	for probe.Next() && !bytes.Equal(probe.Key(), last) {
		if bytes.Compare(probe.Key(), first) > 0 {
			stretch.to = common.CopyBytes(probe.Key())
			stretch.keys++
		}
	}
	if err := probe.Error(); err != nil {
		t.Fatal(err)
	}
	probe.Release()
	if stretch.empty() {
		t.Fatal("found no keys to skip between the entries")
	}

	entries, skipSteps := read(keyRanges{stretch})
	if diff := cmp.Diff([][]byte{first, last}, entries); diff != "" {
		t.Fatalf("entries read while skipping diff (-want +got):\n%s", diff)
	}
	if skipSteps >= fullSteps/10 {
		t.Errorf("took %d steps with the stretch known and %d without; want the known stretch left unread", skipSteps, fullSteps)
	}
}

// TestSkippingIteratorStretchStartsAfterResumeKey stops a run before it reaches
// an entry. The run resumes at the generator's marker, whose entry the flush of
// diff layers writes, so the stretch it keeps must not include that key.
func TestSkippingIteratorStretchStartsAfterResumeKey(t *testing.T) {
	prefix := rawdb.SnapshotAccountPrefix
	db := memorydb.New()
	putSkippedKeys(t, db, prefix, 1_000)

	start := bytes.Repeat([]byte{0x80}, common.HashLength)
	var found keyRanges
	it := newSkippingIterator(db, db.NewIterator(prefix, start), prefix, start, 1+common.HashLength, nil, &found)
	defer it.Release()
	var last []byte
	for range minSkipKeys {
		if !it.Next() {
			t.Fatalf("iterator stopped before %d keys: %v", minSkipKeys, it.Error())
		}
		last = common.CopyBytes(it.Key())
	}
	it.keepRun()

	resume := append(common.CopyBytes(prefix), start...)
	want := keyRanges{{from: append(resume, 0), to: last, keys: minSkipKeys}}
	if diff := keyRangesDiff(want, found); diff != "" {
		t.Errorf("stretch kept diff (-want +got):\n%s", diff)
	}
}

func TestTrimSkipsUpToGeneratorPosition(t *testing.T) {
	stretch := func(prefix []byte, from, to byte) keyRange {
		return keyRange{
			from: append(common.CopyBytes(prefix), bytes.Repeat([]byte{from}, common.HashLength)...),
			to:   append(common.CopyBytes(prefix), bytes.Repeat([]byte{to}, common.HashLength)...),
		}
	}
	ctx := &generatorContext{generatorSkipping: generatorSkipping{skips: generatorSkips{
		account: keyRanges{stretch(rawdb.SnapshotAccountPrefix, 0x20, 0x40)},
		storage: keyRanges{stretch(rawdb.SnapshotStoragePrefix, 0x20, 0x40)},
	}}}

	current := bytes.Repeat([]byte{0x30}, 2*common.HashLength) // an account and one of its slots
	ctx.trimSkipsUpTo(current)

	for name, got := range map[string]keyRanges{"account": ctx.skips.account, "storage": ctx.skips.storage} {
		prefix := rawdb.SnapshotAccountPrefix
		written := append(common.CopyBytes(prefix), current[:common.HashLength]...)
		if name == "storage" {
			prefix = rawdb.SnapshotStoragePrefix
			written = append(common.CopyBytes(prefix), current...)
		}
		if len(got) != 1 || bytes.Compare(got[0].from, written) <= 0 {
			t.Errorf("%s stretches = %x after the generator wrote up to %x; want 1 starting later", name, got, written)
		}
	}

	ctx.trimSkipsUpTo(bytes.Repeat([]byte{0x50}, common.HashLength))
	if len(ctx.skips.account) != 0 || len(ctx.skips.storage) != 0 {
		t.Errorf("stretches not emptied once the generator passed them: account %x, storage %x", ctx.skips.account, ctx.skips.storage)
	}
}

func TestKeyRangesWith(t *testing.T) {
	key := func(b byte) []byte { return []byte{b} }
	stretch := func(from, to byte, keys int) keyRange { return keyRange{from: key(from), to: key(to), keys: keys} }

	var rs keyRanges
	rs = rs.with(stretch(0x10, 0x20, minSkipKeys-1))
	if len(rs) != 0 {
		t.Fatalf("kept a stretch of %d keys; want only stretches of at least %d", minSkipKeys-1, minSkipKeys)
	}
	rs = rs.with(stretch(0x50, 0x60, minSkipKeys))
	rs = rs.with(stretch(0x10, 0x20, minSkipKeys))
	rs = rs.with(stretch(0x18, 0x30, minSkipKeys*2)) // overlaps the first
	want := keyRanges{stretch(0x10, 0x30, minSkipKeys*2), stretch(0x50, 0x60, minSkipKeys)}
	if diff := keyRangesDiff(want, rs); diff != "" {
		t.Fatalf("stretches diff (-want +got):\n%s", diff)
	}

	rs = rs.with(stretch(0x28, 0x58, minSkipKeys)) // bridges both
	if diff := keyRangesDiff(keyRanges{stretch(0x10, 0x60, minSkipKeys*2)}, rs); diff != "" {
		t.Fatalf("stretches after bridging diff (-want +got):\n%s", diff)
	}

	rs = nil
	for i := 0; i <= maxSkipRanges; i++ {
		rs = rs.with(keyRange{from: []byte{byte(i >> 8), byte(i)}, to: []byte{byte(i >> 8), byte(i), 0xff}, keys: minSkipKeys + i})
	}
	if len(rs) != maxSkipRanges || rs[0].keys != minSkipKeys+1 {
		t.Errorf("after adding %d stretches, kept %d starting with %d keys; want %d, without the smallest", maxSkipRanges+1, len(rs), rs[0].keys, maxSkipRanges)
	}
}

func TestKeyRangesAfter(t *testing.T) {
	stretch := func(from, to byte) keyRange { return keyRange{from: []byte{from}, to: []byte{to}, keys: minSkipKeys} }
	ranges := func() keyRanges { return keyRanges{stretch(0x10, 0x20), stretch(0x30, 0x40), stretch(0x50, 0x60)} }
	rs := ranges()

	want := keyRanges{{from: []byte{0x38, 0}, to: []byte{0x40}, keys: minSkipKeys}, stretch(0x50, 0x60)}
	if diff := keyRangesDiff(want, rs.after([]byte{0x38})); diff != "" {
		t.Errorf("after(0x38) diff (-want +got):\n%s", diff)
	}
	if diff := keyRangesDiff(rs[2:], rs.after([]byte{0x40})); diff != "" {
		t.Errorf("after(0x40) diff (-want +got):\n%s", diff)
	}
	if diff := keyRangesDiff(ranges(), rs); diff != "" {
		t.Errorf("after modified its receiver; diff (-want +got):\n%s", diff)
	}
	if allocs := testing.AllocsPerRun(10, func() { rs.after([]byte{0x05}) }); allocs != 0 {
		t.Errorf("after a key before every stretch made %v allocations; want 0", allocs)
	}
}

// TestKeepReadRecordsStretchOnce reads an iterator's stretch twice, as a run
// does when an iterator runs out and then the run ends.
func TestKeepReadRecordsStretchOnce(t *testing.T) {
	ctx := &generatorContext{}
	ctx.accountRead = &skippingIterator{
		found: &ctx.skips.account,
		run:   keyRange{from: []byte{0x10}, to: []byte{0x20}, keys: minSkipKeys},
	}
	ctx.keepRead(snapAccount)
	ctx.skips.account = ctx.skips.account.after([]byte{0x18})
	ctx.keepRead(snapAccount)

	want := keyRanges{{from: []byte{0x18, 0}, to: []byte{0x20}, keys: minSkipKeys}}
	if diff := keyRangesDiff(want, ctx.skips.account); diff != "" {
		t.Errorf("stretches diff (-want +got), the trimmed stretch not recorded again:\n%s", diff)
	}
}

// TestGenerateSkipsStretchesFoundEmpty restarts every generation run a few steps
// in, as a node does on each block, and checks that only the first run reads the
// skipped keys and that no remembered stretch hides the leftover entries among them.
func TestGenerateSkipsStretchesFoundEmpty(t *testing.T) {
	helper := newHelper(rawdb.HashScheme)
	stRoot := helper.makeStorageTrie(common.Hash{}, []string{"key-1", "key-2", "key-3"}, []string{"val-1", "val-2", "val-3"}, false)
	for i := uint64(0); i < 20; i++ {
		acc := &types.StateAccount{Balance: uint256.NewInt(i), Root: types.EmptyRootHash, CodeHash: types.EmptyCodeHash.Bytes()}
		if i%4 == 0 {
			acc.Root = stRoot
			helper.makeStorageTrie(hashData([]byte(fmt.Sprintf("acc-%d", i))), []string{"key-1", "key-2", "key-3"}, []string{"val-1", "val-2", "val-3"}, true)
		}
		helper.addTrieAccount(fmt.Sprintf("acc-%d", i), acc)
		if i%3 == 0 {
			// Left over from an earlier snapshot, and wrong for this state.
			stale := *acc
			stale.Balance = uint256.NewInt(1_000)
			helper.addSnapAccount(fmt.Sprintf("acc-%d", i), &stale)
		}
	}
	for i := 0; i < 5; i++ {
		// Left over from an earlier snapshot, and absent from this state.
		helper.addSnapAccount(fmt.Sprintf("gone-%d", i), &types.StateAccount{Balance: uint256.NewInt(1), Root: types.EmptyRootHash, CodeHash: types.EmptyCodeHash.Bytes()})
		helper.addSnapStorage(fmt.Sprintf("gone-%d", i), []string{"key-1"}, []string{"val-1"})
	}
	root := helper.Commit()

	const skipped = 5_000
	putSkippedKeys(t, helper.diskdb, rawdb.SnapshotAccountPrefix, skipped)
	putSkippedKeys(t, helper.diskdb, rawdb.SnapshotStoragePrefix, skipped)

	db := newCountingDB(helper.diskdb, 500)
	layer := restartUntilDone(t, db, generateSnapshot(db, helper.triedb, 16, root), root, 100)
	steps := db.steps.Load() // before checkSnapRoot iterates the snapshot too
	checkSnapRoot(t, layer, root)
	<-layer.done // the finished run hands on its stretches as it exits
	if skips := layer.genSkips; len(skips.account)+len(skips.storage) != 0 {
		t.Errorf("%T.genSkips holds %d account and %d storage stretches after generation finished; want none", layer, len(skips.account), len(skips.storage))
	}
	// The first run has to read every key in both ranges once.
	if want := int64(3 * skipped); steps > want {
		t.Errorf("took %d iteration steps; want at most %d, reading the skipped keys about once", steps, want)
	}
}

// TestGenerateRereadsEntriesWrittenPastItsMarker has a run write account entries
// to disk past its saved marker, as a flush while deleting dangling storage does,
// and then fail. The next run, on a state that has since dropped one of those
// accounts, has to read that entry to delete it.
func TestGenerateRereadsEntriesWrittenPastItsMarker(t *testing.T) {
	helper := newHelper(rawdb.HashScheme)
	names := make([]string, 50)
	last := 0
	for i := range names {
		names[i] = fmt.Sprintf("acc-%d", i)
		if bytes.Compare(hashData([]byte(names[i])).Bytes(), hashData([]byte(names[last])).Bytes()) > 0 {
			last = i
		}
	}
	broken, gone := names[last], names[(last+1)%len(names)]
	brokenHash, goneHash := hashData([]byte(broken)), hashData([]byte(gone))

	stRoot := helper.makeStorageTrie(brokenHash, []string{"key-1"}, []string{"val-1"}, true)
	for _, name := range names {
		acc := &types.StateAccount{Balance: uint256.NewInt(1), Root: types.EmptyRootHash, CodeHash: types.EmptyCodeHash.Bytes()}
		if name == broken {
			acc.Root = stRoot
		}
		helper.addTrieAccount(name, acc)
	}
	root := helper.Commit()

	// The first run fails on the last account's missing storage, but only after
	// deleting enough dangling storage just before it to flush what it wrote.
	rawdb.DeleteTrieNode(helper.diskdb, brokenHash, nil, stRoot, rawdb.HashScheme)
	dangling := common.BytesToHash(decKey(brokenHash.Bytes()))
	for i := 0; i < 2_000; i++ {
		rawdb.WriteStorageSnapshot(helper.diskdb, dangling, hashData([]byte(fmt.Sprint(i))), []byte{1})
	}
	putSkippedKeys(t, helper.diskdb, rawdb.SnapshotAccountPrefix, 1_000)

	first := generateSnapshot(helper.diskdb, helper.triedb, 16, root)
	select {
	case <-first.done:
	case <-time.After(time.Minute):
		t.Fatal("first run neither failed nor finished")
	}
	var gen journalGenerator
	if err := rlp.DecodeBytes(rawdb.ReadSnapshotGenerator(helper.diskdb), &gen); err != nil {
		t.Fatal(err)
	}
	if gen.Done || rawdb.ReadAccountSnapshot(helper.diskdb, goneHash) == nil || bytes.Compare(goneHash[:], gen.Marker) <= 0 {
		t.Fatalf("first run saved marker %x (done %t); want account %x written past it", gen.Marker, gen.Done, goneHash)
	}

	tr, err := trie.NewStateTrie(trie.StateTrieID(root), helper.triedb)
	if err != nil {
		t.Fatal(err)
	}
	tr.MustDelete([]byte(gone))
	tr.MustDelete([]byte(broken))
	next, nodes, err := tr.Commit(false)
	if err != nil {
		t.Fatal(err)
	}
	if err := helper.triedb.Update(next, root, 0, trienode.NewWithNodeSet(nodes), nil); err != nil {
		t.Fatal(err)
	}
	if err := helper.triedb.Commit(next, false); err != nil {
		t.Fatal(err)
	}

	layer := diffToDisk(newDiffLayer(first, next, nil, nil, nil))
	select {
	case <-layer.genPending:
	case <-time.After(time.Minute):
		t.Fatal("second run did not finish")
	}
	checkSnapRoot(t, layer, next)
}
