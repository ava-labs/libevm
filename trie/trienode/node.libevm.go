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
	"github.com/ava-labs/libevm/libevm/pseudo"
	"github.com/ava-labs/libevm/libevm/register"
)

// MergedNodeSetHooks
type MergedNodeSetHooks interface {
	Merge(into *MergedNodeSet, _ *NodeSet) error
}

// NodeSetHooks
type NodeSetHooks interface {
	Add(into *NodeSet, path []byte, _ *Node)
}

// RegisterExtras
func RegisterExtras[
	MNS, NS, N any,
	MNSPtr interface {
		MergedNodeSetHooks
		*MNS
	},
	NSPtr interface {
		NodeSetHooks
		*NS
	},
]() ExtraPayloads[MNSPtr, NSPtr, N] {
	payloads := ExtraPayloads[MNSPtr, NSPtr, N]{
		MergedNodeSet: pseudo.NewAccessor[*MergedNodeSet, MNSPtr](
			(*MergedNodeSet).extraPayload,
			func(s *MergedNodeSet, t *pseudo.Type) { s.extra = t },
		),
		NodeSet: pseudo.NewAccessor[*NodeSet, NSPtr](
			(*NodeSet).extraPayload,
			func(s *NodeSet, t *pseudo.Type) { s.extra = t },
		),
		Node: pseudo.NewAccessor[*Node, N](
			(*Node).extraPayload,
			func(n *Node, t *pseudo.Type) { n.extra = t },
		),
	}

	registeredExtras.MustRegister(&extraConstructors{
		newMergedNodeSet: pseudo.NewConstructor[MNS]().NewPointer,
		newNodeSet:       pseudo.NewConstructor[NS]().NewPointer,
		newNode:          pseudo.NewConstructor[N]().Zero,
		hooks:            payloads,
	})

	return payloads
}

// TestOnlyClearRegisteredExtras
func TestOnlyClearRegisteredExtras() {
	registeredExtras.TestOnlyClear()
}

var registeredExtras register.AtMostOnce[*extraConstructors]

type extraConstructors struct {
	newMergedNodeSet func() *pseudo.Type
	newNodeSet       func() *pseudo.Type
	newNode          func() *pseudo.Type
	hooks            interface {
		hooksFromMNS(*MergedNodeSet) MergedNodeSetHooks
		hooksFromNS(*NodeSet) NodeSetHooks
	}
}

// Merge merges the provided dirty nodes of a trie into the set. The assumption
// is held that no duplicated set belonging to the same trie will be merged twice.
func (set *MergedNodeSet) Merge(other *NodeSet) error {
	if err := set.merge(other); err != nil {
		return err
	}
	if r := registeredExtras; r.Registered() {
		return r.Get().hooks.hooksFromMNS(set).Merge(set, other)
	}
	return nil
}

func (set *NodeSet) mergePayload(path []byte, n *Node) {
	if r := registeredExtras; r.Registered() {
		r.Get().hooks.hooksFromNS(set).Add(set, path, n)
	}
}

// ExtraPayloads
type ExtraPayloads[
	MNS MergedNodeSetHooks,
	NS NodeSetHooks,
	N any,
] struct {
	MergedNodeSet pseudo.Accessor[*MergedNodeSet, MNS]
	NodeSet       pseudo.Accessor[*NodeSet, NS]
	Node          pseudo.Accessor[*Node, N]
}

func (e ExtraPayloads[MNS, NS, N]) hooksFromMNS(s *MergedNodeSet) MergedNodeSetHooks {
	return e.MergedNodeSet.Get(s)
}

func (e ExtraPayloads[MNS, NS, N]) hooksFromNS(s *NodeSet) NodeSetHooks {
	return e.NodeSet.Get(s)
}

func extraPayloadOrSetDefault(field **pseudo.Type, construct func(*extraConstructors) *pseudo.Type) *pseudo.Type {
	r := registeredExtras
	if !r.Registered() {
		// See params.ChainConfig.extraPayload() for panic rationale.
		panic("<T>.extraPayload() called before RegisterExtras()")
	}
	if *field == nil {
		*field = construct(r.Get())
	}
	return *field
}

func (set *MergedNodeSet) extraPayload() *pseudo.Type {
	return extraPayloadOrSetDefault(&set.extra, func(c *extraConstructors) *pseudo.Type {
		return c.newMergedNodeSet()
	})
}

func (set *NodeSet) extraPayload() *pseudo.Type {
	return extraPayloadOrSetDefault(&set.extra, func(c *extraConstructors) *pseudo.Type {
		return c.newNodeSet()
	})
}

func (n *Node) extraPayload() *pseudo.Type {
	return extraPayloadOrSetDefault(&n.extra, func(c *extraConstructors) *pseudo.Type {
		return c.newNode()
	})
}
