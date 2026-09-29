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

import (
	"fmt"
	"math/big"

	"github.com/holiman/uint256"

	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/libevm/common/math"
	"github.com/ava-labs/libevm/core/types"
	"github.com/ava-labs/libevm/libevm"
	"github.com/ava-labs/libevm/libevm/options"
	"github.com/ava-labs/libevm/params"
)

var _ PrecompileEnvironment = (*environment)(nil)

type environment struct {
	evm      *EVM
	self     *Contract
	callType CallType

	rawSelf, rawCaller common.Address
}

func (e *environment) Gas() uint64            { return e.self.Gas }
func (e *environment) UseGas(gas uint64) bool { return e.self.UseGas(gas) }
func (e *environment) Value() *uint256.Int    { return new(uint256.Int).Set(e.self.Value()) }

func (e *environment) ChainConfig() *params.ChainConfig  { return e.evm.chainConfig }
func (e *environment) Rules() params.Rules               { return e.evm.chainRules }
func (e *environment) ReadOnlyState() libevm.StateReader { return e.evm.StateDB }
func (e *environment) IncomingCallType() CallType        { return e.callType }
func (e *environment) BlockNumber() *big.Int             { return new(big.Int).Set(e.evm.Context.BlockNumber) }
func (e *environment) BlockTime() uint64                 { return e.evm.Context.Time }

func (e *environment) InvalidateExecution(err error) { e.evm.InvalidateExecution(err) }

func (e *environment) refundGas(add uint64) error {
	gas, overflow := math.SafeAdd(e.self.Gas, add)
	if overflow {
		return ErrGasUintOverflow
	}
	e.self.Gas = gas
	return nil
}

func (e *environment) ReadOnly() bool {
	return e.evm.interpreter.readOnly
}

func (e *environment) Addresses() *libevm.AddressContext {
	return &libevm.AddressContext{
		Origin: e.evm.Origin,
		EVMSemantic: libevm.CallerAndSelf{
			Caller: e.self.CallerAddress,
			Self:   e.self.Address(),
		},
		Raw: &libevm.CallerAndSelf{
			Caller: e.rawCaller,
			Self:   e.rawSelf,
		},
	}
}

func (e *environment) StateDB() StateDB {
	if e.ReadOnly() {
		return nil
	}
	return e.evm.StateDB
}

func (e *environment) BlockHeader() (types.Header, error) {
	hdr := e.evm.Context.Header
	if hdr == nil {
		// Although [core.NewEVMBlockContext] sets the field and is in the
		// typical hot path (e.g. miner), there are other ways to create a
		// [vm.BlockContext] (e.g. directly in tests) that may result in no
		// available header.
		return types.Header{}, fmt.Errorf("nil %T in current %T", hdr, e.evm.Context)
	}
	return *hdr, nil
}

func (e *environment) Call(addr common.Address, input []byte, gas uint64, value *uint256.Int, opts ...CallOption) ([]byte, error) {
	return e.callContract(Call, addr, input, gas, value, opts...)
}

func (e *environment) callContract(typ CallType, addr common.Address, input []byte, gas uint64, value *uint256.Int, opts ...CallOption) ([]byte, error) {
	if value == nil {
		// STATIC- and DELEGATECALL pass nil values, which we would otherwise
		// have to check repeatedly.
		value = uint256.NewInt(0)
	}
	if e.ReadOnly() && !value.IsZero() {
		return nil, ErrWriteProtection
	}

	cfg := options.As[callConfig](opts...)

	var caller ContractRef = e.self
	if cfg.unsafeCallerAddressProxying {
		// Note that, in addition to being unsafe, this breaks an EVM
		// assumption that the caller ContractRef is always a *Contract.
		caller = AccountRef(e.self.CallerAddress)
		if e.callType == DelegateCall {
			// self was created with AsDelegate(), which means that
			// CallerAddress was inherited.
			caller = AccountRef(e.self.Address())
		}
	}

	gas, err := e.buyCallGas(typ, cfg, addr, gas, value)
	if err != nil {
		return nil, err
	}

	switch typ {
	case Call:
		ret, returnGas, callErr := e.evm.Call(caller, addr, input, gas, value)
		if err := e.refundGas(returnGas); err != nil {
			return nil, err
		}
		return ret, callErr
	case CallCode, DelegateCall, StaticCall:
		// TODO(arr4n): these cases should be very similar to CALL, hence the
		// early abstraction, to signal to future maintainers. If implementing
		// them, there's likely no need to honour the
		// [callOptUNSAFECallerAddressProxy] because it's purely for backwards
		// compatibility, however the "callTracer" test MUST be extended to
		// demonstrate the correct type.
		fallthrough
	default:
		return nil, fmt.Errorf("unimplemented precompile call type %v", typ)
	}
}

func (e *environment) buyCallGas(typ CallType, cfg *callConfig, addr common.Address, gas uint64, value *uint256.Int) (bought uint64, retErr error) {
	if cfg.legacyOutboundCallGas {
		if !e.UseGas(gas) {
			return 0, ErrOutOfGas
		}
		return gas, nil
	}

	// All dynamic-gas calculators for calls store the amount of gas to
	// propagate in [EVM.callGasTemp] because the [gasFunc] signature doesn't
	// allow for it to be returned.
	old := e.evm.callGasTemp
	stack := newstack()
	defer func() {
		e.evm.callGasTemp = old
		returnStack(stack)
	}()

	// All *CALL op codes have [gas, address] on the top of the stack, while
	// CALL and CALLCODE then have the value, while STATICCALL and DELEGATECALL
	// have the argument offset, which doesn't affect gas.
	stack.push(value)
	stack.push(new(uint256.Int).SetBytes20(addr[:]))
	stack.push(uint256.NewInt(gas))

	// Constant gas cost MUST be charged first otherwise the 63/64 rule of the
	// dynamic cost will be applied to the incorrect value.
	op := e.evm.interpreter.table[typ.OpCode()]
	if !e.UseGas(op.constantGas) {
		return 0, ErrOutOfGas
	}

	// Dynamic-gas calculation might warm the address before we've actually paid
	// for the associated gas. We revert the warming in all error cases, even if
	// it has been paid for inside [operation.dynamicGas], because this is the
	// more conservative approach for DoS protection. The alternative is a
	// convoluted set of checks to see if the address was warmed and whether the
	// gas was used, which is brittle under upstream (geth) code mergers.
	sdb := e.evm.StateDB
	beforeAddrWarming := sdb.Snapshot()
	defer func() {
		if retErr != nil {
			sdb.RevertToSnapshot(beforeAddrWarming)
		}
	}()

	dyn, err := op.dynamicGas(e.evm, e.self, stack, NewMemory(), 0 /*memory expansion*/)
	if err != nil {
		return 0, err
	}
	if !e.UseGas(dyn) {
		return 0, ErrOutOfGas
	}

	// [operation.dynamicGas] for *CALL returns a total that already includes
	// [EVM.callGasTemp], so the propagated gas was charged above.
	bought = e.evm.callGasTemp
	if !value.IsZero() {
		bought += params.CallStipend
	}
	return bought, nil
}
