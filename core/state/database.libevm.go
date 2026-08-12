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

package state

import (
	"github.com/ava-labs/libevm/ethdb"
	"github.com/ava-labs/libevm/libevm"
	"github.com/ava-labs/libevm/libevm/register"
	"github.com/ava-labs/libevm/triedb"
)

// A DatabaseInterceptor accepts the standard [Database] implementation
// returned by [NewDatabaseWithConfig] and [NewDatabaseWithNodeDB], allowing
// the returned [Database] to be re-implemented with custom behavior.
type DatabaseInterceptor func(Database) Database

// RegisterDatabaseInterceptor registers the [DatabaseInterceptor] such that it
// modifies the behaviour of all [StateDB] instances. It MUST NOT be called
// more than once.
func RegisterDatabaseInterceptor(dbi DatabaseInterceptor) {
	registeredInterceptor.MustRegister(dbi)
}

// WithTempRegisteredDatabaseInterceptor temporarily registers `dbi` as if
// calling [RegisterDatabaseInterceptor] with the same argument. After `fn`
// returns, the registration is returned to its former state, be that none or
// the types originally passed to [RegisterDatabaseInterceptor].
//
// This MUST NOT be used on a live chain. It is solely intended for off-chain
// consumers that require access to extras. Said consumers SHOULD NOT, however
// call this function directly. Use the [libevm.WithTemporaryExtrasLock]
// function instead in combination with all other registrations to ensure
// that temporary registrations are atomically applied.
func WithTempRegisteredDatabaseInterceptor(lock libevm.ExtrasLock, dbi DatabaseInterceptor, fn func() error) error {
	if err := lock.Verify(); err != nil {
		return err
	}
	return registeredInterceptor.TempOverride(dbi, fn)
}

// TestOnlyClearRegisteredDatabaseInterceptor clears the argument previously passed to
// [RegisterDatabaseInterceptor]. It panics if called from a non-testing call stack.
//
// In tests it SHOULD be called before every call to [RegisterDatabaseInterceptor] and then
// defer-called afterwards, either directly or via testing.TB.Cleanup(). This is
// a workaround for the single-call limitation on [RegisterDatabaseInterceptor].
func TestOnlyClearRegisteredDatabaseInterceptor() {
	registeredInterceptor.TestOnlyClear()
}

var registeredInterceptor register.AtMostOnce[DatabaseInterceptor]

// NewDatabaseWithConfig creates a backing store for state. The returned database
// is safe for concurrent use and retains a lot of collapsed RLP trie nodes in a
// large memory cache. If a [DatabaseInterceptor] is registered, the returned
// database will be the result of passing the [Database] returned by this
// function to the interceptor.
func NewDatabaseWithConfig(db ethdb.Database, config *triedb.Config) Database {
	return interceptDB(newDatabaseWithConfig(db, config))
}

// NewDatabaseWithNodeDB creates a state database with an already initialized node database.
// If a [DatabaseInterceptor] is registered, the returned database will be the result of
// passing the [Database] returned by this function to the interceptor.
func NewDatabaseWithNodeDB(db ethdb.Database, triedb *triedb.Database) Database {
	return interceptDB(newDatabaseWithNodeDB(db, triedb))
}

func interceptDB(db Database) Database {
	if r := &registeredInterceptor; r.Registered() {
		return r.Get()(db)
	}
	return db
}
