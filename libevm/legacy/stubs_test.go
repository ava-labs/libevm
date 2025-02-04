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

package legacy

import (
	"github.com/ava-labs/libevm/core/vm"
)

var _ vm.PrecompileEnvironment = (*stubPrecompileEnvironment)(nil)

// stubPrecompileEnvironment implements [vm.PrecompileEnvironment] for testing.
type stubPrecompileEnvironment struct {
	vm.PrecompileEnvironment
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
