// Copyright 2026 the libevm authors.
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

package tracers

import (
	"errors"
	"math/big"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/libevm/common/hexutil"
	"github.com/ava-labs/libevm/core"
	"github.com/ava-labs/libevm/core/state"
	"github.com/ava-labs/libevm/core/types"
	"github.com/ava-labs/libevm/internal/ethapi"
	"github.com/ava-labs/libevm/params"
	"github.com/ava-labs/libevm/rpc"
)

// extraExecutorBackend wraps a [testBackend] with [ExtraExecutor] hooks that
// count their invocations and return the configured errors.
type extraExecutorBackend struct {
	*testBackend

	t   *testing.T
	api *API

	account  Account
	txHashes [][]common.Hash // [block-1][txIndex] of the generated blocks

	errBeforeBlock, errAfterTx     error
	beforeBlockCalls, afterTxCalls atomic.Int64
}

var _ ExtraExecutor = (*extraExecutorBackend)(nil)

const (
	extraExecGenBlocks   = 3
	extraExecTxsPerBlock = 2
)

func newExtraExecutorSUT(t *testing.T, errBeforeBlock, errAfterTx error) *extraExecutorBackend {
	t.Helper()

	account := newAccounts(1)[0]
	genesis := &core.Genesis{
		Config: params.TestChainConfig,
		Alloc: types.GenesisAlloc{
			account.addr: {Balance: big.NewInt(params.Ether)},
		},
	}
	signer := types.HomesteadSigner{}

	txHashes := make([][]common.Hash, extraExecGenBlocks)
	backend := newTestBackend(t, extraExecGenBlocks, genesis, func(i int, b *core.BlockGen) {
		for j := range extraExecTxsPerBlock {
			tx, err := types.SignTx(types.NewTx(&types.LegacyTx{
				Nonce:    uint64(i*extraExecTxsPerBlock + j), // #nosec G115 -- known non-negative
				To:       &common.Address{},
				Value:    big.NewInt(1000),
				Gas:      params.TxGas,
				GasPrice: b.BaseFee(),
			}), signer, account.key)
			require.NoError(t, err, "types.SignTx()")
			b.AddTx(tx)
			txHashes[i] = append(txHashes[i], tx.Hash())
		}
	})

	sut := &extraExecutorBackend{
		testBackend:    backend,
		t:              t,
		account:        account,
		txHashes:       txHashes,
		errBeforeBlock: errBeforeBlock,
		errAfterTx:     errAfterTx,
	}
	sut.api = NewAPI(sut)
	return sut
}

func (b *extraExecutorBackend) BeforeExecutingBlock(sdb *state.StateDB, parent, header *types.Header) error {
	b.beforeBlockCalls.Add(1)
	assert.Equalf(b.t, parent.Number.Uint64()+1, header.Number.Uint64(), "BeforeExecutingBlock() header.Number, given parent.Number %d", parent.Number.Uint64())
	return b.errBeforeBlock
}

func (b *extraExecutorBackend) AfterExecutingTransaction(sdb *state.StateDB, baseFee *big.Int, gasUsed uint64) error {
	b.afterTxCalls.Add(1)
	assert.Equal(b.t, params.TxGas, gasUsed, "AfterExecutingTransaction() gasUsed of plain transfer")
	assert.NotNil(b.t, baseFee, "AfterExecutingTransaction() baseFee")
	return b.errAfterTx
}

func (b *extraExecutorBackend) assertCalls(t *testing.T, wantBeforeBlock, wantAfterTx int64) {
	t.Helper()
	assert.Equal(t, wantBeforeBlock, b.beforeBlockCalls.Load(), "number of calls to BeforeExecutingBlock()")
	assert.Equal(t, wantAfterTx, b.afterTxCalls.Load(), "number of calls to AfterExecutingTransaction()")
}

var (
	errBeforeBlock = errors.New("BeforeExecutingBlock error")
	errAfterTx     = errors.New("AfterExecutingTransaction error")
)

func TestExtraExecutorTraceBlockByNumber(t *testing.T) {
	tests := []struct {
		name                       string
		errBeforeBlock, errAfterTx error
		wantErr                    error
		wantAfterTxCalls           int64
	}{
		{
			name:             "no_errors",
			wantAfterTxCalls: extraExecTxsPerBlock,
		},
		{
			name:             "BeforeExecutingBlock_error",
			errBeforeBlock:   errBeforeBlock,
			wantErr:          errBeforeBlock,
			wantAfterTxCalls: 0,
		},
		{
			name:             "AfterExecutingTransaction_error",
			errAfterTx:       errAfterTx,
			wantErr:          errAfterTx,
			wantAfterTxCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sut := newExtraExecutorSUT(t, tt.errBeforeBlock, tt.errAfterTx)
			defer sut.teardown()

			results, err := sut.api.TraceBlockByNumber(t.Context(), rpc.LatestBlockNumber, nil)
			require.ErrorIsf(t, err, tt.wantErr, "%T.TraceBlockByNumber() error", sut.api)
			sut.assertCalls(t, 1, tt.wantAfterTxCalls)
			if tt.wantErr != nil {
				return
			}
			require.Len(t, results, extraExecTxsPerBlock, "number of traced transactions")
		})
	}
}

func TestExtraExecutorIntermediateRoots(t *testing.T) {
	tests := []struct {
		name                       string
		errBeforeBlock, errAfterTx error
		wantErr                    error
		wantAfterTxCalls           int64
	}{
		{
			name:             "no_errors",
			wantAfterTxCalls: extraExecTxsPerBlock,
		},
		{
			name:           "BeforeExecutingBlock_error",
			errBeforeBlock: errBeforeBlock,
			wantErr:        errBeforeBlock,
		},
		{
			name:       "AfterExecutingTransaction_error",
			errAfterTx: errAfterTx,
			// IntermediateRoots intentionally swallows transaction errors and
			// returns the roots accumulated up to that point.
			wantErr:          nil,
			wantAfterTxCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sut := newExtraExecutorSUT(t, tt.errBeforeBlock, tt.errAfterTx)
			defer sut.teardown()

			head := sut.chain.CurrentHeader()
			_, err := sut.api.IntermediateRoots(t.Context(), head.Hash(), nil)
			require.ErrorIsf(t, err, tt.wantErr, "%T.IntermediateRoots() error", sut.api)
			sut.assertCalls(t, 1, tt.wantAfterTxCalls)
		})
	}
}

func TestExtraExecutorTraceChain(t *testing.T) {
	sut := newExtraExecutorSUT(t, nil, nil)
	defer sut.teardown()

	// TraceChain is only available via subscription, which requires an actual
	// RPC connection.
	server := rpc.NewServer()
	defer server.Stop()
	require.NoError(t, server.RegisterName("debug", sut.api), "rpc.Server.RegisterName()")
	client := rpc.DialInProc(server)
	defer client.Close()

	ch := make(chan *blockTraceResult)
	sub, err := client.Subscribe(t.Context(), "debug", ch, "traceChain", rpc.BlockNumber(0), rpc.BlockNumber(extraExecGenBlocks), nil)
	require.NoError(t, err, `rpc.Client.Subscribe("traceChain")`)
	defer sub.Unsubscribe()

	for i := 1; i <= extraExecGenBlocks; i++ {
		select {
		case res := <-ch:
			assert.Equalf(t, hexutil.Uint64(i), res.Block, "block number of result %d", i)
			assert.Lenf(t, res.Traces, extraExecTxsPerBlock, "number of traced transactions in block %d", i)
		case err := <-sub.Err():
			t.Fatalf("subscription error before receiving all results: %v", err)
		}
	}
	sut.assertCalls(t, extraExecGenBlocks, extraExecGenBlocks*extraExecTxsPerBlock)
}

// The state given to TraceTransaction and TraceCall is sourced at the
// transaction, without re-executing the block, so BeforeExecutingBlock is
// never called (its error MUST NOT be reported) and AfterExecutingTransaction
// runs after the result has been captured (only its error is observable).

func TestExtraExecutorTraceTransaction(t *testing.T) {
	tests := []struct {
		name                       string
		errBeforeBlock, errAfterTx error
		wantErr                    error
	}{
		{
			name: "no_errors",
		},
		{
			name:           "BeforeExecutingBlock_error_not_reported",
			errBeforeBlock: errBeforeBlock,
			wantErr:        nil,
		},
		{
			name:       "AfterExecutingTransaction_error",
			errAfterTx: errAfterTx,
			wantErr:    errAfterTx,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sut := newExtraExecutorSUT(t, tt.errBeforeBlock, tt.errAfterTx)
			defer sut.teardown()

			_, err := sut.api.TraceTransaction(t.Context(), sut.txHashes[extraExecGenBlocks-1][0], nil)
			require.ErrorIsf(t, err, tt.wantErr, "%T.TraceTransaction() error", sut.api)
			sut.assertCalls(t, 0, 1)
		})
	}
}

func TestExtraExecutorTraceCall(t *testing.T) {
	tests := []struct {
		name                       string
		errBeforeBlock, errAfterTx error
		wantErr                    error
	}{
		{
			name: "no_errors",
		},
		{
			name:           "BeforeExecutingBlock_error_not_reported",
			errBeforeBlock: errBeforeBlock,
			wantErr:        nil,
		},
		{
			name:       "AfterExecutingTransaction_error",
			errAfterTx: errAfterTx,
			wantErr:    errAfterTx,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sut := newExtraExecutorSUT(t, tt.errBeforeBlock, tt.errAfterTx)
			defer sut.teardown()

			args := ethapi.TransactionArgs{
				From: &sut.account.addr,
				To:   &common.Address{},
			}
			_, err := sut.api.TraceCall(t.Context(), args, rpc.BlockNumberOrHashWithNumber(rpc.LatestBlockNumber), nil)
			require.ErrorIsf(t, err, tt.wantErr, "%T.TraceCall() error", sut.api)
			sut.assertCalls(t, 0, 1)
		})
	}
}
