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

package firewood

import (
	"fmt"

	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/libevm/trie/trienode"
)

// RegisterExtras
func RegisterExtras() {
	extras = trienode.RegisterExtras[proposal, proposal, struct{}]()
}

var extras trienode.ExtraPayloads[*proposal, *proposal, struct{}]

type proposal struct {
	handle *handle
}

// TODO(alarso16) this type is entirely arbitrary and exists only to allow
// initial integration testing.
type handle struct {
	root        common.Hash
	memoryFreed bool
}

// MergeNodeSet implements [trienode.MergedNodeSetHooks], copying at most one
// proposal handle into the merged set.
func (h *proposal) MergeNodeSet(into *trienode.MergedNodeSet, set *trienode.NodeSet) error {
	merged := extras.MergedNodeSet.Get(into)
	if merged.handle != nil {
		return fmt.Errorf(">1 %T carrying non-nil %T", set, merged.handle)
	}
	merged.handle = extras.NodeSet.Get(set).handle
	return nil
}

// AddNode implements [trienode.NodeSetHooks] as a noop.
func (h *proposal) AddNode(*trienode.NodeSet, []byte, *trienode.Node) {}
