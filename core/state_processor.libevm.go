// Copyright 2026 the libevm authors.
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

package core

import (
	"encoding/binary"

	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/libevm/core/state"
	"github.com/ava-labs/libevm/core/types"
	"github.com/ava-labs/libevm/params"
)

var beaconRootsCodeHash = common.HexToHash(`0xf57acd40259872606d76197ef052f3d35588dadf919ee1f0e3cb9b62d3f4b02c`)

// SetBeaconBlockRoot is equivalent to [ProcessBeaconBlockRoot] except that it
// acts directly on the [state.StateDB] instead of calling the EIP-4788 `set()`
// function. If [types.Header.ParentBeaconRoot] is nil or the EIP-4788 contract
// hasn't been deployed, then this function is a no-op.
func SetBeaconBlockRoot(sdb *state.StateDB, hdr *types.Header) {
	if hdr.ParentBeaconRoot == nil || sdb.GetCodeHash(params.BeaconRootsStorageAddress) != beaconRootsCodeHash {
		return
	}
	const bufferLen = 8191
	timeIdx := hdr.Time % bufferLen
	rootIdx := timeIdx + bufferLen

	sdb.SetState(params.BeaconRootsStorageAddress, uint64ToHash(timeIdx), uint64ToHash(hdr.Time))
	sdb.SetState(params.BeaconRootsStorageAddress, uint64ToHash(rootIdx), *hdr.ParentBeaconRoot)
	sdb.Finalise(true)
}

func uint64ToHash(x uint64) (h common.Hash) {
	binary.BigEndian.PutUint64(h[24:], x)
	return
}
