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

package parallel

import (
	"math/big"

	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/libevm/core/types"
	"github.com/ava-labs/libevm/core/vm"
	"github.com/ava-labs/libevm/crypto"
	"github.com/ava-labs/libevm/libevm"
	"github.com/ava-labs/libevm/libevm/eventual"
)

// A mixedHandler is a [Handler] implementation that demonstrates how to combine
// parallel and sequential steps when the interim sequential output needs to be
// reported at transaction resolution.
//
// A contrived "expensive" (but embarrassingly parallel) step of interpreting
// calldata as a [big.Int] is performed in the per-tx `Process()` method, while
// a running total is computed in `PostProcess()`. This total is reported back
// to the [PrecompileResult] via an [eventual.Value], which means
// `PostProcess()` is never blocked by the EVM, only the other way round.
type mixedHandler struct {
	self common.Address
}

var _ PrecompileResult = mixedResult{}

type mixedResult struct {
	// Parallel component, written by [mixedHandler.Process].
	x *big.Int
	// Sequential component, set by [mixedHandler.PostProcess] to be read by
	// [mixedResult.PrecompileOutput].
	effects eventual.Value[mixedResultEffects]
}

type mixedResultEffects struct {
	logs []*types.Log
	// Other state mutations could be included here too.
}

// Process implements the parallelisable component of the [mixedHandler].
func (mixedHandler) Process(sdb libevm.StateReader, tx IndexedTx, _, _ struct{}) mixedResult {
	return mixedResult{
		x:       new(big.Int).SetBytes(tx.Data()),
		effects: eventual.New[mixedResultEffects](),
	}
}

// PostProcess implements the sequential component of the [mixedHandler].
func (h mixedHandler) PostProcess(_ struct{}, rs Results[mixedResult]) struct{} {
	total := big.NewInt(0)
	topic := crypto.Keccak256Hash([]byte("RunningTotal"))

	for tx := range rs.TxOrder {
		r := tx.Result
		total.Add(total, r.x)
		// Note that [eventual.Value.Put] is non-blocking, allowing
		// [mixedHandler.PostProcess] to be ahead of the primary EVM execution.
		r.effects.Put(mixedResultEffects{
			logs: []*types.Log{{
				Address: h.self,
				Topics:  []common.Hash{topic},
				Data:    total.Bytes(),
			}},
		})
	}
	return struct{}{}
}

// PrecompileOutput waits until the respective transaction's _sequential_
// component has completed, recording effects within the context of the
// respective transaction.
//
// IMPORTANT: the given example assumes that transactions don't revert
// (otherwise the running total would be incorrect). This can only be guaranteed
// if the precompile is called directly by the EOA.
func (r mixedResult) PrecompileOutput(env vm.PrecompileEnvironment, calldata []byte) ([]byte, error) {
	sdb := env.StateDB()
	if sdb == nil {
		return nil, vm.ErrWriteProtection
	}
	for _, l := range r.effects.Peek().logs {
		sdb.AddLog(l)
	}
	return nil, nil
}

func ExampleHandler_mixedParallelAndSequential() {
}

// Here be dragons (and boilerplate code)!

var _ Handler[struct{}, struct{}, mixedResult, struct{}] = mixedHandler{}

func (mixedHandler) BeforeBlock(libevm.StateReader, *types.Header) struct{}     { return struct{}{} }
func (mixedHandler) Prefetch(libevm.StateReader, IndexedTx, struct{}) struct{}  { return struct{}{} }
func (mixedHandler) AfterBlock(StateDB, struct{}, *types.Block, types.Receipts) {}

func (h mixedHandler) ShouldProcess(tx IndexedTx, _ struct{}) (bool, uint64) {
	if to := tx.To(); to != nil && *to == h.self {
		return true, 42
	}
	return false, 0
}
