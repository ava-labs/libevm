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

package rawdb_test

import (
	"github.com/ava-labs/libevm/common"
	// To ensure that all methods are available to importing packages, this test
	// is defined in package `rawdb_test` instead of `rawdb`.
	"github.com/ava-labs/libevm/core/rawdb"
)

// ExampleDatabaseStat demonstrates the method signatures of DatabaseStat, which
// exposes an otherwise unexported type that won't have its methods documented.
func ExampleDatabaseStat() {
	var (
		stat rawdb.DatabaseStat
		size common.StorageSize
	)

	stat.Add(size)

	var (
		_ string = stat.Size()  // sum of all values passed to Add()
		_ string = stat.Count() // number of calls to Add()
	)
}
