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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/libevm/core/rawdb"
	"github.com/ava-labs/libevm/core/state/snapshot"
	"github.com/ava-labs/libevm/core/types"
	"github.com/ava-labs/libevm/ethdb"
	"github.com/ava-labs/libevm/libevm/stateconf"
	"github.com/ava-labs/libevm/trie"
	"github.com/ava-labs/libevm/trie/trienode"
	"github.com/ava-labs/libevm/trie/triestate"
	"github.com/ava-labs/libevm/triedb"
	"github.com/ava-labs/libevm/triedb/database"
	"github.com/ava-labs/libevm/triedb/hashdb"
)

func TestStateDBCommitPropagatesOptions(t *testing.T) {
	memdb := rawdb.NewMemoryDatabase()
	triedb := triedb.NewDatabase(
		memdb,
		&triedb.Config{
			DBOverride: func(_ ethdb.Database) triedb.DBOverride {
				return &triedbRecorder{Database: hashdb.New(memdb, nil, &trie.MerkleResolver{})}
			},
		},
	)
	var rec snapTreeRecorder
	sdb, err := New(types.EmptyRootHash, NewDatabaseWithNodeDB(memdb, triedb), &rec)
	require.NoError(t, err, "New()")

	// Ensures that rec.Update() will be called.
	sdb.SetNonce(common.Address{}, 42)

	const snapshotPayload = "hello world"
	const trieDBPayload = "goodbye world"
	snapshotOpt := stateconf.WithSnapshotUpdatePayload(snapshotPayload)
	triedbOpt := stateconf.WithTrieDBUpdatePayload(trieDBPayload)
	_, err = sdb.Commit(0, false, stateconf.WithSnapshotUpdateOpts(snapshotOpt), stateconf.WithTrieDBUpdateOpts(triedbOpt))
	require.NoErrorf(t, err, "%T.Commit(..., %T, %T)", sdb, snapshotOpt, triedbOpt)

	assert.Equalf(t, snapshotPayload, rec.gotPayload, "%T payload propagated via %T.Commit() to %T.Update()", snapshotOpt, sdb, rec)
	innerTrieDB, ok := triedb.Backend().(*triedbRecorder)
	if !ok {
		t.Fatalf("expected %T to be a *triedbRecorder", triedb.Backend())
	}
	assert.Equalf(t, trieDBPayload, innerTrieDB.gotPayload, "%T payload propagated via %T.Commit() to %T.Update()", triedbOpt, sdb, rec)
}

type snapTreeRecorder struct {
	SnapshotTree
	gotPayload any
}

func (*snapTreeRecorder) Cap(common.Hash, int) error {
	return nil
}

func (r *snapTreeRecorder) Update(
	_, _ common.Hash,
	_ map[common.Hash]struct{}, _ map[common.Hash][]byte, _ map[common.Hash]map[common.Hash][]byte,
	opts ...stateconf.SnapshotUpdateOption,
) error {
	r.gotPayload = stateconf.ExtractSnapshotUpdatePayload(opts...)
	return nil
}

func (*snapTreeRecorder) Snapshot(common.Hash) snapshot.Snapshot {
	return snapshotStub{}
}

type snapshotStub struct {
	snapshot.Snapshot
}

func (snapshotStub) Account(common.Hash) (*types.SlimAccount, error) {
	return &types.SlimAccount{}, nil
}

func (snapshotStub) Root() common.Hash {
	return common.Hash{}
}

type triedbRecorder struct {
	*hashdb.Database
	gotPayload any
}

func (r *triedbRecorder) Update(
	root common.Hash,
	parent common.Hash,
	block uint64,
	nodes *trienode.MergedNodeSet,
	states *triestate.Set,
	opts ...stateconf.TrieDBUpdateOption,
) error {
	r.gotPayload = stateconf.ExtractTrieDBUpdatePayload(opts...)
	return r.Database.Update(root, parent, block, nodes, states)
}

func (r *triedbRecorder) Reader(_ common.Hash) (database.Reader, error) {
	return r.Database.Reader(common.Hash{})
}
