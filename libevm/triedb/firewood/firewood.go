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
	"slices"
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

var (
	_ triedb.DBConstructor = Config{}.BackendConstructor
	_ triedb.DBOverride    = (*Database)(nil)
)

// Database is a triedb.DBOverride implementation backed by Firewood.
// It acts as a HashDB for backwards compatibility with most of the blockchain code.
type Database struct {
	// The underlying Firewood database, used for storing proposals and revisions.
	// This is exported as read-only, with knowledge that the consumer will not close it
	// and the latest state can be modified at any time.
	Firewood *ffi.Database

	proposals
}

type proposals struct {
	sync.RWMutex

	byStateRoot map[common.Hash][]*proposal
	// The proposal tree tracks the structure of the current proposals, and which proposals are children of which.
	// This is used to ensure that we can dereference proposals correctly and commit the correct ones
	// in the case of duplicate state roots.
	// The root of the tree is stored here, and represents the top-most layer on disk.
	tree *proposal
	// possible temporarily holds proposals created during a trie update.
	// This is cleared after the update is complete and the proposals have been sent to the database.
	possible []*proposal
}

// A proposal carries a Firewood FFI proposal (i.e. Rust-owned memory).
// The Firewood library adds a finalizer to the proposal handle to ensure that
// the memory is freed when the Go object is garbage collected. However, because
// we form a tree of proposals, the `proposal.Proposal` field may be the only
// reference to a given proposal. To ensure that all proposals in the tree
// can be freed in a finalizer, this cannot be included in the tree structure.
type proposal struct {
	*proposalMeta
	handle *ffi.Proposal
}

type proposalMeta struct {
	parent      *proposalMeta
	children    []*proposalMeta
	blockHashes map[common.Hash]struct{} // All corresponding block hashes
	root        common.Hash
	height      uint64
}

// Config provides necessary parameters for creating a Firewood database.
type Config struct {
	DatabasePath         string
	CacheSizeBytes       uint
	FreeListCacheEntries uint
	MaxRevisions         uint
	CacheStrategy        ffi.CacheStrategy
}

// DefaultConfig returns a default Config with the given directory.
// The default config is:
//   - CacheSizeBytes: 1MB
//   - FreeListCacheEntries: 40,000
//   - MaxRevisions: 100
//   - CacheStrategy: [ffi.CacheAllReads]
func DefaultConfig(dir string) Config {
	return Config{
		DatabasePath:         dir,
		CacheSizeBytes:       1024 * 1024, // 1MB
		FreeListCacheEntries: 40_000,
		MaxRevisions:         100,
		CacheStrategy:        ffi.CacheAllReads,
	}
}

// BackendConstructor implements the [triedb.DBConstructor] interface.
// It creates a new Firewood database with the given configuration.
// Since Firewood uses its own on-disk format, the provided ethdb.Database is ignored.
// Any error during creation will cause the program to exit.
func (c Config) BackendConstructor(ethdb.Database) triedb.DBOverride {
	db, err := New(c)
	if err != nil {
		log.Crit("firewood: error creating database", "error", err)
	}
	return db
}

// New creates a new Firewood database with the given configuration.
// The database will not be opened on error.
func New(config Config) (*Database, error) {
	if err := validateDir(config.DatabasePath); err != nil {
		return nil, err
	}

	logPath := filepath.Join(config.DatabasePath, logFileName)
	if err := ffi.StartLogs(&ffi.LogConfig{Path: logPath}); err != nil {
		// This shouldn't be a fatal error, as this can only be called once per thread.
		// Specifically, this will return an error in unit tests.
		log.Warn("firewood: error starting logs", "error", err)
	}

	dbPath := filepath.Join(config.DatabasePath, dbFileName)
	fw, err := ffi.New(dbPath, &ffi.Config{
		NodeCacheEntries:     config.CacheSizeBytes / 256, // TODO(alarso16): 256 bytes per node may not be accurate
		FreeListCacheEntries: config.FreeListCacheEntries,
		Revisions:            config.MaxRevisions,
		ReadCacheStrategy:    config.CacheStrategy,
	})
	if err != nil {
		return nil, err
	}

	intialRoot, err := fw.Root()
	if err != nil {
		if closeErr := fw.Close(context.Background()); closeErr != nil {
			return nil, fmt.Errorf("%w: error while closing: %w", err, closeErr)
		}
		return nil, err
	}

	return &Database{
		Firewood: fw,
		proposals: proposals{
			byStateRoot: make(map[common.Hash][]*proposal),
			tree: &proposal{
				proposalMeta: &proposalMeta{
					root: common.Hash(intialRoot),
					blockHashes: map[common.Hash]struct{}{
						{}: {}, // genesis block has empty hash
					},
				},
			},
		},
	}, nil
}

func validateDir(dir string) error {
	if dir == "" {
		return errors.New("chain data directory must be set")
	}

	switch info, err := os.Stat(dir); {
	case os.IsNotExist(err):
		log.Info("Database directory not found, creating", "path", dir)
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return fmt.Errorf("creating database directory: %v", err)
		}
		return nil
	case err != nil:
		return fmt.Errorf("os.Stat() on database directory: %v", err)
	case !info.IsDir():
		return fmt.Errorf("database directory path is not a directory: %q", dir)
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
	root, err := db.Firewood.Root()
	if err != nil {
		log.Error("firewood: error getting current root", "error", err)
		return false
	}

	// If the current root isn't empty, then unless the genesis block is empty,
	// the database is initialized.
	return common.Hash(root) != types.EmptyRootHash
}

// Size returns the storage size of diff layer nodes above the persistent disk
// layer and the dirty nodes buffered within the disk layer
// Only used for metrics and Commit intervals in APIs.
// This will be implemented in the firewood database eventually.
// Currently, Firewood stores all revisions in disk and proposals in memory.
func (*Database) Size() (common.StorageSize, common.StorageSize) {
	return 0, 0
}

// Reference is no-op because proposals are only referenced when created.
// Additionally, internal nodes do not need tracked by consumers.
func (*Database) Reference(common.Hash, common.Hash) {}

// Dereference is no-op because unused proposals are dereferenced when no longer valid.
// We cannot dereference at this call. Consider the following case:
// Chain 1 has root A and root C
// Chain 2 has root B and root C
// We commit root A, and immediately dereference root B and its child.
// Root C is Rejected, (which is intended to be 2C) but there's now only one record of root C in the proposal map.
// Thus, we recognize the single root C as the only proposal, and dereference it.
func (*Database) Dereference(common.Hash) {}

// Cap is a no-op because it isn't supported by Firewood.
func (*Database) Cap(common.StorageSize) error {
	return nil
}

// Close closes the database, freeing all associated resources.
// This may hang for a short period while waiting for finalizers to complete.
// If it does not close as expected, this indicates that there are still references
// to proposals or revisions in memory, and an error will be returned.
// The database should not be used after calling Close.
func (db *Database) Close() error {
	p := &db.proposals
	p.Lock()
	defer p.Unlock()

	if p.tree == nil {
		return nil // already closed
	}

	// All remaining proposals can explicitly be dropped.
	p.cleanupPossible(nil)
	for _, child := range p.tree.children {
		p.removeProposalAndChildren(child)
	}
	p.byStateRoot = nil
	p.tree = nil

	// Close the database
	// We must provide a context since it may hang while waiting for the finalizers to complete.
	return db.Firewood.Close(context.Background())
}

// Update updates the database to the given root at the given height.
// The parent block hash and block hash must be provided in the options.
// A proposal must have already been created from [AccountTrie.Commit] with the same root,
// parent root, and height.
// If no such proposal exists, an error will be returned.
func (db *Database) Update(root, parent common.Hash, height uint64, _ *trienode.MergedNodeSet, _ *triestate.Set, opts ...stateconf.TrieDBUpdateOption) error {
	// We require block hashes to be provided for all blocks in production.
	// However, many tests cannot reasonably provide a block blockHash for genesis, so we allow it to be omitted.
	parentBlockHash, blockHash, ok := stateconf.ExtractTrieDBUpdatePayload(opts...)
	if !ok {
		log.Error("firewood: no block hash provided for block %d", height)
	}

	// The rest of the operations except key-value arranging must occur with a lock
	db.proposals.Lock()
	defer db.proposals.Unlock()

	if db.proposals.exists(root, blockHash, parentBlockHash) {
		return nil
	}

	p, ok := db.proposals.findUnverified(height, parentBlockHash)
	defer db.proposals.cleanupPossible(p)
	if !ok {
		return fmt.Errorf("firewood: no unverified proposal found for block %d, root %s, hash %s", height, root.Hex(), blockHash.Hex())
	}

	switch {
	case p.root != root:
		return fmt.Errorf("firewood: proposal root mismatch, expected %#x, got %#x", root, p.root)
	case p.parent.root != parent:
		return fmt.Errorf("firewood: proposal parent root mismatch, expected %#x, got %#x", parent, p.parent.root)
	case p.height != height:
		return fmt.Errorf("firewood: proposal block mismatch, expected %d, got %d", height, p.height)
	}

	// Track the proposal context in the tree and map.
	p.parent.children = append(p.parent.children, p.proposalMeta)
	db.proposals.byStateRoot[root] = append(db.proposals.byStateRoot[root], p)
	p.blockHashes[blockHash] = struct{}{}

	// Now, all unused proposals have no other references, since we didn't store them
	// in the proposal map or tree, so they will be garbage collected.
	db.proposals.possible = nil
	return nil
}

func (ps *proposals) exists(root, block, parentBlock common.Hash) bool {
	// Check if this proposal already exists.
	// During reorgs, we may have already created this proposal.
	// Additionally, we may have already created this proposal with a different block hash.
	proposals, ok := ps.byStateRoot[root]
	if !ok {
		return false
	}

	for _, p := range proposals {
		// If the block hash is already tracked, we can skip proposing this again.
		if _, ok := p.blockHashes[block]; ok {
			log.Debug("firewood: proposal already exists", "root", root.Hex(), "parent", parentBlock.Hex(), "block", block, "hash", block.Hex())
			return true
		}

		// We already have this proposal, but should create a new context with the correct hash.
		// This solves the case of a unique block hash, but the same underlying proposal.
		if _, ok := p.parent.blockHashes[parentBlock]; ok {
			log.Debug("firewood: proposal already exists, updating hash", "root", root.Hex(), "parent", parentBlock.Hex(), "block", block, "hash", block.Hex())
			p.blockHashes[block] = struct{}{}
			return true
		}
	}

	return false
}

func (ps *proposals) findUnverified(height uint64, parentBlock common.Hash) (*proposal, bool) {
	// First, try to find a proposal with the correct parent block hash.
	p, ok := find(ps.possible, func(p *proposal) bool {
		_, ok := p.parent.blockHashes[parentBlock]
		return ok
	})
	// The exact proposal was found.
	if ok {
		return p, true
	}

	// Otherwise, this may be the first block after startup, and the parent hash wasn't recorded.
	p, ok = find(ps.possible, func(p *proposal) bool {
		_, ok := p.parent.blockHashes[common.Hash{}]
		return ok
	})
	if !ok {
		// No suitable proposal found.
		return nil, false
	}

	// For clarity, while using this proposal, we should update the meta.
	p.proposalMeta.height = height
	p.proposalMeta.parent.blockHashes[parentBlock] = struct{}{}
	return p, true
}

// cleanupPossible drops all possible proposals except the given one.
// If nil, all possible proposals are dropped.
func (ps *proposals) cleanupPossible(p *proposal) {
	for _, candidate := range ps.possible {
		if candidate == p {
			continue
		}
		if err := candidate.handle.Drop(); err != nil {
			log.Error("error dropping proposal", "root", candidate.root, "height", candidate.height, "err", err)
		}
	}
	ps.possible = nil
}

// find is equivalent to [slices.IndexFunc], returning the element instead of
// the index. The returned boolean indicates whether a suitable element was
// found.
func find[S ~[]E, E comparable](s S, fn func(E) bool) (E, bool) {
	idx := slices.IndexFunc(s, fn)
	if idx == -1 {
		var zero E
		return zero, false
	}
	return s[idx], true
}

// Commit persists a proposal as a revision to the database.
//
// Any time this is called, we expect either:
//  1. The root is the same as the current root of the database (empty block during bootstrapping)
//  2. We have created a valid propsal with that root, and it is of height +1 above the proposal tree root.
//     Additionally, this will be unique.
//
// Afterward, we know that no other proposal at this height can be committed, so we can dereference all
// children in the the other branches of the proposal tree.
func (db *Database) Commit(root common.Hash, report bool) error {
	db.proposals.Lock()
	defer db.proposals.Unlock()

	p, err := db.proposals.findProposalToCommitWhenLocked(root)
	if err != nil {
		return err
	}

	if err := p.handle.Commit(); err != nil {
		return fmt.Errorf("firewood: error committing proposal %s: %w", root.Hex(), err)
	}
	p.handle = nil // The proposal has been committed.

	newRoot, err := db.Firewood.Root()
	if err != nil {
		return fmt.Errorf("firewood: error getting current root after commit: %w", err)
	}
	if common.Hash(newRoot) != root {
		return fmt.Errorf("firewood: root after commit (%#x) does not match expected root %#x", newRoot, root)
	}

	var logFn = log.Debug
	if report {
		logFn = log.Info
	}
	logFn("Persisted proposal to firewood database", "root", root)

	// On success, we should dereference all children of the committed proposal.
	// By removing all uncommittable proposals from the tree and map,
	// we ensure that there are no more references.
	db.cleanupCommittedProposal(p)
	return nil
}

func (ps *proposals) findProposalToCommitWhenLocked(root common.Hash) (*proposal, error) {
	var candidate *proposal

	for _, p := range ps.byStateRoot[root] {
		if p.parent.root != ps.tree.root || p.parent.height != ps.tree.height {
			continue
		}
		if candidate != nil {
			// This should never happen, as we ensure that we don't create duplicate proposals in `propose`.
			return nil, fmt.Errorf("firewood: multiple proposals found for root %#x", root)
		}
		candidate = p
	}

	if candidate == nil {
		return nil, fmt.Errorf("firewood: committable proposal not found for %d:%#x", ps.tree.height+1, root)
	}
	return candidate, nil
}

// createProposal creates a new proposal from the given layer
// If there are no changes, it will return nil.
func (db *Database) createProposal(parent *proposal, keys, values [][]byte) (*proposal, error) {
	propose := db.Firewood.Propose
	if h := parent.handle; h != nil {
		propose = h.Propose
	}
	handle, err := propose(keys, values)
	if err != nil {
		return nil, fmt.Errorf("firewood: unable to create proposal from parent root %s: %w", parent.root.Hex(), err)
	}

	// Edge case: genesis block
	block := parent.height + 1
	if _, ok := parent.blockHashes[common.Hash{}]; ok && parent.root == types.EmptyRootHash {
		block = 0
	}

	p := &proposal{
		handle: handle,
		proposalMeta: &proposalMeta{
			blockHashes: make(map[common.Hash]struct{}),
			parent:      parent.proposalMeta,
			height:      block,
		},
	}

	root, err := handle.Root()
	if err != nil {
		return nil, fmt.Errorf("firewood: error getting root of proposals: %w", err)
	}
	p.root = common.Hash(root)

	return p, nil
}

// cleanupCommittedProposal dereferences the proposal and removes it from the proposal map.
// It also recursively dereferences all children of the proposal.
func (ps *proposals) cleanupCommittedProposal(p *proposal) {
	oldChildren := ps.tree.children
	ps.tree = p
	ps.tree.parent = nil
	ps.tree.handle = nil

	ps.removeProposalFromMap(p.proposalMeta, false)

	for _, child := range oldChildren {
		if child != p.proposalMeta {
			ps.removeProposalAndChildren(child)
		}
	}
}

// Internally removes all references of the proposal from the database.
// Should only be accessed with the proposal lock held.
// Consumer must not be iterating the proposal map at this root.
func (ps *proposals) removeProposalAndChildren(p *proposalMeta) {
	// Base case: if there are children, we need to dereference them as well.
	for _, child := range p.children {
		ps.removeProposalAndChildren(child)
	}

	// Remove the proposal from the map.
	ps.removeProposalFromMap(p, true)
}

// removeProposalFromMap removes the proposal from the proposal map.
// The proposal lock must be held when calling this function.
func (ps *proposals) removeProposalFromMap(meta *proposalMeta, drop bool) {
	rootList := ps.byStateRoot[meta.root]
	for i, p := range rootList {
		if p.proposalMeta == meta { // pointer comparison - guaranteed to be unique
			rootList[i] = rootList[len(rootList)-1]
			rootList[len(rootList)-1] = nil
			rootList = rootList[:len(rootList)-1]

			if drop {
				if err := p.handle.Drop(); err != nil {
					log.Error("error dropping proposal", "root", meta.root, "height", meta.height, "err", err)
				}
			}
			break
		}
	}
	if len(rootList) == 0 {
		delete(ps.byStateRoot, meta.root)
	} else {
		ps.byStateRoot[meta.root] = rootList
	}
}

// getProposalHash calculates the hash if the set of keys and values are
// proposed from the given parent root.
func (db *Database) getProposalHash(parentRoot common.Hash, keys, values [][]byte) (common.Hash, error) {
	if len(keys) != len(values) {
		return common.Hash{}, fmt.Errorf("firewood: keys and values must have the same length, got %d keys and %d values", len(keys), len(values))
	}

	// This function only reads from existing tracked proposals, so we can use a read lock.
	db.proposals.RLock()
	defer db.proposals.RUnlock()

	var handles []*proposal
	if db.proposals.tree.root == parentRoot {
		// Propose from the database root.
		p, err := db.createProposal(db.proposals.tree, keys, values)
		if err != nil {
			return common.Hash{}, fmt.Errorf("firewood: error proposing from root %s: %w", parentRoot.Hex(), err)
		}
		handles = append(handles, p)
	}

	// Find any proposal with the given parent root.
	// Since we are only using the proposal to find the root hash,
	// we can use the first proposal found.
	for _, parent := range db.proposals.byStateRoot[parentRoot] {
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
	db.proposals.possible = handles

	// Get the root of the first proposal - they should all match.
	root := handles[0].root
	return root, nil
}

// Reader retrieves a node reader belonging to the given state root.
// An error will be returned if the requested state is not available.
func (db *Database) Reader(root common.Hash) (database.Reader, error) {
	if _, err := db.Firewood.GetFromRoot(ffi.Hash(root), []byte{}); err != nil {
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
	return reader.db.Firewood.GetFromRoot(ffi.Hash(reader.root), path)
}
