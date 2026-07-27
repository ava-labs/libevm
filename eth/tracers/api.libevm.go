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

	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/libevm/core/types"
)

// BlockHashOverrider is an optional extension to [Backend], allowing an
// arbitrary block hash to be returned for a block. If not implemented,
// [types.Block.Hash] is used instead.
//
// This would be used in the case that the [types.Header] is manipulated in the
// [Backend] for tracer-specific behavior. However, all hashes directly
// returned by the [Backend] are expected to be correct.
type BlockHashOverrider interface {
	// BlockHash returns the hash of the provided block.
	BlockHash(*types.Block) common.Hash
}

func (a *API) blockHash(block *types.Block) common.Hash {
	overrider, ok := a.backend.(BlockHashOverrider)
	if !ok {
		return block.Hash()
	}
	return overrider.BlockHash(block)
}

// TraceFullBlock traces all transactions in the given block. It is equivalent
// to [API.TraceBlock] (which takes the block RLP-encoded), so we rename it
// here. It is a function, not a method on [API], as the RPC server registers
// all exported methods as endpoints.
func TraceFullBlock(ctx context.Context, api *API, block *types.Block, config *TraceConfig) ([]*TxTraceResult, error) {
	return api.traceBlock(ctx, block, config)
}

// TxTraceResult exports the [txTraceResult] type, returned per transaction
// by the TraceBlock* methods.
type TxTraceResult = txTraceResult

// BlockTraceResult exports the [blockTraceResult] type, streamed per block
// by the TraceChain subscription.
type BlockTraceResult = blockTraceResult
