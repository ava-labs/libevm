package vm_test

import (
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/libevm"
	"github.com/ethereum/go-ethereum/libevm/ethtest"
	"github.com/ethereum/go-ethereum/libevm/hookstest"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/require"
)

func TestCanCreateContract(t *testing.T) {
	rng := ethtest.NewPseudoRand(142857)
	account := rng.Address()
	slot := rng.Hash()

	makeErr := func(cc *libevm.AddressContext, stateVal common.Hash) error {
		return fmt.Errorf("Origin: %v Caller: %v Contract: %v State: %v", cc.Origin, cc.Caller, cc.Self, stateVal)
	}
	hooks := &hookstest.Stub{
		CanCreateContractFn: func(cc *libevm.AddressContext, s libevm.StateReader) error {
			return makeErr(cc, s.GetState(account, slot))
		},
	}
	hooks.Register(t)

	origin := rng.Address()
	caller := rng.Address()
	value := rng.Hash()
	code := rng.Bytes(8)
	salt := rng.Hash()

	create := crypto.CreateAddress(caller, 0)
	create2 := crypto.CreateAddress2(caller, salt, crypto.Keccak256(code))

	tests := []struct {
		name    string
		create  func(*vm.EVM) ([]byte, common.Address, uint64, error)
		wantErr error
	}{
		{
			name: "Create",
			create: func(evm *vm.EVM) ([]byte, common.Address, uint64, error) {
				return evm.Create(vm.AccountRef(caller), code, 1e6, uint256.NewInt(0))
			},
			wantErr: makeErr(&libevm.AddressContext{Origin: origin, Caller: caller, Self: create}, value),
		},
		{
			name: "Create2",
			create: func(evm *vm.EVM) ([]byte, common.Address, uint64, error) {
				return evm.Create2(vm.AccountRef(caller), code, 1e6, uint256.NewInt(0), new(uint256.Int).SetBytes(salt[:]))
			},
			wantErr: makeErr(&libevm.AddressContext{Origin: origin, Caller: caller, Self: create2}, value),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state, evm := ethtest.NewZeroEVM(t)
			state.SetState(account, slot, value)
			evm.TxContext.Origin = origin

			_, _, _, err := tt.create(evm)
			require.EqualError(t, err, tt.wantErr.Error())
		})
	}
}
