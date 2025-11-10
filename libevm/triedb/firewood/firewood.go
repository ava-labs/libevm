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

// The firewood package provides a [triedb.DBOverride] backed by [Firewood].
//
// [Firewood]: https://github.com/ava-labs/firewood
package firewood

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/ava-labs/firewood-go-ethhash/ffi"

	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/libevm/core/rawdb"
	"github.com/ava-labs/libevm/core/types"
	"github.com/ava-labs/libevm/ethdb"
	"github.com/ava-labs/libevm/libevm/stateconf"
	"github.com/ava-labs/libevm/log"
	"github.com/ava-labs/libevm/trie/trienode"
	"github.com/ava-labs/libevm/trie/triestate"
	"github.com/ava-labs/libevm/triedb"
	"github.com/ava-labs/libevm/triedb/database"
)

const (
	dbFileName  = "firewood.db"
	logFileName = "firewood.log"
)

var _ triedb.DBOverride = (*Database)(nil)

type Database struct {
	// The underlying Firewood database, used for storing proposals and revisions.
	// This must be exported so other packages (e.g. state sync) can access firewood-specific methods.
	Firewood *ffi.Database

	// proposalLock protects the proposal map and tree during updates.
	proposalLock sync.RWMutex
	// proposalMap provides O(1) access by state root to all proposals stored in the proposalTree
	proposalMap map[common.Hash][]*proposal
	// The proposal tree tracks the structure of the current proposals, and which proposals are children of which.
	// This is used to ensure that we can dereference proposals correctly and commit the correct ones
	// in the case of duplicate state roots.
	// The root of the tree is stored here, and represents the top-most layer on disk.
	proposalTree *proposal

	// possibleProposals temporarily holds proposals created during a trie update.
	// This is cleared after the update is complete and the proposals have been sent to the database.
	possibleProposals []*proposal
}

type Config struct {
	ChainDir             string
	CleanCacheSize       int  // Size of the clean cache in bytes
	FreeListCacheEntries uint // Number of free list entries to cache
	Revisions            uint
	ReadCacheStrategy    ffi.CacheStrategy
}

// Note that `FilePath` is not specified, and must always be set by the user.
var Defaults = Config{
	CleanCacheSize:       1024 * 1024, // 1MB
	FreeListCacheEntries: 40_000,
	Revisions:            100,
	ReadCacheStrategy:    ffi.CacheAllReads,
}

func (c Config) BackendConstructor(ethdb.Database) triedb.DBOverride {
	db, err := New(c)
	if err != nil {
		log.Crit("firewood: error creating database", "error", err)
	}
	return db
}

// New creates a new Firewood database with the given disk database and configuration.
// Any error during creation will cause the program to exit.
func New(config Config) (*Database, error) {
	if err := validatePath(config.ChainDir); err != nil {
		return nil, err
	}

	// Start the logs prior to opening the database
	logPath := filepath.Join(config.ChainDir, logFileName)
	if err := ffi.StartLogs(&ffi.LogConfig{Path: logPath}); err != nil {
		// This shouldn't be a fatal error, as this can only be called once per thread.
		// Specifically, this will return an error in unit tests.
		log.Warn("firewood: error starting logs", "error", err)
	}

	dbPath := filepath.Join(config.ChainDir, dbFileName)
	fw, err := ffi.New(dbPath, &ffi.Config{
		NodeCacheEntries:     uint(config.CleanCacheSize) / 256, // TODO: estimate 256 bytes per node
		FreeListCacheEntries: config.FreeListCacheEntries,
		Revisions:            config.Revisions,
		ReadCacheStrategy:    config.ReadCacheStrategy,
	})
	if err != nil {
		return nil, err
	}

	currentRoot, err := fw.Root()
	if err != nil {
		return nil, err
	}

	return &Database{
		Firewood:    fw,
		proposalMap: make(map[common.Hash][]*proposal),
		proposalTree: &proposal{
			info: &blockInfo{
				root: common.Hash(currentRoot),
				hashes: map[common.Hash]struct{}{
					{}: {}, // genesis block has empty hash
				},
			},
		},
	}, nil
}

func validatePath(dir string) error {
	if dir == "" {
		return errors.New("chain data directory must be set")
	}

	// Check that the directory exists
	switch info, err := os.Stat(dir); {
	case os.IsNotExist(err):
		log.Info("Database directory not found, creating", "path", dir)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("error creating database directory: %w", err)
		}
		return nil
	case err != nil:
		return fmt.Errorf("error checking database directory: %w", err)
	case !info.IsDir():
		return fmt.Errorf("database directory path is not a directory: %s", dir)
	}

	return nil
}

// Scheme returns the scheme of the database.
// This is only used in some API calls
// and in StateDB to avoid iterating through deleted storage tries.
// WARNING: If anything from go-ethereum uses this,
// it must be overwritten to use something like:
// `_, ok := db.(*Database); if !ok { return "" }`
// to recognize the Firewood database.
func (*Database) Scheme() string {
	return rawdb.HashScheme
}

// Initialized checks whether a non-empty genesis block has been written.
func (db *Database) Initialized(common.Hash) bool {
	rootBytes, err := db.Firewood.Root()
	if err != nil {
		log.Error("firewood: error getting current root", "error", err)
		return false
	}
	root := common.BytesToHash(rootBytes)
	// If the current root isn't empty, then unless the database is empty, we have a genesis block recorded.
	return root != types.EmptyRootHash
}

// Size returns the storage size of diff layer nodes above the persistent disk
// layer and the dirty nodes buffered within the disk layer
// Only used for metrics and Commit intervals in APIs.
// This will be implemented in the firewood database eventually.
// Currently, Firewood stores all revisions in disk and proposals in memory.
func (*Database) Size() (common.StorageSize, common.StorageSize) {
	return 0, 0
}

// This isn't called anywhere in coreth
func (*Database) Reference(common.Hash, common.Hash) {}

// Dereference drops a proposal from the database.
// This function is no-op because unused proposals are dereferenced when no longer valid.
// We cannot dereference at this call. Consider the following case:
// Chain 1 has root A and root C
// Chain 2 has root B and root C
// We commit root A, and immediately dereference root B and its child.
// Root C is Rejected, (which is intended to be 2C) but there's now only one record of root C in the proposal map.
// Thus, we recognize the single root C as the only proposal, and dereference it.
func (*Database) Dereference(common.Hash) {}

// Firewood does not support this.
func (*Database) Cap(common.StorageSize) error {
	return nil
}

func (db *Database) Close() error {
	db.proposalLock.Lock()
	defer db.proposalLock.Unlock()

	// All remaining proposals must be dereferenced.
	db.possibleProposals = nil
	db.proposalMap = nil
	db.proposalTree = nil

	// Close the database
	// We must provide a context since it may hang while waiting for the finalizers to complete.
	return db.Firewood.Close(context.Background())
}

func (db *Database) Update(root, parent common.Hash, block uint64, _ *trienode.MergedNodeSet, _ *triestate.Set, opts ...stateconf.TrieDBUpdateOption) error {
	// We require block hashes to be provided for all blocks in production.
	// However, many tests cannot reasonably provide a block hash for genesis, so we allow it to be omitted.
	parentHash, hash, ok := stateconf.ExtractTrieDBUpdatePayload(opts...)
	if !ok {
		log.Error("firewood: no block hash provided for block %d", block)
	}

	// The rest of the operations except key-value arranging must occur with a lock
	db.proposalLock.Lock()
	defer db.proposalLock.Unlock()

	// Check if this proposal already exists.
	// During reorgs, we may have already created this proposal.
	// Additionally, we may have already created this proposal with a different block hash.
	if existingProposals, ok := db.proposalMap[root]; ok {
		for _, existing := range existingProposals {
			// If the block hash is already tracked, we can skip proposing this again.
			if _, exists := existing.info.hashes[hash]; exists {
				log.Debug("firewood: proposal already exists", "root", root.Hex(), "parent", parent.Hex(), "block", block, "hash", hash.Hex())
				return nil
			}
			// We already have this proposal, but should create a new context with the correct hash.
			// This solves the case of a unique block hash, but the same underlying proposal.
			if _, exists := existing.info.parent.hashes[parentHash]; exists {
				log.Debug("firewood: proposal already exists, updating hash", "root", root.Hex(), "parent", parent.Hex(), "block", block, "hash", hash.Hex())
				existing.info.hashes[hash] = struct{}{}
				return nil
			}
		}
	}

	// We must use one of the unverified proposals if it exists.
	var h *proposal
	for _, possible := range db.possibleProposals {
		if _, ok := possible.info.parent.hashes[parentHash]; ok {
			h = possible
			break
		}
	}
	if h == nil {
		// Edge case: first set of proposals before `Commit`, or empty genesis block
		// Neither `Update` nor `Commit` is called for genesis, so we can accept a proposal with parentHash of empty.
		for _, possible := range db.possibleProposals {
			if _, ok := possible.info.parent.hashes[common.Hash{}]; ok {
				h = possible
				h.info.height = block
				h.info.parent.hashes[parentHash] = struct{}{}
				break
			}
		}
		if h == nil {
			return fmt.Errorf("firewood: no unverified proposal found for block %d, root %s, hash %s", block, root.Hex(), hash.Hex())
		}
	}

	// Verify that the proposal context matches what we expect.
	switch {
	case h.info.root != root:
		return fmt.Errorf("firewood: proposal root mismatch, expected %s, got %s", root.Hex(), h.info.root.Hex())
	case h.info.parent.root != parent:
		return fmt.Errorf("firewood: proposal parent root mismatch, expected %s, got %s", parent.Hex(), h.info.parent.root.Hex())
	case h.info.height != block:
		return fmt.Errorf("firewood: proposal block mismatch, expected %d, got %d", block, h.info.height)
	}

	// Track the proposal context in the tree and map.
	h.info.parent.children = append(h.info.parent.children, h.info)
	db.proposalMap[root] = append(db.proposalMap[root], h)
	h.info.hashes[hash] = struct{}{}

	// Now, all unused proposals have no other references, since we didn't store them
	// in the proposal map or tree, so they will be garbage collected.
	db.possibleProposals = nil
	return nil
}

// Commit persists a proposal as a revision to the database.
//
// Any time this is called, we expect either:
//  1. The root is the same as the current root of the database (empty block during bootstrapping)
//  2. We have created a valid propsal with that root, and it is of height +1 above the proposal tree root.
//     Additionally, this should be unique.
//
// Afterward, we know that no other proposal at this height can be committed, so we can dereference all
// children in the the other branches of the proposal tree.
func (db *Database) Commit(root common.Hash, report bool) error {
	// We need to lock the proposal tree to prevent concurrent writes.
	var p *proposal
	db.proposalLock.Lock()
	defer db.proposalLock.Unlock()

	// Find the proposal with the given root.
	for _, possible := range db.proposalMap[root] {
		if possible.info.parent.root == db.proposalTree.info.root && possible.info.parent.height == db.proposalTree.info.height {
			// We found the proposal with the correct parent.
			if p != nil {
				// This should never happen, as we ensure that we don't create duplicate proposals in `propose`.
				return fmt.Errorf("firewood: multiple proposals found for %s", root.Hex())
			}
			p = possible
		}
	}
	if p == nil {
		return fmt.Errorf("firewood: committable proposal not found for %d:%s", db.proposalTree.info.height+1, root.Hex())
	}

	// Commit the proposal to the database.
	if err := p.proposal.Commit(); err != nil {
		return fmt.Errorf("firewood: error committing proposal %s: %w", root.Hex(), err)
	}
	p.proposal = nil // The proposal has been committed.

	// fmt.Printf("commit proposal %p\n", p)
	// Assert that the root of the database matches the committed proposal root.
	currentRootBytes, err := db.Firewood.Root()
	if err != nil {
		return fmt.Errorf("firewood: error getting current root after commit: %w", err)
	}
	currentRoot := common.BytesToHash(currentRootBytes)
	if currentRoot != root {
		return fmt.Errorf("firewood: current root %s does not match expected root %s", currentRoot.Hex(), root.Hex())
	}

	if report {
		log.Info("Persisted proposal to firewood database", "root", root)
	} else {
		log.Debug("Persisted proposal to firewood database", "root", root)
	}

	// On success, we should dereference all children of the committed proposal.
	// By removing all uncommittable proposals from the tree and map,
	// we ensure that there are no more references.
	db.cleanupCommittedProposal(p)
	return nil
}

// createProposal creates a new proposal from the given layer
// If there are no changes, it will return nil.
func (db *Database) createProposal(parent *proposal, keys, values [][]byte) (*proposal, error) {
	var (
		p   *ffi.Proposal
		err error
	)

	if parent.proposal == nil {
		p, err = db.Firewood.Propose(keys, values)
	} else {
		p, err = parent.proposal.Propose(keys, values)
	}
	if err != nil {
		return nil, fmt.Errorf("firewood: unable to create proposal from parent root %s: %w", parent.info.root.Hex(), err)
	}

	// Edge case: genesis block
	block := parent.info.height + 1
	if _, ok := parent.info.hashes[common.Hash{}]; ok && parent.info.root == types.EmptyRootHash {
		block = 0
	}

	h := &proposal{
		proposal: p,
		info: &blockInfo{
			hashes: make(map[common.Hash]struct{}),
			parent: parent.info,
			height: block,
		},
	}

	currentRootBytes, err := p.Root()
	if err != nil {
		return nil, fmt.Errorf("firewood: error getting root of proposals: %w", err)
	}
	h.info.root = common.BytesToHash(currentRootBytes)

	return h, nil
}

// cleanupCommittedProposal dereferences the proposal and removes it from the proposal map.
// It also recursively dereferences all children of the proposal.
func (db *Database) cleanupCommittedProposal(p *proposal) {
	oldChildren := db.proposalTree.info.children
	db.proposalTree = p
	db.proposalTree.info.parent = nil
	db.proposalTree.proposal = nil

	db.removeProposalFromMap(p.info)

	for _, child := range oldChildren {
		if child != p.info {
			db.removeProposalAndChildren(child)
		}
	}
}

// Internally removes all references of the proposal from the database.
// Should only be accessed with the proposal lock held.
// Consumer must not be iterating the proposal map at this root.
func (db *Database) removeProposalAndChildren(p *blockInfo) {
	// Base case: if there are children, we need to dereference them as well.
	for _, child := range p.children {
		db.removeProposalAndChildren(child)
	}

	// Remove the proposal from the map.
	db.removeProposalFromMap(p)
}

// removeProposalFromMap removes the proposal from the proposal map.
// The proposal lock must be held when calling this function.
func (db *Database) removeProposalFromMap(info *blockInfo) {
	rootList := db.proposalMap[info.root]
	for i, p := range rootList {
		if p.info == info { // pointer comparison - guaranteed to be unique
			rootList[i] = rootList[len(rootList)-1]
			rootList[len(rootList)-1] = nil
			rootList = rootList[:len(rootList)-1]
			break
		}
	}
	if len(rootList) == 0 {
		delete(db.proposalMap, info.root)
	} else {
		db.proposalMap[info.root] = rootList
	}
}

// getProposalHash calculates the hash if the set of keys and values are
// proposed from the given parent root.
func (db *Database) getProposalHash(parentRoot common.Hash, keys, values [][]byte) (common.Hash, error) {
	if len(keys) != len(values) {
		return common.Hash{}, fmt.Errorf("firewood: keys and values must have the same length, got %d keys and %d values", len(keys), len(values))
	}

	// This function only reads from existing tracked proposals, so we can use a read lock.
	db.proposalLock.RLock()
	defer db.proposalLock.RUnlock()

	var handles []*proposal
	if db.proposalTree.info.root == parentRoot {
		// Propose from the database root.
		p, err := db.createProposal(db.proposalTree, keys, values)
		if err != nil {
			return common.Hash{}, fmt.Errorf("firewood: error proposing from root %s: %w", parentRoot.Hex(), err)
		}
		handles = append(handles, p)
	}

	// Find any proposal with the given parent root.
	// Since we are only using the proposal to find the root hash,
	// we can use the first proposal found.
	for _, parent := range db.proposalMap[parentRoot] {
		p, err := db.createProposal(parent, keys, values)
		if err != nil {
			return common.Hash{}, fmt.Errorf("firewood: error proposing from parent root %s: %w", parentRoot.Hex(), err)
		}
		handles = append(handles, p)
	}

	if len(handles) == 0 {
		return common.Hash{}, fmt.Errorf("firewood: no proposals found with parent root %s", parentRoot.Hex())
	}

	// Store the proposals for later
	db.possibleProposals = handles

	// Get the root of the first proposal - they should all match.
	root := handles[0].info.root
	return root, nil
}

// Reader retrieves a node reader belonging to the given state root.
// An error will be returned if the requested state is not available.
func (db *Database) Reader(root common.Hash) (database.Reader, error) {
	if _, err := db.Firewood.GetFromRoot(root.Bytes(), []byte{}); err != nil {
		return nil, fmt.Errorf("firewood: unable to retrieve from root %s: %w", root.Hex(), err)
	}
	return &reader{db: db, root: root}, nil
}

// reader is a state reader of Database which implements the Reader interface.
type reader struct {
	db   *Database
	root common.Hash // The root of the state this reader is reading.
}

// Node retrieves the trie node with the given node hash. No error will be
// returned if the node is not found.
func (reader *reader) Node(_ common.Hash, path []byte, _ common.Hash) ([]byte, error) {
	// This function relies on Firewood's internal locking to ensure concurrent reads are safe.
	// This is safe even if a proposal is being committed concurrently.
	return reader.db.Firewood.GetFromRoot(reader.root.Bytes(), path)
}
