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

package vm

// An OperationBuilder is a factory for a new operations to include in a
// [JumpTable].
type OperationBuilder struct {
	Execute            OperationFunc
	ConstantGas        uint64
	DynamicGas         func(_ *EVM, _ *Contract, _ *Stack, _ *Memory, requestedMemorySize uint64) (uint64, error)
	MinStack, MaxStack int
	MemorySize         func(s *Stack) (size uint64, overflow bool)
}

// Build constructs the operation.
func (b OperationBuilder) Build() *operation {
	o := &operation{
		execute:     b.Execute.internal(),
		constantGas: b.ConstantGas,
		dynamicGas:  b.DynamicGas,
		minStack:    b.MinStack,
		maxStack:    b.MaxStack,
		memorySize:  b.MemorySize,
	}
	return o
}

// An OperationFunc is the execution function of a custom instruction.
type OperationFunc func(_ Environment, pc *uint64, _ *EVMInterpreter, _ *ScopeContext) ([]byte, error)

// internal converts an exported [OperationFunc] into an un-exported
// [executionFunc] as required to build an [operation].
func (fn OperationFunc) internal() executionFunc {
	return func(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
		env := &environment{
			evm:  interpreter.evm,
			self: scope.Contract,
			// The CallType isn't exposed by an instruction's [Environment] and,
			// although [UnknownCallType] is the default value, it's explicitly
			// set to avoid future accidental setting without proper
			// justification.
			callType: UnknownCallType,
		}
		return fn(env, pc, interpreter, scope)
	}
}
