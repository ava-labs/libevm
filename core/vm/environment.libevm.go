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
	// TODO(arr4n) remove this and export the function when reviewing and
	// merging PR 277.
	opts = append(opts, legacyOnlyDisableEIP150Gas64th())

	ret, _, err := e.callOrCreateContract(Call, &addr, input, gas, value, nil /*salt*/, opts...)
	return ret, err
}

func (e *environment) Create(code []byte, value *uint256.Int) ([]byte, common.Address, error) {
	return e.callOrCreateContract(create, nil /*to*/, code, e.Gas(), value, nil /*salt*/)
}

func (e *environment) Create2(code []byte, value, salt *uint256.Int) ([]byte, common.Address, error) {
	return e.callOrCreateContract(create2, nil /*to*/, code, e.Gas(), value, salt)
}

func (e *environment) callOrCreateContract(typ CallType, addr *common.Address, input []byte, gas uint64, value, salt *uint256.Int, opts ...CallOption) ([]byte, common.Address, error) {
	conf := options.As[callConfig](opts...)

	if e.Rules().IsEIP150 && !conf.legacyOnlyNoEIP150Gas64th {
		gas = min(gas, e.Gas()-e.Gas()/64)
	}

	var caller ContractRef = e.self
	if conf.unsafeCallerAddressProxying {
		// Note that, in addition to being unsafe, this breaks an EVM
		// assumption that the caller ContractRef is always a *Contract.
		caller = AccountRef(e.self.CallerAddress)
		if e.callType == DelegateCall {
			// self was created with AsDelegate(), which means that
			// CallerAddress was inherited.
			caller = AccountRef(e.self.Address())
		}
	}

	writes := (value != nil && !value.IsZero()) || typ == create || typ == create2
	if e.ReadOnly() && writes {
		return nil, common.Address{}, ErrWriteProtection
	}
	if !e.UseGas(gas) {
		return nil, common.Address{}, ErrOutOfGas
	}

	var (
		frameRet    []byte
		created     common.Address
		leftOverGas uint64
		frameErr    error
	)
	switch typ {
	case Call:
		frameRet, leftOverGas, frameErr = e.evm.Call(caller, *addr, input, gas, value)

	case create:
		frameRet, created, leftOverGas, frameErr = e.evm.Create(caller, input, gas, value)

	case create2:
		frameRet, created, leftOverGas, frameErr = e.evm.Create2(caller, input, gas, value, salt)

	case CallCode, DelegateCall, StaticCall:
		// TODO(arr4n): these cases should be very similar to CALL, hence the
		// early abstraction, to signal to future maintainers. If implementing
		// them, there's likely no need to honour the
		// [callOptUNSAFECallerAddressProxy] because it's purely for backwards
		// compatibility, however the "callTracer" test MUST be extended to
		// demonstrate the correct type.
		fallthrough
	default:
		return nil, common.Address{}, fmt.Errorf("unimplemented precompile call type %v", typ)
	}

	if err := e.refundGas(leftOverGas); err != nil {
		return nil, common.Address{}, err
	}
	return frameRet, created, frameErr
}
