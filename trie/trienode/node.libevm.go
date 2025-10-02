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

// MergedNodeSetHooks are called as part of standard [MergedNodeSet] behaviour.
type MergedNodeSetHooks interface {
	AfterMergeNodeSet(into *MergedNodeSet, _ *NodeSet) error
}

// NodeSetHooks are called as part of standard [NodeSet] behaviour.
type NodeSetHooks interface {
	AfterAddNode(into *NodeSet, path []byte, _ *Node)
}

// RegisterExtras registers types `MNSPtr`, `NSPtr`, and `N` to be carried as
// extra payloads in [MergedNodeSet], [NodeSet], and [Node] objects
// respectively. It MUST NOT be called more than once.
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
	NPtr interface{ *N },
]() ExtraPayloads[MNSPtr, NSPtr, NPtr] {
	payloads := ExtraPayloads[MNSPtr, NSPtr, NPtr]{
		MergedNodeSet: pseudo.NewAccessor[*MergedNodeSet, MNSPtr](
			(*MergedNodeSet).extraPayload,
			func(s *MergedNodeSet, t *pseudo.Type) { s.extra = t },
		),
		NodeSet: pseudo.NewAccessor[*NodeSet, NSPtr](
			(*NodeSet).extraPayload,
			func(s *NodeSet, t *pseudo.Type) { s.extra = t },
		),
		Node: pseudo.NewAccessor[*Node, NPtr](
			(*Node).extraPayload,
			func(n *Node, t *pseudo.Type) { n.extra = t },
		),
	}

	registeredExtras.MustRegister(&extraConstructors{
		newMergedNodeSet: pseudo.NewConstructor[MNS]().NewPointer, // i.e. non-nil MNSPtr
		newNodeSet:       pseudo.NewConstructor[NS]().NewPointer,  // i.e. non-nil NSPtr
		newNode:          pseudo.NewConstructor[N]().NewPointer,   // i.e. non-nil N
		hooks:            payloads,
	})

	return payloads
}

// TestOnlyClearRegisteredExtras clears any previous call to [RegisterExtras].
// It panics if called from a non-testing call stack.
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
// is held that no duplicated set belonging to the same trie will be merged
// twice.
func (set *MergedNodeSet) Merge(other *NodeSet) error {
	if err := set.merge(other); err != nil {
		return err
	}
	if r := registeredExtras; r.Registered() {
		return r.Get().hooks.hooksFromMNS(set).AfterMergeNodeSet(set, other)
	}
	return nil
}

func (set *NodeSet) mergePayload(path []byte, n *Node) {
	if r := registeredExtras; r.Registered() {
		r.Get().hooks.hooksFromNS(set).AfterAddNode(set, path, n)
	}
}

// ExtraPayloads provides strongly typed access to the extra payloads carried by
// [MergedNodeSet], [NodeSet], and [Node] ojects. The only valid way to
// construct an instance is by a call to [RegisterExtras]. The default `MNSPtr`
// and `NSPtr` default values, returned by [pseudo.Accessor.Get] are guaranteed
// to be non-nil pointers to zero values, equivalent to, e.g. `new(MNS)`.
type ExtraPayloads[
	MNSPtr MergedNodeSetHooks,
	NSPtr NodeSetHooks,
	NPtr any,
] struct {
	MergedNodeSet pseudo.Accessor[*MergedNodeSet, MNSPtr]
	NodeSet       pseudo.Accessor[*NodeSet, NSPtr]
	Node          pseudo.Accessor[*Node, NPtr]
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
