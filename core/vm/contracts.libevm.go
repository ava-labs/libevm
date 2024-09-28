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

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/libevm"
	"github.com/ethereum/go-ethereum/params"
)

// evmCallArgs mirrors the parameters of the [EVM] methods Call(), CallCode(),
// DelegateCall() and StaticCall(). Its fields are identical to those of the
// parameters, prepended with the receiver name and appended with additional
// values. As {Delegate,Static}Call don't accept a value, they MUST set the
// respective field to nil.
//
// Instantiation can be achieved by merely copying the parameter names, in
// order, which is trivially achieved with AST manipulation:
//
//	func (evm *EVM) Call(caller ContractRef, addr common.Address, input []byte, gas uint64, value *uint256.Int) ... {
//	    ...
//	    args := &evmCallArgs{evm, caller, addr, input, gas, value, false}
type evmCallArgs struct {
	evm      *EVM
	callType callType

	// args:start
	caller ContractRef
	addr   common.Address
	input  []byte
	gas    uint64
	value  *uint256.Int
	// args:end

	// evm.interpreter.readOnly is only set to true via a call to
	// EVMInterpreter.Run() so, if a precompile is called directly with
	// StaticCall(), then readOnly might not be set yet. StaticCall() MUST set
	// this to forceReadOnly and all other methods MUST set it to
	// inheritReadOnly; i.e. equivalent to the boolean they each pass to
	// EVMInterpreter.Run().

	// If a precompile issues its own Call() then caller semantics are dependent
	// on whether the precompile was delegate-called or not. DelegateCall() MUST
	// set this to delegated and all other methods MUST set it to notDelegated.
}

type callType uint8

const (
	call callType = iota + 1
	callCode
	delegateCall
	staticCall
)

// run runs the [PrecompiledContract], differentiating between stateful and
// regular types.
func (args *evmCallArgs) run(p PrecompiledContract, input []byte, suppliedGas uint64) (ret []byte, remainingGas uint64, err error) {
	if p, ok := p.(statefulPrecompile); ok {
		return p(args, input, suppliedGas)
	}
	// Gas consumption for regular precompiles was already handled by the native
	// RunPrecompiledContract(), which called this method.
	ret, err = p.Run(input)
	return ret, suppliedGas, err
}

// PrecompiledStatefulContract is the stateful equivalent of a
// [PrecompiledContract].
type PrecompiledStatefulContract func(env PrecompileEnvironment, input []byte, suppliedGas uint64) (ret []byte, remainingGas uint64, err error)

// NewStatefulPrecompile constructs a new PrecompiledContract that can be used
// via an [EVM] instance but MUST NOT be called directly; a direct call to Run()
// reserves the right to panic. See other requirements defined in the comments
// on [PrecompiledContract].
func NewStatefulPrecompile(run PrecompiledStatefulContract) PrecompiledContract {
	return statefulPrecompile(run)
}

// statefulPrecompile implements the [PrecompiledContract] interface to allow a
// [PrecompiledStatefulContract] to be carried with regular geth plumbing. The
// methods are defined on this unexported type instead of directly on
// [PrecompiledStatefulContract] to hide implementation details.
type statefulPrecompile PrecompiledStatefulContract

// RequiredGas always returns zero as this gas is consumed by native geth code
// before the contract is run.
func (statefulPrecompile) RequiredGas([]byte) uint64 { return 0 }

func (p statefulPrecompile) Run([]byte) ([]byte, error) {
	// https://google.github.io/styleguide/go/best-practices.html#when-to-panic
	// This would indicate an API misuse and would occur in tests, not in
	// production.
	panic(fmt.Sprintf("BUG: call to %T.Run(); MUST call %T itself", p, p))
}

// A PrecompileEnvironment provides information about the context in which a
// precompiled contract is being run.
type PrecompileEnvironment interface {
	ChainConfig() *params.ChainConfig
	Rules() params.Rules
	ReadOnly() bool
	// StateDB will be non-nil i.f.f !ReadOnly().
	StateDB() StateDB
	// ReadOnlyState will always be non-nil.
	ReadOnlyState() libevm.StateReader
	Addresses() *libevm.AddressContext

	BlockHeader() (types.Header, error)
	BlockNumber() *big.Int
	BlockTime() uint64

	// Call is equivalent to [EVM.Call], with the caller defaulting to the
	// precompile receiving the environment, or to its own caller if invoked via
	// a delegated call.
	Call(addr common.Address, input []byte, gas uint64, value *uint256.Int, _ ...CallOption) (ret []byte, gasRemaining uint64, _ error)
}

var _ PrecompileEnvironment = (*evmCallArgs)(nil)

func (args *evmCallArgs) ChainConfig() *params.ChainConfig { return args.evm.chainConfig }
func (args *evmCallArgs) Rules() params.Rules              { return args.evm.chainRules }

func (args *evmCallArgs) ReadOnly() bool {
	// A switch statement provides clearer code coverage for difficult-to-test
	// cases.
	switch {
	case args.callType == staticCall:
		return true
	case args.evm.interpreter.readOnly:
		return true
	default:
		return false
	}
}

func (args *evmCallArgs) StateDB() StateDB {
	if args.ReadOnly() {
		return nil
	}
	return args.evm.StateDB
}

func (args *evmCallArgs) ReadOnlyState() libevm.StateReader {
	// Even though we're actually returning a full state database, the user
	// would have to actively circumvent the returned interface to use it. At
	// that point they're off-piste and it's not our problem.
	return args.evm.StateDB
}

func (args *evmCallArgs) self() common.Address { return args.addr }

func (args *evmCallArgs) Addresses() *libevm.AddressContext {
	return &libevm.AddressContext{
		Origin: args.evm.TxContext.Origin,
		Caller: args.caller.Address(),
		Self:   args.self(),
	}
}

func (args *evmCallArgs) BlockHeader() (types.Header, error) {
	hdr := args.evm.Context.Header
	if hdr == nil {
		// Although [core.NewEVMBlockContext] sets the field and is in the
		// typical hot path (e.g. miner), there are other ways to create a
		// [vm.BlockContext] (e.g. directly in tests) that may result in no
		// available header.
		return types.Header{}, fmt.Errorf("nil %T in current %T", hdr, args.evm.Context)
	}
	return *hdr, nil
}

func (args *evmCallArgs) BlockNumber() *big.Int {
	return new(big.Int).Set(args.evm.Context.BlockNumber)
}

func (args *evmCallArgs) BlockTime() uint64 { return args.evm.Context.Time }

func (args *evmCallArgs) Call(addr common.Address, input []byte, gas uint64, value *uint256.Int, opts ...CallOption) ([]byte, uint64, error) {
	in := args.evm.interpreter

	// The precompile run didn't increment the depth so this is necessary even
	// though Call() will eventually result in it being done again.
	in.evm.depth++
	defer func() { in.evm.depth-- }()

	if args.callType == staticCall && !in.readOnly {
		in.readOnly = true
		defer func() { in.readOnly = false }()
	}

	// This is equivalent to the `contract` variables created by evm.*Call*()
	// methods to pass to [EVMInterpreter.Run], which are then propagated by the
	// *CALL* opcodes as the caller.
	precompile := NewContract(args.caller, AccountRef(args.self()), args.value, args.gas)
	if args.callType == delegateCall {
		precompile = precompile.AsDelegate()
	}
	var caller ContractRef = precompile

	for _, o := range opts {
		switch o := o.(type) {
		case callOptUNSAFECallerAddressProxy:
			// Note that, in addition to being unsafe, this breaks an EVM
			// assumption that the caller ContractRef is always a *Contract.
			caller = AccountRef(args.caller.Address())
		case nil:
		default:
			return nil, gas, fmt.Errorf("unsupported option %T", o)
		}
	}

	return args.evm.Call(caller, addr, input, gas, value)
}

var (
	// These lock in the assumptions made when implementing [evmCallArgs]. If
	// these break then the struct fields SHOULD be changed to match these
	// signatures.
	_ = [](func(ContractRef, common.Address, []byte, uint64, *uint256.Int) ([]byte, uint64, error)){
		(*EVM)(nil).Call,
		(*EVM)(nil).CallCode,
	}
	_ = [](func(ContractRef, common.Address, []byte, uint64) ([]byte, uint64, error)){
		(*EVM)(nil).DelegateCall,
		(*EVM)(nil).StaticCall,
	}
)
