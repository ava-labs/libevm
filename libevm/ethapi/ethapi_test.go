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

package ethapi

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/libevm/core/types"
	"github.com/ava-labs/libevm/ethclient"
	"github.com/ava-labs/libevm/params"
	"github.com/ava-labs/libevm/rpc"
)

// Header and block hooks write to the same map in eth_getBlockBy*, so they
// MUST use different keys.
const (
	headerExtraKey = "libevm_header_extra_field"
	blockExtraKey  = "libevm_block_extra_field"
)

// decodeKey returns a map containing only the key field of the JSON object
// raw, decoded as a T.
func decodeKey[T any](raw []byte, key string) (map[string]any, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return nil, err
	}
	var v T
	if err := json.Unmarshal(fields[key], &v); err != nil {
		return nil, err
	}
	return map[string]any{key: v}, nil
}

type headerHooks struct {
	add map[string]any
	types.NOOPHeaderHooks
}

func (hh *headerHooks) PostRPCMarshal(_ *types.Header, m map[string]any) {
	maps.Copy(m, hh.add)
}

func (hh *headerHooks) DecodeJSON(h *types.Header, raw []byte) error {
	if err := hh.NOOPHeaderHooks.DecodeJSON(h, raw); err != nil {
		return err
	}
	add, err := decodeKey[int](raw, headerExtraKey)
	if err != nil {
		return err
	}
	hh.add = add
	return nil
}

type blockHooks struct {
	add map[string]any
	types.NOOPBlockBodyHooks
}

func (bh *blockHooks) PostRPCMarshal(_ *types.Block, m map[string]any) {
	maps.Copy(m, bh.add)
}

func (bh *blockHooks) PostRPCUnmarshal(_ *types.Block, raw []byte) error {
	add, err := decodeKey[int](raw, blockExtraKey)
	if err != nil {
		return err
	}
	bh.add = add
	return nil
}

func (b *blockHooks) Copy() *blockHooks {
	return &blockHooks{
		add: maps.Clone(b.add),
	}
}

type backend struct {
	blocks map[common.Hash]*types.Block
	Backend
}

func (be *backend) BlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error) {
	b, ok := be.blocks[hash]
	if !ok {
		return nil, fmt.Errorf("%v not found", hash)
	}
	return b, nil
}

func (be *backend) HeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error) {
	b, err := be.BlockByHash(ctx, hash)
	if err != nil {
		return nil, err
	}
	return b.Header(), nil
}

// Ancillary methods required by getters.
func (*backend) ChainConfig() *params.ChainConfig            { return params.MergedTestChainConfig }
func (*backend) GetTd(context.Context, common.Hash) *big.Int { return big.NewInt(0) }

func TestPostRPCHooks(t *testing.T) {
	extras := types.RegisterExtras[headerHooks, *headerHooks, blockHooks, *blockHooks, struct{}]()
	t.Cleanup(types.TestOnlyClearRegisteredExtras)

	const (
		headerValue int = 42
		blockValue  int = 1e6
	)

	hdr := &types.Header{
		// Required for decoding by [ethclient.Client].
		Number:     big.NewInt(0),
		Difficulty: big.NewInt(0),
		UncleHash:  types.EmptyUncleHash,
		TxHash:     types.EmptyTxsHash,
	}
	extras.Header.Set(hdr, &headerHooks{
		add: map[string]any{headerExtraKey: headerValue},
	})

	blk := types.NewBlockWithHeader(hdr)
	extras.Block.Set(blk, &blockHooks{
		add: map[string]any{blockExtraKey: blockValue},
	})

	api := NewBlockChainAPI(&backend{
		blocks: map[common.Hash]*types.Block{
			blk.Hash(): blk,
		},
	})

	t.Run("RPC API", func(t *testing.T) {
		t.Run("HeaderHooks", func(t *testing.T) {
			got := api.GetHeaderByHash(t.Context(), blk.Hash())
			assert.Equalf(t, headerValue, got[headerExtraKey], "%T.GetHeaderByHash(...)[%q]", api, headerExtraKey)
		})
		t.Run("BlockBodyHooks", func(t *testing.T) {
			got, err := api.GetBlockByHash(t.Context(), blk.Hash(), false)
			require.NoErrorf(t, err, "%T.GetBlockByHash(...)", api)
			assert.Equalf(t, headerValue, got[headerExtraKey], "%T.GetBlockByHash(...).Header()[%q]", api, headerExtraKey)
			assert.Equalf(t, blockValue, got[blockExtraKey], "%T.GetBlockByHash(...)[%q]", api, blockExtraKey)
		})
	})

	t.Run("ethclient", func(t *testing.T) {
		srv := rpc.NewServer()
		t.Cleanup(srv.Stop)
		require.NoError(t, srv.RegisterName("eth", api), "RegisterName()")
		client := ethclient.NewClient(rpc.DialInProc(srv))
		t.Cleanup(client.Close)

		t.Run("HeaderHooks", func(t *testing.T) {
			got, err := client.HeaderByHash(t.Context(), blk.Hash())
			require.NoErrorf(t, err, "%T.HeaderByHash(...)", client)
			assert.Equalf(t, headerValue, extras.Header.Get(got).add[headerExtraKey], "%T.HeaderByHash(...) extra", client)
		})
		t.Run("BlockBodyHooks", func(t *testing.T) {
			got, err := client.BlockByHash(t.Context(), blk.Hash())
			require.NoErrorf(t, err, "%T.BlockByHash(...)", client)
			assert.Equalf(t, blk.Hash(), got.Hash(), "%T.BlockByHash(...).Hash()", client)
			assert.Equalf(t, blockValue, extras.Block.Get(got).add[blockExtraKey], "%T.BlockByHash(...) extra", client)
			assert.Equalf(t, headerValue, extras.Header.Get(got.Header()).add[headerExtraKey], "%T.BlockByHash(...).Header() extra", client)
		})
	})
}
