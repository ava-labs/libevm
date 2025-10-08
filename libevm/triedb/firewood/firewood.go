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
	"errors"
	"runtime"

	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/libevm/libevm/stateconf"
	"github.com/ava-labs/libevm/trie/trienode"
	"github.com/ava-labs/libevm/trie/triestate"
	"github.com/ava-labs/libevm/triedb"
)

var _ triedb.DBOverride = (*database)(nil)

type database struct {
	triedb.DBOverride // TODO(alarso16) remove once this type implements the interface
}

func (db *database) Update(root, parent common.Hash, block uint64, nodes *trienode.MergedNodeSet, states *triestate.Set, opts ...stateconf.TrieDBUpdateOption) error {
	// TODO(alarso16)
	var _ *proposals = extras.MergedNodeSet.Get(nodes)

	db.afterUpdate(nodes) // MUST be the last statement before the final return
	return errors.New("unimplemented")
}

// afterUpdate MUST be called at the end of [database.Update] to ensure that the
// Rust handle isn't freed any earlier. This is an overly cautious, defensive
// approach that will make Rustaceans scream "I told you so".
func (db *database) afterUpdate(nodes *trienode.MergedNodeSet) {
	runtime.KeepAlive(extras.MergedNodeSet.Get(nodes))
}
