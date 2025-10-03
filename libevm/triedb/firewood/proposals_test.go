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

package firewood

import (
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/libevm/core/rawdb"
	"github.com/ava-labs/libevm/core/state"
	"github.com/ava-labs/libevm/core/types"
	"github.com/ava-labs/libevm/ethdb"
	"github.com/ava-labs/libevm/libevm/stateconf"
	"github.com/ava-labs/libevm/trie"
	"github.com/ava-labs/libevm/trie/trienode"
	"github.com/ava-labs/libevm/trie/triestate"
	"github.com/ava-labs/libevm/triedb"
	databasepkg "github.com/ava-labs/libevm/triedb/database"
	"github.com/ava-labs/libevm/triedb/hashdb"
)

func TestMain(m *testing.M) {
	RegisterExtras()
	os.Exit(m.Run())
}

// A cacheWithDummyProposals overrides `OpenTrie()` to returns a
// [trieWithDummyProposals].
type cacheWithDummyProposals struct {
	state.Database
}

func (db *cacheWithDummyProposals) OpenTrie(root common.Hash) (state.Trie, error) {
	t, err := db.Database.OpenTrie(root)
	if err != nil {
		return nil, err
	}
	return &trieWithDummyProposals{Trie: t}, nil
}

// A trieWithDummyProposals overrides `Commit()` to inject a [proposal] into the
// returned [trienode.NodeSet].
type trieWithDummyProposals struct {
	state.Trie
}

func (t *trieWithDummyProposals) Commit(collectLeaf bool) (common.Hash, *trienode.NodeSet, error) {
	root, set, err := t.Trie.Commit(collectLeaf)
	if err != nil {
		return common.Hash{}, nil, err
	}
	// This, combined with [proposalPayload.MergeNodeSet], is where the magic
	// happens. We use the existing geth plumbing to carry the proposal back to
	// [hashDBWithDummyProposals.Update], knowing that the Go GC will trigger
	// the FFI call to free the Rust memory.
	p := &proposal{root: root}
	p.setFinalizer()
	p.injectInto(set)

	return root, set, nil
}

// A hashDBWithDummyProposals overrides `Update()` to capture the [proposal]
// propagated from [trieWithDummyProposals.Commit].
type hashDBWithDummyProposals struct {
	*hashdb.Database
	got *proposal
}

func (db *hashDBWithDummyProposals) Reader(root common.Hash) (databasepkg.Reader, error) {
	return db.Database.Reader(root)
}

func (db *hashDBWithDummyProposals) Update(root, parent common.Hash, block uint64, nodes *trienode.MergedNodeSet, states *triestate.Set, opts ...stateconf.TrieDBUpdateOption) error {
	db.got = extras.MergedNodeSet.Get(nodes)
	return db.Database.Update(root, parent, block, nodes, states, opts...)
}

func TestProposalPropagation(t *testing.T) {
	db := rawdb.NewMemoryDatabase()
	backend := &hashDBWithDummyProposals{
		Database: hashdb.New(db, nil, trie.MerkleResolver{}),
	}
	tdb := triedb.NewDatabase(db, &triedb.Config{
		DBOverride: func(db ethdb.Database) triedb.DBOverride {
			return backend
		},
	})

	cache := &cacheWithDummyProposals{
		Database: state.NewDatabaseWithNodeDB(db, tdb),
	}
	sdb, err := state.New(types.EmptyRootHash, cache, nil)
	require.NoError(t, err, "state.New([empty root], ...)")

	sdb.SetState(common.Address{}, common.Hash{}, common.Hash{42})
	root, err := sdb.Commit(1, false)
	require.NoErrorf(t, err, "%T.Commit()", sdb)

	if got, want := backend.got.root, root; got != want {
		t.Errorf("got %v; want %v", got, want)
	}

	t.Run("GC_finalizer_invoked", func(t *testing.T) {
		finalized := backend.got.finalized

		// Everything that might still hold a reference to the `proposal`,
		// stopping it from being garbage collected.
		sdb = nil
		cache = nil
		tdb = nil
		backend = nil

		// Note that [runtime.GC] doesn't block on finalizers; see
		// https://go.dev/doc/gc-guide#Testing_object_death
		runtime.GC()
		select {
		case <-finalized:
		case <-time.After(5 * time.Second):
			t.Errorf("%T finalizer did not run", &proposal{})
		}
	})
}
