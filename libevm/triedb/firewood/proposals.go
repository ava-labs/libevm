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
	"runtime"

	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/libevm/trie/trienode"
)

// RegisterExtras registers Firewood proposals with [trienode.RegisterExtras].
// This MUST be called in and only in tests / package main to avoid polluting
// other packages. A call to RegisterExtras is required for the rest of this
// package to function correctly.
func RegisterExtras() {
	extras = trienode.RegisterExtras[proposal, proposal, struct{}]()
}

var extras trienode.ExtraPayloads[*proposal, *proposal, *struct{}]

// A proposal is embedded as a payload in the [trienode.NodeSet] object returned
// by trie `Commit()`. A preceding call to [RegisterExtras] ensures that the
// proposal will be propagated to [Database.Update].
//
// After construction, [proposal.setFinalizer] SHOULD be called to ensure
// release of resources via [proposal.free] once the proposal is garbage
// collected.
type proposal struct {
	// root MUST match the argument returned by the trie's `Commit()` method.
	root common.Hash

	// TODO(alarso16) add handles etc. here and clean them up in [proposal.free]

	finalized chan struct{} // https://go.dev/doc/gc-guide#Testing_object_death
}

func (p *proposal) injectInto(ns *trienode.NodeSet) {
	extras.NodeSet.Set(ns, p)
}

// setFinalizer calls [runtime.SetFinalizer] with `p`.
func (p *proposal) setFinalizer() {
	p.finalized = make(chan struct{})
	runtime.SetFinalizer(p, (*proposal).finalizer)
}

// finalizer is expected to be passed to [runtime.SetFinalizer], abstracted as a
// method to guarantee that it doesn't accidentally capture the value being
// collected, thus resurrecting it.
func (p *proposal) finalizer() {
	p.free()
	close(p.finalized)
}

// free is called when the [proposal] is no longer reachable.
func (p *proposal) free() {
	// TODO(alarso16) free the Rust object(s).
}

// AfterMergeNodeSet implements [trienode.MergedNodeSetHooks], copying at most
// one proposal handle into the merged set.
func (h *proposal) AfterMergeNodeSet(into *trienode.MergedNodeSet, ns *trienode.NodeSet) error {
	if p := extras.MergedNodeSet.Get(into); p.root != (common.Hash{}) {
		return fmt.Errorf(">1 %T carrying non-zero %T", ns, p)
	}
	// The GC finalizer is attached to the [payload], not to the [handle], so
	// we have to copy the entire object to ensure that it remains reachable.
	extras.MergedNodeSet.Set(into, extras.NodeSet.Get(ns))
	return nil
}

// AfterAddNode implements [trienode.NodeSetHooks] as a noop.
func (h *proposal) AfterAddNode(*trienode.NodeSet, []byte, *trienode.Node) {}
