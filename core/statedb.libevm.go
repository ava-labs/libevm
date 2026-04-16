package core

// import (
// 	"github.com/holiman/uint256"

// 	"github.com/ava-labs/libevm/common"
// 	"github.com/ava-labs/libevm/core/types"
// 	"github.com/ava-labs/libevm/libevm"
// 	"github.com/ava-labs/libevm/libevm/stateconf"
// 	"github.com/ava-labs/libevm/params"
// )

// type StateDB interface {
// 	libevm.StateReader

// 	CreateAccount(common.Address)

// 	SetBalance(common.Address, *uint256.Int)
// 	SubBalance(common.Address, *uint256.Int)
// 	AddBalance(common.Address, *uint256.Int)

// 	SetNonce(common.Address, uint64)

// 	SetCode(common.Address, []byte)

// 	AddRefund(uint64)
// 	SubRefund(uint64)

// 	SetState(common.Address, common.Hash, common.Hash, ...stateconf.StateDBStateOption)

// 	SetTransientState(addr common.Address, key, value common.Hash)

// 	SelfDestruct(common.Address)
// 	Selfdestruct6780(common.Address)

// 	AddAddressToAccessList(addr common.Address)
// 	AddSlotToAccessList(addr common.Address, slot common.Hash)
// 	Prepare(rules params.Rules, sender, coinbase common.Address, dest *common.Address, precompiles []common.Address, txAccesses types.AccessList)

// 	RevertToSnapshot(int)
// 	Snapshot() int

// 	AddLog(*types.Log)
// 	AddPreimage(common.Hash, []byte)

// 	Finalise(bool)
// 	IntermediateRoot(bool) common.Hash

// 	GetLogs(common.Hash, uint64, common.Hash) []*types.Log
// 	TxIndex() int
// }
