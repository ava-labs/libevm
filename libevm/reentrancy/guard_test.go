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

package reentrancy

import (
	"testing"

	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/libevm/core/vm"
	"github.com/ava-labs/libevm/libevm"
	"github.com/ava-labs/libevm/libevm/ethtest"
	"github.com/ava-labs/libevm/libevm/hookstest"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/require"
	"gotest.tools/assert"
)

func TestGuard(t *testing.T) {
	sut := common.HexToAddress("7E57ED")
	eve := common.HexToAddress("BAD")
	eveCalled := false

	zero := func() *uint256.Int {
		return uint256.NewInt(0)
	}

	returnIfGuarded := []byte("guarded")

	hooks := &hookstest.Stub{
		PrecompileOverrides: map[common.Address]libevm.PrecompiledContract{
			eve: vm.NewStatefulPrecompile(func(env vm.PrecompileEnvironment, input []byte) (ret []byte, err error) {
				eveCalled = true
				return env.Call(sut, []byte{}, env.Gas(), zero()) // i.e. reenter
			}),
			sut: vm.NewStatefulPrecompile(func(env vm.PrecompileEnvironment, input []byte) (ret []byte, err error) {
				// The argument is optional and used only to allow more than one
				// guard in a contract.
				if err := Guard(env, nil); err != nil {
					return returnIfGuarded, err
				}
				if env.Addresses().EVMSemantic.Caller == eve {
					// A real precompile MUST NOT panic under any circumstances.
					// It is done here to avoid a loop should the guard not
					// work.
					panic("reentrancy")
				}
				return env.Call(eve, []byte{}, env.Gas(), zero())
			}),
		},
	}
	hooks.Register(t)

	_, evm := ethtest.NewZeroEVM(t)
	got, _, err := evm.Call(vm.AccountRef{}, sut, []byte{}, 1e6, zero())
	require.True(t, eveCalled, "Malicious contract called")
	// The error is propagated Guard() -> reentered SUT -> Eve -> top-level SUT -> evm.Call()
	// This MUST NOT be [assert.ErrorIs] as such errors are never wrapped in geth.
	assert.Equal(t, err, vm.ErrExecutionReverted, "Precompile reverted")
	assert.Equal(t, returnIfGuarded, got, "Precompile reverted with expected data")
}
