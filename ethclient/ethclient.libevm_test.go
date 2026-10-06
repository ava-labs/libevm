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

package ethclient

import (
	"context"
	"encoding/json"
	"math/big"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ava-labs/libevm/common/hexutil"
	"github.com/ava-labs/libevm/core/types"
	"github.com/ava-labs/libevm/internal/ethapi"
	"github.com/ava-labs/libevm/params"
	"github.com/ava-labs/libevm/rpc"
)

const blockExtraKey = "extra"

type blockHooks struct {
	Extra hexutil.Bytes
	types.NOOPBlockBodyHooks
}

func (bh *blockHooks) Copy() *blockHooks {
	return &blockHooks{
		Extra: slices.Clone(bh.Extra),
	}
}

func (bh *blockHooks) PostRPCMarshal(_ *types.Block, m map[string]any) {
	m[blockExtraKey] = bh.Extra
}

func (bh *blockHooks) PostRPCUnmarshal(_ *types.Block, raw []byte) error {
	var fields struct {
		Extra hexutil.Bytes `json:"extra"`
	}
	if err := json.Unmarshal(raw, &fields); err != nil {
		return err
	}
	bh.Extra = fields.Extra
	return nil
}

// blockService serves a single block via eth_getBlockByNumber, using the same
// marshalling as the canonical API so that [types.BlockBodyHooks.PostRPCMarshal]
// is invoked.
type blockService struct {
	block    *types.Block
	addField map[string]any
}

func (s *blockService) GetBlockByNumber(context.Context, rpc.BlockNumber, bool) (map[string]any, error) {
	m := ethapi.RPCMarshalBlock(s.block, true, false, params.TestChainConfig)
	for k, v := range s.addField {
		m[k] = v
	}
	return m, nil
}

func TestBlockBodyHooksRPCRoundTrip(t *testing.T) {
	extras := types.RegisterExtras[
		types.NOOPHeaderHooks, *types.NOOPHeaderHooks,
		blockHooks, *blockHooks,
		struct{},
	]()
	t.Cleanup(types.TestOnlyClearRegisteredExtras)

	block := types.NewBlockWithHeader(&types.Header{
		Number:     big.NewInt(1),
		Difficulty: big.NewInt(0),
		UncleHash:  types.EmptyUncleHash,
		TxHash:     types.EmptyTxsHash,
	})
	want := hexutil.Bytes("hello")
	extras.Block.Set(block, &blockHooks{Extra: want})

	tests := []struct {
		name        string
		addField    map[string]any
		wantErrType error
	}{
		{
			name: "extra_payload",
		},
		{
			name:        "hook_error_propagated",
			addField:    map[string]any{blockExtraKey: 42},
			wantErrType: new(json.UnmarshalTypeError),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := rpc.NewServer()
			t.Cleanup(srv.Stop)
			require.NoError(t, srv.RegisterName("eth", &blockService{
				block:    block,
				addField: tt.addField,
			}))
			client := NewClient(rpc.DialInProc(srv))
			t.Cleanup(client.Close)

			got, err := client.BlockByNumber(t.Context(), big.NewInt(1))
			if tt.wantErrType != nil {
				require.ErrorAs(t, err, &tt.wantErrType)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, block.Hash(), got.Hash(), "block hash")
			assert.Equal(t, want, extras.Block.Get(got).Extra, "extra payload")
		})
	}
}
