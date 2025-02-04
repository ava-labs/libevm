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

// Package legacy provides converters between legacy types and their refactored
// equivalents.

package legacy

import (
	"math/big"

	"github.com/holiman/uint256"

	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/libevm/core/types"
	"github.com/ava-labs/libevm/core/vm"
	"github.com/ava-labs/libevm/libevm"
	"github.com/ava-labs/libevm/params"
)

var _ vm.PrecompileEnvironment = (*stubPrecompileEnvironment)(nil)

// stubPrecompileEnvironment implements [vm.PrecompileEnvironment] for testing.
type stubPrecompileEnvironment struct {
	gasToReturn uint64
	gasUsed     uint64
}

// Gas returns the gas supplied to the precompile.
func (s *stubPrecompileEnvironment) Gas() uint64 {
	return s.gasToReturn
}

// UseGas records the gas used by the precompile.
func (s *stubPrecompileEnvironment) UseGas(gas uint64) bool {
	s.gasUsed += gas
	return true
}

func (s *stubPrecompileEnvironment) Call(addr common.Address, input []byte, gas uint64, value *uint256.Int, _ ...vm.CallOption) (ret []byte, _ error) {
	return nil, nil
}

func (s *stubPrecompileEnvironment) ChainConfig() *params.ChainConfig         { return nil }
func (s *stubPrecompileEnvironment) Rules() params.Rules                      { return params.Rules{} }
func (s *stubPrecompileEnvironment) StateDB() vm.StateDB                      { return nil }
func (s *stubPrecompileEnvironment) ReadOnlyState() libevm.StateReader        { return nil }
func (s *stubPrecompileEnvironment) IncomingCallType() vm.CallType            { return vm.Call }
func (s *stubPrecompileEnvironment) Addresses() *libevm.AddressContext        { return nil }
func (s *stubPrecompileEnvironment) ReadOnly() bool                           { return false }
func (s *stubPrecompileEnvironment) Value() *uint256.Int                      { return nil }
func (s *stubPrecompileEnvironment) BlockHeader() (h types.Header, err error) { return h, nil }
func (s *stubPrecompileEnvironment) BlockNumber() *big.Int                    { return nil }
func (s *stubPrecompileEnvironment) BlockTime() uint64                        { return 0 }
