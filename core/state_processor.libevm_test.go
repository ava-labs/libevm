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

package core_test

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"testing"

	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/libevm/core"
	"github.com/ava-labs/libevm/core/state"
	"github.com/ava-labs/libevm/core/types"
	"github.com/ava-labs/libevm/core/vm"
	"github.com/ava-labs/libevm/crypto"
	"github.com/ava-labs/libevm/libevm/ethtest"
	"github.com/ava-labs/libevm/params"
)

func TestSetBeaconBlockRoot(t *testing.T) {
	for time := uint64(1); time < 1<<14; time += 100 {
		root := crypto.Keccak256Hash(binary.BigEndian.AppendUint64(nil, time))
		tests := []struct {
			name  string
			setup func(*vm.EVM, *state.StateDB)
		}{
			{
				name: "SetBeaconBlockRoot_SUT",
				setup: func(_ *vm.EVM, sdb *state.StateDB) {
					hdr := &types.Header{
						Time:             time,
						ParentBeaconRoot: &root,
					}
					core.SetBeaconBlockRoot(sdb, hdr)
				},
			},
			{
				name: "ProcessBeaconBlockRoot_gold_standard",
				setup: func(evm *vm.EVM, sdb *state.StateDB) {
					core.ProcessBeaconBlockRoot(root, evm, sdb)
				},
			},
		}

		var gotStateRoots [2]common.Hash

		for i, tt := range tests {
			t.Run(fmt.Sprintf("%s_time_%d", tt.name, time), func(t *testing.T) {
				sdb, evm := ethtest.NewZeroEVM(t,
					ethtest.WithBlockContext(vm.BlockContext{
						CanTransfer: core.CanTransfer,
						Transfer:    core.Transfer,
						BlockNumber: big.NewInt(1),
						Time:        time,
						Random:      &common.Hash{}, // implies post-Merge, required for PUSH0
					}),
					ethtest.WithChainConfig(params.MergedTestChainConfig),
				)

				_, addr, _, err := evm.Create(
					// https://eips.ethereum.org/EIPS/eip-4788
					vm.AccountRef(common.HexToAddress(`0x0B799C86a49DEeb90402691F1041aa3AF2d3C875`)),
					common.FromHex(`0x60618060095f395ff33373fffffffffffffffffffffffffffffffffffffffe14604d57602036146024575f5ffd5b5f35801560495762001fff810690815414603c575f5ffd5b62001fff01545f5260205ff35b5f5ffd5b62001fff42064281555f359062001fff015500`),
					0x3d090,
					uint256.NewInt(0),
				)
				require.NoErrorf(t, err, "%T.Create([EIP-4788 contract])", evm)
				require.Equalf(t, params.BeaconRootsStorageAddress, addr, "%T.Create([EIP-4788 contract]) deployed address", evm)

				tt.setup(evm, sdb)
				gotStateRoots[i] = sdb.IntermediateRoot(true)

				got, _, err := evm.StaticCall(
					vm.AccountRef{},
					params.BeaconRootsStorageAddress,
					uint256.NewInt(time).PaddedBytes(32),
					1e6,
				)
				require.NoErrorf(t, err, "%T.StaticCall([EIP-4788 contract], [just-set root's time])", evm)
				assert.Equal(t, root.Bytes(), got, "%T.StaticCall([EIP-4788 contract], [just-set root's time])")
			})
		}

		assert.Equalf(t, gotStateRoots[1], gotStateRoots[0], "%T.IntermediateRoot() after SetBeaconBlockRoot() vs gold-standard ProcessBeaconBlockRoot()", &state.StateDB{})
	}
}
