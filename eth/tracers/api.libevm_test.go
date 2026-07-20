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
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/libevm/core"
	"github.com/ava-labs/libevm/core/state"
	"github.com/ava-labs/libevm/core/types"
	"github.com/ava-labs/libevm/core/vm"
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
	DefaultDirectory.Register(blockHashCaptureTracerName, newBlockHashCaptureTracer, false /*isJS*/)
	DefaultDirectory.Register(blockHashCaptureTracerJSName, newBlockHashCaptureTracer, true /*isJS*/)
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

// alteredBackend models a chain that uses a different design for their blocks,
// likely due to an asynchronous execution model.
// The state root in the header is ALWAYS altered before returning to the API,
// resulting in a different block hash than the canonical one.
type alteredBackend struct {
	*testBackend
}

var _ BlockHashOverrider = (*alteredBackend)(nil)

func (b *alteredBackend) BlockHash(block *types.Block) common.Hash {
	return b.canonicalHash(block.NumberU64())
}

func (b *alteredBackend) BlockByNumber(ctx context.Context, number rpc.BlockNumber) (*types.Block, error) {
	block, err := b.testBackend.BlockByNumber(ctx, number)
	return alterBlock(block), err
}

func (b *alteredBackend) BlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error) {
	return alterBlock(b.chain.GetBlockByHash(hash)), nil
}

func (b *alteredBackend) HeaderByNumber(ctx context.Context, number rpc.BlockNumber) (*types.Header, error) {
	header, err := b.testBackend.HeaderByNumber(ctx, number)
	return alterHeader(header), err
}

func (b *alteredBackend) HeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error) {
	header, err := b.testBackend.HeaderByHash(ctx, hash)
	return alterHeader(header), err
}

func (b *alteredBackend) StateAtBlock(ctx context.Context, block *types.Block, reexec uint64, base *state.StateDB, readOnly, preferDisk bool) (*state.StateDB, StateReleaseFunc, error) {
	return b.testBackend.StateAtBlock(ctx, b.canonicalBlock(block.NumberU64()), reexec, base, readOnly, preferDisk)
}

func (b *alteredBackend) StateAtTransaction(ctx context.Context, block *types.Block, txIndex int, reexec uint64) (*core.Message, vm.BlockContext, *state.StateDB, StateReleaseFunc, error) {
	return b.testBackend.StateAtTransaction(ctx, b.canonicalBlock(block.NumberU64()), txIndex, reexec)
}

func (b *alteredBackend) canonicalBlock(num uint64) *types.Block {
	return b.chain.GetBlockByNumber(num)
}

func (b *alteredBackend) canonicalHash(num uint64) common.Hash {
	return b.chain.GetCanonicalHash(num)
}

// alterHeader simulates post-consensus population. ParentHash is preserved and
// stays canonical.
func alterHeader(header *types.Header) *types.Header {
	if header == nil {
		return nil
	}
	altered := types.CopyHeader(header)
	altered.Root = common.Hash{'a', 'l', 't', 'e', 'r', 'e', 'd'}
	return altered
}

func alterBlock(block *types.Block) *types.Block {
	if block == nil {
		return nil
	}
	return types.NewBlockWithHeader(alterHeader(block.Header())).WithBody(types.Body{
		Transactions: block.Transactions(),
		Uncles:       block.Uncles(),
	})
}

// numBlocks is the height of the chain [newAlteredBackend] build.
const numBlocks = 5

// newAlteredBackend returns a backend over [numBlocks] blocks of one value
// transfer each.
func newAlteredBackend(t *testing.T) *alteredBackend {
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
	backend := newTestBackend(t, numBlocks, genesis, func(i int, b *core.BlockGen) {
		tx, err := types.SignTx(types.NewTransaction(nonce, accounts[1].addr, big.NewInt(1000), params.TxGas, b.BaseFee(), nil), signer, accounts[0].key)
		require.NoError(t, err, "types.SignTx()")
		b.AddTx(tx)
		nonce++
	})
	t.Cleanup(backend.teardown)
	return &alteredBackend{testBackend: backend}
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

// traceBlockHashes returns the hashes [API.TraceBlockByNumber] surfaces, keyed by block.
func traceBlockHashes(t *testing.T, api *API, tracer string) map[uint64][]common.Hash {
	t.Helper()

	seen := make(map[uint64][]common.Hash)
	for n := uint64(1); n <= numBlocks; n++ {
		results, err := api.TraceBlockByNumber(t.Context(), rpc.BlockNumber(n), &TraceConfig{
			Tracer: &tracer,
		})
		require.NoErrorf(t, err, "TraceBlockByNumber(%d)", n)
		for i, res := range results {
			require.Emptyf(t, res.Error, "trace error for tx %d of block %d", i, n)
			seen[n] = append(seen[n], hashFromTraceResult(t, res.Result))
		}
	}
	return seen
}

// traceChainHashes returns the hashes [API.traceChain] surfaces, keyed by block.
func traceChainHashes(t *testing.T, api *API) map[uint64][]common.Hash {
	t.Helper()

	from, err := api.blockByNumber(t.Context(), 0)
	require.NoError(t, err, "blockByNumber(0)")
	to, err := api.blockByNumber(t.Context(), numBlocks)
	require.NoErrorf(t, err, "blockByNumber(%d)", numBlocks)

	tracer := blockHashCaptureTracerName
	seen := make(map[uint64][]common.Hash)

	next := uint64(1)
	for result := range api.traceChain(from, to, &TraceConfig{Tracer: &tracer}, nil) {
		require.Equal(t, next, uint64(result.Block), "traced block number")
		seen[next] = append(seen[next], result.Hash)
		for i, trace := range result.Traces {
			require.Emptyf(t, trace.Error, "trace error for tx %d of block %d", i, next)
			seen[next] = append(seen[next], hashFromTraceResult(t, trace.Result))
		}
		next++
	}
	require.Equal(t, uint64(numBlocks+1), next, "all blocks traced")
	return seen
}

// requireCanonicalHashes asserts every surfaced hash is canonical and no block is missing.
func requireCanonicalHashes(t *testing.T, backend *alteredBackend, seen map[uint64][]common.Hash) {
	t.Helper()

	require.Lenf(t, seen, numBlocks, "blocks with surfaced hashes")
	for n := uint64(1); n <= numBlocks; n++ {
		require.NotEmptyf(t, seen[n], "hashes surfaced for block %d", n)
		for _, got := range seen[n] {
			require.Equalf(t, backend.canonicalHash(n), got, "block hash surfaced for block %d", n)
		}
	}
}

// TestOverrideBlockHash_Propagation tests that the tracing APIs surface canonical hashes.
func TestOverrideBlockHash_Propagation(t *testing.T) {
	backend := newAlteredBackend(t)
	api := NewAPI(backend)

	t.Run("traceBlock_sequential", func(t *testing.T) {
		requireCanonicalHashes(t, backend, traceBlockHashes(t, api, blockHashCaptureTracerName))
	})

	t.Run("traceBlock_parallel", func(t *testing.T) {
		requireCanonicalHashes(t, backend, traceBlockHashes(t, api, blockHashCaptureTracerJSName))
	})

	t.Run("traceChain", func(t *testing.T) {
		requireCanonicalHashes(t, backend, traceChainHashes(t, api))
	})
}

func blockAt(t *testing.T, b *alteredBackend, num uint64) *types.Block {
	t.Helper()

	block, err := b.BlockByNumber(t.Context(), rpc.BlockNumber(num))
	require.NoErrorf(t, err, "BlockByNumber(%d)", num)
	return block
}

func headerAt(t *testing.T, b *alteredBackend, num uint64) *types.Header {
	t.Helper()

	header, err := b.HeaderByNumber(t.Context(), rpc.BlockNumber(num))
	require.NoErrorf(t, err, "HeaderByNumber(%d)", num)
	return header
}

// TestAlteredHashBackend_AccessorsAgree tests that the header and block accessors serve
// the same header, so header-only paths cannot see the unaltered form.
//
// This test failure indicates a bug in the test implementation.
func TestAlteredHashBackend_AccessorsAgree(t *testing.T) {
	backend := newAlteredBackend(t)

	for n := uint64(1); n <= numBlocks; n++ {
		block, header := blockAt(t, backend, n), headerAt(t, backend, n)

		byHash, err := backend.HeaderByHash(t.Context(), backend.canonicalHash(n))
		require.NoErrorf(t, err, "HeaderByHash(canonical %d)", n)

		require.Equalf(t, block.Hash(), header.Hash(), "HeaderByNumber() vs BlockByNumber() for block %d", n)
		require.Equalf(t, header.Hash(), byHash.Hash(), "HeaderByHash() vs HeaderByNumber() for block %d", n)
	}
}

// TestAlteredHashBackend_ParentHashIsCanonical tests that ParentHash stays canonical
// even though block.Hash() does not.
//
// This test failure indicates a bug in the test implementation.
func TestAlteredHashBackend_ParentHashIsCanonical(t *testing.T) {
	backend := newAlteredBackend(t)

	for n := uint64(1); n <= numBlocks; n++ {
		block, parent := blockAt(t, backend, n), blockAt(t, backend, n-1)

		require.NotEqualf(t, backend.canonicalHash(n), block.Hash(),
			"altering the header changes block.Hash() away from the canonical hash, block %d", n)
		require.Equalf(t, backend.canonicalHash(n-1), block.ParentHash(),
			"ParentHash of block %d is the canonical hash of its parent", n)
		require.NotEqualf(t, parent.Hash(), block.ParentHash(),
			"ParentHash of block %d is not the altered parent's block.Hash()", n)
	}
}

// TestBlockHashOverriderOptional tests that [API.blockHash] falls back to block.Hash() unless
// the [Backend] is a [BlockHashOverrider].
func TestBlockHashOverriderOptional(t *testing.T) {
	const blockNum = 1 // must be <= numBlocks
	backend := newAlteredBackend(t)

	block, err := backend.BlockByNumber(t.Context(), blockNum)
	require.NoErrorf(t, err, "BlockByNumber(%d)", blockNum)
	require.NotEqual(t, backend.canonicalHash(blockNum), block.Hash(), "test precondition: altered header hashes differently")

	t.Run("not_implemented", func(t *testing.T) {
		api := NewAPI(backend.testBackend)
		require.Equal(t, block.Hash(), api.blockHash(block), "API.blockHash() with a Backend that is not a BlockHashOverrider")
	})

	t.Run("implemented", func(t *testing.T) {
		api := NewAPI(backend)
		require.Equal(t, backend.canonicalHash(blockNum), api.blockHash(block), "API.blockHash() with a Backend that is a BlockHashOverrider")
	})
}

// TestOverrideBlockHash_TraceTransactionResolvesBlock tests block resolution by
// canonical hash. [API.TraceTransaction] never calls [API.blockHash].
func TestOverrideBlockHash_TraceTransactionResolvesBlock(t *testing.T) {
	const blockNum = 1 // must be <= numBlocks
	backend := newAlteredBackend(t)
	api := NewAPI(backend)

	block, err := backend.BlockByNumber(t.Context(), blockNum)
	require.NoErrorf(t, err, "BlockByNumber(%d)", blockNum)
	require.NotEmptyf(t, block.Transactions(), "%T.Transactions()", block)
	txHash := block.Transactions()[0].Hash()

	tracer := blockHashCaptureTracerName
	res, err := api.TraceTransaction(t.Context(), txHash, &TraceConfig{Tracer: &tracer})
	require.NoErrorf(t, err, "TraceTransaction()")
	require.Equal(t, backend.canonicalHash(blockNum), hashFromTraceResult(t, res), "Context.BlockHash received by tracer")
}

// TestOverrideBlockHash_StandardTraceBlockToFile tests that dump files are named
// after the canonical block hash.
func TestOverrideBlockHash_StandardTraceBlockToFile(t *testing.T) {
	const blockNum = 1 // must be <= numBlocks
	backend := newAlteredBackend(t)
	api := NewAPI(backend)

	files, err := api.StandardTraceBlockToFile(t.Context(), backend.canonicalHash(blockNum), nil)
	for _, file := range files {
		defer os.Remove(file)
	}
	require.NoError(t, err, "StandardTraceBlockToFile()")
	require.NotEmpty(t, files, "dump files")

	wantPrefix := fmt.Sprintf("block_%#x-", backend.canonicalHash(blockNum).Bytes()[:4])
	for _, file := range files {
		require.Containsf(t, file, wantPrefix, "dump file %q named after canonical block hash", file)
	}
}
