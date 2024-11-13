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

package triedb

import (
	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/libevm/ethdb"
	"github.com/ava-labs/libevm/log"
	"github.com/ava-labs/libevm/trie/triestate"
	"github.com/ava-labs/libevm/triedb/database"
	"github.com/ava-labs/libevm/triedb/hashdb"
	"github.com/ava-labs/libevm/triedb/pathdb"
)

// Backend defines the intersection of methods shared by [hashdb.Database] and
// [pathdb.Database].
type Backend backend

// A ReaderProvider exposes its underlying Reader as an interface. Both
// [hashdb.Database] and [pathdb.Database] return concrete types so Go's lack of
// support for [covariant types] means that this method can't be defined on
// [Backend].
//
// [covariant types]: https://go.dev/doc/faq#covariant_types
type ReaderProvider interface {
	Reader(common.Hash) (database.Reader, error)
}

// A BackendConstructor constructs alternative backend implementations.
type BackendConstructor func(ethdb.Database, *Config) BackendOverride

// A BackendOverride is an arbitrary implementation of a [Database] backend. It
// MUST be either a [HashBackend] or a [PathBackend].
type BackendOverride interface {
	Backend
	ReaderProvider
}

func (db *Database) overrideBackend(diskdb ethdb.Database, config *Config) {
	if config.DBOverride == nil {
		return
	}
	if config.HashDB != nil || config.PathDB != nil {
		log.Crit("Database override provided when 'hash' or 'path' mode are configured")
	}
	db.backend.Close() //nolint:gosec // geth defaults to hashdb instances, which always return nil from Close()

	db.backend = config.DBOverride(diskdb, config)
	switch db.backend.(type) {
	case HashBackend:
	case PathBackend:
	default:
		log.Crit("Database override is neither hash- nor path-based")
	}
}

var (
	// If either of these break then the respective interface SHOULD be updated.
	_ HashBackend = (*hashdb.Database)(nil)
	_ PathBackend = (*pathdb.Database)(nil)
)

// A HashBackend mirrors the functionality of a [hashdb.Database].
type HashBackend interface {
	Backend

	Cap(limit common.StorageSize) error
	Reference(root common.Hash, parent common.Hash)
	Dereference(root common.Hash)
}

// A PathBackend mirrors the functionality of a [pathdb.Database].
type PathBackend interface {
	Backend

	Recover(root common.Hash, loader triestate.TrieLoader) error
	Recoverable(root common.Hash) bool
	Disable() error
	Enable(root common.Hash) error
	Journal(root common.Hash) error
	SetBufferSize(size int) error
}
