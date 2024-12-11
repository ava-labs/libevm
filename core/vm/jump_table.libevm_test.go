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

package vm_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/libevm/core/vm"
	"github.com/ava-labs/libevm/libevm/ethtest"
	"github.com/ava-labs/libevm/params"
)

type vmHooksStub struct {
	replacement *vm.JumpTable
	overridden  bool
}

var _ vm.Hooks = (*vmHooksStub)(nil)

// OverrideJumpTable overrides all non-nil operations from s.replacement .
func (s *vmHooksStub) OverrideJumpTable(_ params.Rules, jt *vm.JumpTable) *vm.JumpTable {
	s.overridden = true
	for op, instr := range s.replacement {
		if instr != nil {
			fmt.Println(op, instr)
			jt[op] = instr
		}
	}
	return jt
}

func (*vmHooksStub) OverrideNewEVMArgs(a *vm.NewEVMArgs) *vm.NewEVMArgs { return a }

func (*vmHooksStub) OverrideEVMResetArgs(r params.Rules, a *vm.EVMResetArgs) *vm.EVMResetArgs {
	return a
}

// An opRecorder is an instruction that records its inputs.
type opRecorder struct {
	stateVal common.Hash
}

func (op *opRecorder) execute(env vm.Environment, pc *uint64, interpreter *vm.EVMInterpreter, scope *vm.ScopeContext) ([]byte, error) {
	op.stateVal = env.StateDB().GetState(scope.Contract.Address(), common.Hash{})
	return nil, nil
}

func TestOverrideJumpTable(t *testing.T) {
	const (
		opcode          = 1
		gasLimit uint64 = 1e6
	)
	rng := ethtest.NewPseudoRand(142857)
	gasCost := 1 + rng.Uint64n(gasLimit)
	spy := &opRecorder{}

	vmHooks := &vmHooksStub{
		replacement: &vm.JumpTable{
			opcode: vm.OperationBuilder{
				Execute:     spy.execute,
				ConstantGas: gasCost,
				MemorySize: func(s *vm.Stack) (size uint64, overflow bool) {
					return 0, false
				},
			}.Build(),
		},
	}
	vm.RegisterHooks(vmHooks)

	state, evm := ethtest.NewZeroEVM(t)

	contract := rng.Address()
	state.CreateAccount(contract)
	state.SetCode(contract, []byte{opcode})
	value := rng.Hash()
	state.SetState(contract, common.Hash{}, value)

	_, gasRemaining, err := evm.Call(vm.AccountRef(rng.Address()), contract, []byte{}, gasLimit, uint256.NewInt(0))
	require.NoError(t, err, "evm.Call([contract with overridden opcode])")
	assert.Equal(t, gasLimit-gasCost, gasRemaining, "gas remaining")
	assert.Equal(t, spy.stateVal, value, "StateDB propagated")
}

func TestOperationFieldCount(t *testing.T) {
	// The libevm OperationBuilder assumes that the 6 struct fields are the only
	// ones.
	op := vm.OperationBuilder{}.Build()
	require.Equalf(t, 6, reflect.TypeOf(*op).NumField(), "number of fields in %T struct", *op)
}
