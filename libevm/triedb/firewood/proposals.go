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
	"github.com/ava-labs/firewood-go-ethhash/ffi"
	"github.com/ava-labs/libevm/common"
)

// A proposal carries a Firewood FFI proposal (i.e. Rust-owned memory).
// The Firewood library adds a finalizer to the proposal handle to ensure that
// the memory is freed when the Go object is garbage collected. However, because
// we form a tree of proposals, the `proposal.Proposal` field may be the only
// reference to a given proposal. To ensure that all proposals in the tree
// can be freed in a finalizer, this cannot be included in the tree structure.
type proposal struct {
	info     *blockInfo
	proposal *ffi.Proposal
}

type blockInfo struct {
	parent   *blockInfo
	children []*blockInfo
	hashes   map[common.Hash]struct{} // All corresponding block hashes
	root     common.Hash
	height   uint64
}
