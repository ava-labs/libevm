// Copyright 2024-2025 the libevm authors.
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
	"errors"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/libevm/stateconf"
	"github.com/ethereum/go-ethereum/trie/trienode"
)

func TestDBOverride(t *testing.T) {
	config := &Config{
		DBOverride: func(ethdb.Database) DBOverride {
			return override{}
		},
	}

	db := NewDatabase(nil, config)
	switch got := db.Backend().(type) {
	case override:
		// woot
	default:
		t.Errorf("with non-nil %T.DBOverride, %T.Backend() got concrete type %T; want %T", config, db, got, override{})
	}
}

type override struct {
	PathDB
}

func (override) Update(root common.Hash, parent common.Hash, block uint64, nodes *trienode.MergedNodeSet, states *StateSet, opts ...stateconf.TrieDBUpdateOption) error {
	return errors.New("unimplemented")
}
