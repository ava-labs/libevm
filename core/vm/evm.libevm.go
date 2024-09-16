package vm

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/libevm"
)

func (evm *EVM) canCreateContract(
	caller ContractRef,
	deployedCode []byte,
	address common.Address,
	gas uint64,
	err error,
) ([]byte, common.Address, uint64, error) {
	if err == nil && address != (common.Address{}) { // NOTE `err ==` not !=
		addrs := &libevm.AddressContext{Origin: evm.Origin, Caller: caller.Address(), Self: address}
		gas, err = evm.chainRules.Hooks().CanCreateContract(addrs, gas, evm.StateDB)
	}
	return deployedCode, address, gas, err
}
