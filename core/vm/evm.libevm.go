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

package vm

import (
	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/libevm/libevm"
	"github.com/ava-labs/libevm/log"
)

// canCreateContract is a convenience wrapper for calling the
// [params.RulesHooks.CanCreateContract] hook.
func (evm *EVM) canCreateContract(caller ContractRef, contractToCreate common.Address, gas uint64) (remainingGas uint64, _ error) {
	addrs := &libevm.AddressContext{Origin: evm.Origin, Caller: caller.Address(), Self: contractToCreate}
	gas, err := evm.chainRules.Hooks().CanCreateContract(addrs, gas, evm.StateDB)

	// NOTE that this block only performs logging and that all paths propagate
	// `(gas, err)` unmodified.
	if err != nil {
		log.Debug(
			"Contract creation blocked by libevm hook",
			"origin", addrs.Origin,
			"caller", addrs.Caller,
			"contract", addrs.Self,
			"hooks", log.TypeOf(evm.chainRules.Hooks()),
			"reason", err,
		)
	}

	return gas, err
}

// InvalidateExecution sets the error that will be returned by
// [EVM.ExecutionInvalidated] for the length of the current transaction; i.e.
// until [EVM.Reset] is called. This is honoured by state-transition logic to
// render the execution itself void (as against reverted).
//
// This method MUST NOT be exposed in a manner that allows contracts to set
// the error; it MAY be exposed to precompiles.
func (evm *EVM) InvalidateExecution(err error) {
	evm.executionInvalidated = err
}

// ExecutionInvalidated returns the last value passed to
// [EVM.InvalidateExecution] or nil if no such call has occurred or if
// [EVM.Reset] has been called.
func (evm *EVM) ExecutionInvalidated() error {
	return evm.executionInvalidated
}
