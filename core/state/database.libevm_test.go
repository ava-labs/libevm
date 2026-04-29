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
	"testing"

	"github.com/ava-labs/libevm/libevm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type dbWrapper struct {
	Database
}

func dbInterceptor(db Database) Database {
	return &dbWrapper{Database: db}
}

func TestDatabaseRegistration(t *testing.T) {
	assertDBWrapped(t, false)

	RegisterDatabaseInterceptor(dbInterceptor)
	t.Cleanup(TestOnlyClearRegisteredDatabaseInterceptor)

	assertDBWrapped(t, true)
}
func TestTempDatabaseRegistration(t *testing.T) {
	err := libevm.WithTemporaryExtrasLock(func(lock libevm.ExtrasLock) error {
		return WithTempRegisteredDatabaseInterceptor(lock, dbInterceptor, func() error {
			assertDBWrapped(t, true)
			return nil
		})
	})
	require.NoError(t, err, "WithTempRegisteredDatabaseInterceptor")

	assertDBWrapped(t, false)
}

func assertDBWrapped(t *testing.T, wantWrapped bool) {
	t.Helper()

	var wantType any
	if wantWrapped {
		wantType = &dbWrapper{}
	} else {
		wantType = &cachingDB{}
	}
	assert.IsType(t, wantType, NewDatabase(nil), "NewDatabase(nil)")
	assert.IsType(t, wantType, NewDatabaseWithConfig(nil, nil), "NewDatabaseWithConfig(nil, nil)")
	assert.IsType(t, wantType, NewDatabaseWithNodeDB(nil, nil), "NewDatabaseWithNodeDB(nil, nil)")
}
