// Copyright 2024 the libevm authors.
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

package native

import (
	"math/big"

	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/libevm/core/tracing"
)

func init() {
	var (
		p *prestateTracer
		_ tracing.EnterHook = p.OnEnter
	)
}

// OnEnter implements [tracing.EnterHook].
func (t *prestateTracer) OnEnter(depth int, typ byte, from common.Address, to common.Address, input []byte, gas uint64, value *big.Int) {
	// Although [prestateTracer.lookupStorage] expects
	// [prestateTracer.lookupAccount] to have been called, the invariant is
	// maintained by [prestateTracer.OnOpcode] when it encounters an OpCode
	// corresponding to scope entry. This, however, doesn't work when using a
	// call method exposed by [vm.PrecompileEnvironment], and is restored by a
	// call to this OnEnter implementation. Note that lookupAccount(x) is
	// idempotent.
	t.lookupAccount(to)
}
