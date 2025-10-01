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

// The firewood package provides a [triedb.BackendDB] based on [Firewood].
//
// [Firewood]: https://github.com/ava-labs/firewood
package firewood

import "github.com/ava-labs/libevm/triedb"

// Protects the import in this file so [triedb.BackendDB] linked comments work.
var _ triedb.BackendDB
