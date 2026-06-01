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
	"fmt"
	"maps"
	"math/big"
	"testing"

	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/libevm/core/types"
	"github.com/ava-labs/libevm/params"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type headerHooks struct {
	add map[string]any
	types.NOOPHeaderHooks
}

func (hh *headerHooks) PostRPCMarshal(_ *types.Header, m map[string]any) {
	maps.Copy(m, hh.add)
}

type blockHooks struct {
	add map[string]any
	types.NOOPBlockBodyHooks
}

func (bh *blockHooks) PostRPCMarshal(_ *types.Block, m map[string]any) {
	maps.Copy(m, bh.add)
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

func TestPostRPCMarshalHooks(t *testing.T) {
	extras := types.RegisterExtras[headerHooks, *headerHooks, blockHooks, *blockHooks, struct{}]()
	t.Cleanup(types.TestOnlyClearRegisteredExtras)

	const (
		extraKey        = "libevm_extra_field"
		headerValue int = 42
		blockValue  int = 1e6
	)

	hdr := &types.Header{}
	extras.Header.Set(hdr, &headerHooks{
		add: map[string]any{extraKey: headerValue},
	})

	blk := types.NewBlockWithHeader(hdr)
	extras.Block.Set(blk, &blockHooks{
		add: map[string]any{extraKey: blockValue},
	})

	api := NewBlockChainAPI(&backend{
		blocks: map[common.Hash]*types.Block{
			blk.Hash(): blk,
		},
	})

	t.Run("HeaderHooks", func(t *testing.T) {
		got := api.GetHeaderByHash(t.Context(), blk.Hash())
		assert.Equalf(t, headerValue, got[extraKey], "%T.GetHeaderByHash(...)[%q]", api, extraKey)
	})
	t.Run("BlockBodyHooks", func(t *testing.T) {
		got, err := api.GetBlockByHash(t.Context(), blk.Hash(), false)
		require.NoErrorf(t, err, "%T.GetBlockByHash(...)", api)
		assert.Equalf(t, blockValue, got[extraKey], "%T.GetBlockByHash(...)[%q]", api, extraKey)
	})
}
