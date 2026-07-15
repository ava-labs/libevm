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
	"fmt"
	"math/big"

	"github.com/ava-labs/libevm/core"
	"github.com/ava-labs/libevm/core/state"
	"github.com/ava-labs/libevm/core/types"
	"github.com/ava-labs/libevm/core/vm"
)

// DefaultTraceTimeout exports defaultTraceTimeout.
const DefaultTraceTimeout = defaultTraceTimeout

// ExtraExecutor is an optional interface that a [Backend] can implement to
// inject extra state changes when the tracing APIs re-execute blocks.
type ExtraExecutor interface {
	// BeforeExecutingBlock is called on the state returned by
	// [Backend.StateAtBlock] before any of the block's transactions are
	// applied.
	BeforeExecutingBlock(sdb *state.StateDB, parent, header *types.Header) error
	// AfterExecutingTransaction is called on the state for each transaction
	// after being applied.
	AfterExecutingTransaction(sdb *state.StateDB, bf *big.Int, gasUsed uint64) error
}

func beforeExecutingBlock(b Backend, sdb *state.StateDB, parent, header *types.Header) error {
	ex, ok := b.(ExtraExecutor)
	if !ok {
		return nil
	}
	if err := ex.BeforeExecutingBlock(sdb, parent, header); err != nil {
		return fmt.Errorf("before executing block failed: %w", err)
	}
	return nil
}

func applyMessage(b Backend, vmenv *vm.EVM, sdb *state.StateDB, message *core.Message) error {
	res, err := core.ApplyMessage(vmenv, message, new(core.GasPool).AddGas(message.GasLimit))
	if err != nil {
		return fmt.Errorf("tracing failed: %w", err)
	}

	ex, ok := b.(ExtraExecutor)
	if !ok {
		return nil
	}
	if err := ex.AfterExecutingTransaction(sdb, vmenv.Context.BaseFee, res.UsedGas); err != nil {
		return fmt.Errorf("after executing transaction failed: %w", err)
	}
	return nil
}
