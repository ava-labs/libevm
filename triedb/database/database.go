// Copyright 2024 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package database

import (
	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/libevm/trie/trienode"
	"github.com/ava-labs/libevm/trie/triestate"
)

// Reader wraps the Node method of a backing trie reader.
type Reader interface {
	// Node retrieves the trie node blob with the provided trie identifier,
	// node path and the corresponding node hash. No error will be returned
	// if the node is not found.
	Node(owner common.Hash, path []byte, hash common.Hash) ([]byte, error)
}

// PreimageStore wraps the methods of a backing store for reading and writing
// trie node preimages.
type PreimageStore interface {
	// Preimage retrieves the preimage of the specified hash.
	Preimage(hash common.Hash) []byte

	// InsertPreimage commits a set of preimages along with their hashes.
	InsertPreimage(preimages map[common.Hash][]byte)
}

// Database wraps the methods of a backing trie store.
type Database interface {
	PreimageStore

	// Reader returns a node reader associated with the specific state.
	// An error will be returned if the specified state is not available.
	Reader(stateRoot common.Hash) (Reader, error)
}

// Backend defines the methods needed to access/update trie nodes in different
// state scheme.
type Backend interface {
	// Scheme returns the identifier of used storage scheme.
	Scheme() string

	// Initialized returns an indicator if the state data is already initialized
	// according to the state scheme.
	Initialized(genesisRoot common.Hash) bool

	// Size returns the current storage size of the diff layers on top of the
	// disk layer and the storage size of the nodes cached in the disk layer.
	//
	// For hash scheme, there is no differentiation between diff layer nodes
	// and dirty disk layer nodes, so both are merged into the second return.
	Size() (common.StorageSize, common.StorageSize)

	// Update performs a state transition by committing dirty nodes contained
	// in the given set in order to update state from the specified parent to
	// the specified root.
	//
	// The passed in maps(nodes, states) will be retained to avoid copying
	// everything. Therefore, these maps must not be changed afterwards.
	Update(root common.Hash, parent common.Hash, block uint64, nodes *trienode.MergedNodeSet, states *triestate.Set) error

	// Commit writes all relevant trie nodes belonging to the specified state
	// to disk. Report specifies whether logs will be displayed in info level.
	Commit(root common.Hash, report bool) error

	// Close closes the trie database backend and releases all held resources.
	Close() error

	// Reader returns a node reader associated with the specific state.
	// An error will be returned if the specified state is not available.
	Reader(stateRoot common.Hash) (Reader, error)
}

type HashBackend interface {
	Backend

	Cap(limit common.StorageSize) error
	Reference(root common.Hash, parent common.Hash)
	Dereference(root common.Hash)
}

type PathBackend interface {
	Backend

	Recover(root common.Hash, loader triestate.TrieLoader) error
	Recoverable(root common.Hash) bool
	Disable() error
	Enable(root common.Hash) error
	Journal(root common.Hash) error
	SetBufferSize(size int) error
}
