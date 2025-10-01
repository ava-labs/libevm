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

package trienode

import (
	"maps"
	"testing"

	"github.com/ava-labs/libevm/common"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
)

type nodePayload struct {
	x uint64
}

type setPayload struct {
	added map[string]uint64
}

func (p *setPayload) Add(_ *NodeSet, path []byte, n *Node) {
	if p.added == nil {
		p.added = make(map[string]uint64)
	}
	p.added[string(path)] = extras.Node.Get(n).x
}

type mergedSetPayload struct {
	merged []map[string]uint64
}

func (p *mergedSetPayload) Merge(_ *MergedNodeSet, ns *NodeSet) error {
	p.merged = append(p.merged, maps.Clone(extras.NodeSet.Get(ns).added))
	return nil
}

var extras ExtraPayloads[*mergedSetPayload, *setPayload, nodePayload]

func TestExtras(t *testing.T) {
	extras = RegisterExtras[mergedSetPayload, setPayload, nodePayload]()
	t.Cleanup(TestOnlyClearRegisteredExtras)

	n1 := New(common.Hash{0}, nil)
	extras.Node.Set(n1, nodePayload{x: 1})
	n42 := New(common.Hash{1}, nil)
	extras.Node.Set(n42, nodePayload{x: 42})

	set := NewNodeSet(common.Hash{})
	merge := NewMergedNodeSet()
	set.AddNode([]byte("n1"), n1)
	require.NoError(t, merge.Merge(set))

	set.AddNode([]byte("n42"), n42)
	require.NoError(t, merge.Merge(set))

	got := extras.MergedNodeSet.Get(merge).merged
	want := []map[string]uint64{
		{
			"n1": 1,
		},
		{
			"n1":  1,
			"n42": 42,
		},
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("%T payload diff (-want +got):\n%s", merge, diff)
	}
}
