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
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/libevm/core"
	"github.com/ava-labs/libevm/core/types"
	"github.com/ava-labs/libevm/core/vm"
	"github.com/ava-labs/libevm/crypto"
	"github.com/ava-labs/libevm/params"
	"github.com/ava-labs/libevm/rpc"
)

const (
	// blockHashCaptureTracerName is registered as a native tracer, exercising
	// the sequential path of [API.traceBlock].
	blockHashCaptureTracerName = "libevmBlockHashCaptureTracer"
	// blockHashCaptureTracerJSName is the same tracer registered with
	// isJS=true, forcing [API.traceBlock] onto the [API.traceBlockParallel]
	// path.
	blockHashCaptureTracerJSName = "libevmBlockHashCaptureTracerJS"
)

func TestMain(m *testing.M) {
	DefaultDirectory.Register(blockHashCaptureTracerName, newBlockHashCaptureTracer, false)
	DefaultDirectory.Register(blockHashCaptureTracerJSName, newBlockHashCaptureTracer, true)
	os.Exit(m.Run())
}

// blockHashCaptureTracer records the [Context.BlockHash] it was constructed
// with and returns it as its result, making the hash propagated by the API
// observable from trace results.
type blockHashCaptureTracer struct {
	blockHash common.Hash
}

func newBlockHashCaptureTracer(ctx *Context, _ json.RawMessage) (Tracer, error) {
	return &blockHashCaptureTracer{blockHash: ctx.BlockHash}, nil
}

func (t *blockHashCaptureTracer) GetResult() (json.RawMessage, error) {
	return json.Marshal(t.blockHash)
}

func (*blockHashCaptureTracer) Stop(error)            {}
func (*blockHashCaptureTracer) CaptureTxStart(uint64) {}
func (*blockHashCaptureTracer) CaptureTxEnd(uint64)   {}
func (*blockHashCaptureTracer) CaptureStart(*vm.EVM, common.Address, common.Address, bool, []byte, uint64, *big.Int) {
}
func (*blockHashCaptureTracer) CaptureEnd([]byte, uint64, error) {}
func (*blockHashCaptureTracer) CaptureEnter(vm.OpCode, common.Address, common.Address, []byte, uint64, *big.Int) {
}
func (*blockHashCaptureTracer) CaptureExit([]byte, uint64, error) {}
func (*blockHashCaptureTracer) CaptureState(uint64, vm.OpCode, uint64, uint64, *vm.ScopeContext, []byte, int, error) {
}
func (*blockHashCaptureTracer) CaptureFault(uint64, vm.OpCode, uint64, uint64, *vm.ScopeContext, int, error) {
}

// overriddenBlockHash returns the hash reported by
// [blockHashOverrideBackend.BlockHash] for the block at height `num`. It is
// deliberately different from any real block hash.
func overriddenBlockHash(num uint64) common.Hash {
	bytes := binary.BigEndian.AppendUint64(nil, num)
	return crypto.Keccak256Hash(bytes)
}

// blockHashOverrideBackend wraps a [testBackend], implementing
// [BlockHashOverrider] such that the "correct" hash of every block differs
// from block.Hash(). As a real overriding backend would, it also indexes
// transactions and resolves blocks by the overridden hashes.
type blockHashOverrideBackend struct {
	*testBackend
}

var _ BlockHashOverrider = (*blockHashOverrideBackend)(nil)

func (b *blockHashOverrideBackend) BlockHash(block *types.Block) common.Hash {
	return overriddenBlockHash(block.NumberU64())
}

func (b *blockHashOverrideBackend) GetTransaction(ctx context.Context, txHash common.Hash) (bool, *types.Transaction, common.Hash, uint64, uint64, error) {
	found, tx, _, blockNumber, index, err := b.testBackend.GetTransaction(ctx, txHash)
	return found, tx, overriddenBlockHash(blockNumber), blockNumber, index, err
}

func (b *blockHashOverrideBackend) BlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error) {
	for n := uint64(0); n <= b.chain.CurrentBlock().Number.Uint64(); n++ {
		if overriddenBlockHash(n) == hash {
			return b.chain.GetBlockByNumber(n), nil
		}
	}
	return b.testBackend.BlockByHash(ctx, hash)
}

// newBlockHashOverrideBackend returns a [blockHashOverrideBackend] over a
// chain of `n` blocks, each containing a single value-transfer transaction.
func newBlockHashOverrideBackend(t *testing.T, n int) *blockHashOverrideBackend {
	t.Helper()

	accounts := newAccounts(2)
	genesis := &core.Genesis{
		Config: params.TestChainConfig,
		Alloc: types.GenesisAlloc{
			accounts[0].addr: {Balance: big.NewInt(params.Ether)},
		},
	}
	signer := types.HomesteadSigner{}
	nonce := uint64(0)
	backend := newTestBackend(t, n, genesis, func(i int, b *core.BlockGen) {
		tx, err := types.SignTx(types.NewTransaction(nonce, accounts[1].addr, big.NewInt(1000), params.TxGas, b.BaseFee(), nil), signer, accounts[0].key)
		require.NoError(t, err, "types.SignTx()")
		b.AddTx(tx)
		nonce++
	})
	t.Cleanup(backend.teardown)
	return &blockHashOverrideBackend{testBackend: backend}
}

// hashFromTraceResult extracts the [common.Hash] reported by a
// [blockHashCaptureTracer] from a trace result.
func hashFromTraceResult(t *testing.T, res interface{}) common.Hash {
	t.Helper()

	raw, err := json.Marshal(res)
	require.NoError(t, err, "json.Marshal() trace result")
	var h common.Hash
	require.NoError(t, json.Unmarshal(raw, &h), "json.Unmarshal() trace result into common.Hash")
	return h
}

// TestOverrideBlockHash_TraceBlock tests that the [Context.BlockHash]
// received by tracers is the overridden hash, on both the sequential
// ([API.traceBlock]) and parallel ([API.traceBlockParallel]) paths.
func TestOverrideBlockHash_TraceBlock(t *testing.T) {
	const blockNum = 1
	backend := newBlockHashOverrideBackend(t, blockNum)
	api := NewAPI(backend)

	for _, tracer := range []string{blockHashCaptureTracerName, blockHashCaptureTracerJSName} {
		t.Run(tracer, func(t *testing.T) {
			results, err := api.TraceBlockByNumber(t.Context(), rpc.BlockNumber(blockNum), &TraceConfig{
				Tracer: &tracer,
			})
			require.NoErrorf(t, err, "TraceBlockByNumber(%d)", blockNum)
			require.NotEmpty(t, results, "trace results")

			for i, res := range results {
				require.Emptyf(t, res.Error, "trace error for tx %d", i)
				require.Equalf(t, overriddenBlockHash(blockNum), hashFromTraceResult(t, res.Result),
					"Context.BlockHash received by tracer for tx %d of block %d", i, blockNum)
			}
		})
	}
}

// TestOverrideBlockHash_TraceChain tests that both the block hash
// reported in [blockTraceResult] and the [Context.BlockHash] received by
// tracers during [API.traceChain] are the overridden hashes.
func TestOverrideBlockHash_TraceChain(t *testing.T) {
	const numBlocks = 5
	backend := newBlockHashOverrideBackend(t, numBlocks)
	api := NewAPI(backend)

	from, err := api.blockByNumber(t.Context(), 0)
	require.NoError(t, err, "blockByNumber(0)")
	to, err := api.blockByNumber(t.Context(), numBlocks)
	require.NoErrorf(t, err, "blockByNumber(%d)", numBlocks)

	tracer := blockHashCaptureTracerName
	resCh := api.traceChain(from, to, &TraceConfig{Tracer: &tracer}, nil)

	next := uint64(1)
	for result := range resCh {
		require.Equal(t, next, uint64(result.Block), "traced block number")
		require.Equalf(t, overriddenBlockHash(next), result.Hash, "blockTraceResult.Hash of block %d", next)

		require.NotEmpty(t, result.Traces, "trace results of block %d", next)
		for i, trace := range result.Traces {
			require.Emptyf(t, trace.Error, "trace error for tx %d of block %d", i, next)
			require.Equalf(t, overriddenBlockHash(next), hashFromTraceResult(t, trace.Result),
				"Context.BlockHash received by tracer for tx %d of block %d", i, next)
		}
		next++
	}
	require.Equal(t, uint64(numBlocks+1), next, "all blocks traced")
}

// TestOverrideBlockHash_TraceTransaction tests that a transaction found
// via [Backend.GetTransaction] under an overridden block hash can be traced,
// and that the tracer receives that hash as [Context.BlockHash].
func TestOverrideBlockHash_TraceTransaction(t *testing.T) {
	const blockNum = 1
	backend := newBlockHashOverrideBackend(t, blockNum)
	api := NewAPI(backend)

	block, err := backend.BlockByNumber(t.Context(), blockNum)
	require.NoErrorf(t, err, "BlockByNumber(%d)", blockNum)
	require.NotEmptyf(t, block.Transactions(), "%T.Transactions()", block)
	txHash := block.Transactions()[0].Hash()

	tracer := blockHashCaptureTracerName
	res, err := api.TraceTransaction(t.Context(), txHash, &TraceConfig{Tracer: &tracer})
	require.NoErrorf(t, err, "TraceTransaction()")
	require.Equal(t, overriddenBlockHash(blockNum), hashFromTraceResult(t, res), "Context.BlockHash received by tracer")
}

// TestOverrideBlockHash_StandardTraceBlockToFile tests that the dump
// files created by [API.StandardTraceBlockToFile] are named after the
// overridden block hash rather than block.Hash().
func TestOverrideBlockHash_StandardTraceBlockToFile(t *testing.T) {
	const blockNum = 1
	backend := newBlockHashOverrideBackend(t, blockNum)
	api := NewAPI(backend)

	block, err := backend.BlockByNumber(t.Context(), blockNum)
	require.NoErrorf(t, err, "BlockByNumber(%d)", blockNum)

	files, err := api.StandardTraceBlockToFile(t.Context(), block.Hash(), nil)
	for _, file := range files {
		defer os.Remove(file)
	}
	require.NoError(t, err, "StandardTraceBlockToFile()")
	require.NotEmpty(t, files, "dump files")

	wantPrefix := fmt.Sprintf("block_%#x-", overriddenBlockHash(blockNum).Bytes()[:4])
	for _, file := range files {
		require.Containsf(t, file, wantPrefix, "dump file %q named after overridden block hash", file)
	}
}
