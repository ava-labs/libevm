package ethtest

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"testing"

	ethereum "github.com/ava-labs/libevm"
	"github.com/ava-labs/libevm/accounts/abi/bind"
	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/libevm/core"
	"github.com/ava-labs/libevm/core/state"
	"github.com/ava-labs/libevm/core/types"
	"github.com/ava-labs/libevm/core/vm"
)

func NewMinimalBackend(tb testing.TB, opts ...EVMOption) (*MinimalBackend, types.Signer) {
	tb.Helper()

	sdb, evm := NewZeroEVM(tb, opts...)
	signer := types.LatestSigner(evm.ChainConfig())

	return &MinimalBackend{
		evm:     evm,
		sdb:     sdb,
		signer:  signer,
		results: make(map[common.Hash]*core.ExecutionResult),
	}, signer
}

var _ bind.ContractBackend = (*MinimalBackend)(nil)

// A MinimalBackend provides the smallest subset of methods from
// [bind.ContractBackend] to allow for testing of smart contracts via `abigen`
// bindings. It only supports the tx-sending and read-only calling methods of
// the bindings, with the [MinimalBackend.ResultOf] method available in lieu of
// receipts.
type MinimalBackend struct {
	evm    *vm.EVM
	sdb    *state.StateDB
	signer types.Signer

	txIndex int
	results map[common.Hash]*core.ExecutionResult

	// ContractBackend is embedded to allow use in `abigen` bindings, but not
	// all methods are implemented so unsupported behaviour will panic.
	bind.ContractBackend
}

// CallContract returns the results of [vm.EVM.StaticCall], called with
// arguments sourced from `msg`. The block number is ignored, and
// 100e6 gas is always used.
func (b *MinimalBackend) CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	if msg.To == nil {
		return nil, errors.New("calling nil contract address")
	}
	ret, _, err := b.evm.StaticCall(
		vm.AccountRef(msg.From),
		*msg.To,
		msg.Data,
		100e6,
	)
	return ret, err
}

// HeaderByNumber returns an entirely empty, but non-nil header.
func (b *MinimalBackend) HeaderByNumber(ctx context.Context, num *big.Int) (*types.Header, error) {
	return &types.Header{}, nil
}

// SuggestGasPrice returns 1.
func (b *MinimalBackend) SuggestGasPrice(context.Context) (*big.Int, error) {
	return big.NewInt(1), nil
}

// PendingCodeAt returns the current code as available from the underlying
// [state.StateDB].
func (b *MinimalBackend) PendingCodeAt(ctx context.Context, addr common.Address) ([]byte, error) {
	return b.evm.StateDB.GetCode(addr), nil
}

// PendingNonceAt returns the current nonce as available from the underlying
// [state.StateDB].
func (b *MinimalBackend) PendingNonceAt(ctx context.Context, addr common.Address) (uint64, error) {
	return b.evm.StateDB.GetNonce(addr), nil
}

// EstimateGas returns 100e6.
func (b *MinimalBackend) EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error) {
	return 100e6, nil
}

func (b *MinimalBackend) SendTransaction(ctx context.Context, tx *types.Transaction) (retErr error) {
	b.sdb.SetTxContext(tx.Hash(), b.txIndex)
	defer func() {
		if retErr == nil {
			b.txIndex++
		}
	}()

	msg, err := core.TransactionToMessage(tx, b.signer, big.NewInt(1))
	if err != nil {
		return err
	}
	b.evm.Reset(core.NewEVMTxContext(msg), b.sdb)

	gp := core.GasPool(tx.Gas())
	res, err := core.ApplyMessage(b.evm, msg, &gp)
	if err != nil {
		return err
	}
	b.results[tx.Hash()] = res

	if res.Err != nil {
		return fmt.Errorf("%w: %s", res.Err, res.ReturnData)
	}
	return nil
}

type ExecutionResult struct {
	*core.ExecutionResult
	Logs []*types.Log
}

func (b *MinimalBackend) ResultOf(tb testing.TB) func(tx *types.Transaction, err error) *ExecutionResult {
	tb.Helper()
	var used bool

	return func(tx *types.Transaction, err error) *ExecutionResult {
		tb.Helper()
		if used {
			tb.Fatal()
		}
		used = true

		if err != nil {
			tb.Fatal(err)
		}

		h := tx.Hash()
		res, ok := b.results[h]
		if !ok {
			tb.Fatalf("Transaction %#x not found", h)
		}
		return &ExecutionResult{
			ExecutionResult: res,
			Logs:            b.sdb.GetLogs(h, 0, common.Hash{}),
		}
	}
}
