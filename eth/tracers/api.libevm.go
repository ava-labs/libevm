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
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/libevm/core"
	"github.com/ava-labs/libevm/core/state"
	"github.com/ava-labs/libevm/core/types"
	"github.com/ava-labs/libevm/core/vm"
	"github.com/ava-labs/libevm/eth/tracers/logger"
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

// TraceTx traces a transaction executed by `execute`, which MUST apply the
// transaction with the provided [vm.Config] attached. The [Tracer] specified
// by config observes the execution.
func TraceTx(ctx context.Context, config *TraceConfig, txctx *Context, execute func(vm.Config) error) (interface{}, error) {
	var (
		tracer  Tracer
		err     error
		timeout = defaultTraceTimeout
	)
	if config == nil {
		config = &TraceConfig{}
	}
	// Default tracer is the struct logger
	tracer = logger.NewStructLogger(config.Config)
	if config.Tracer != nil {
		tracer, err = DefaultDirectory.New(*config.Tracer, txctx, config.TracerConfig)
		if err != nil {
			return nil, err
		}
	}
	cancelling := &evmCancellingTracer{Tracer: tracer}

	// Define a meaningful timeout of a single transaction trace
	if config.Timeout != nil {
		if timeout, err = time.ParseDuration(*config.Timeout); err != nil {
			return nil, err
		}
	}
	deadlineCtx, cancel := context.WithTimeout(ctx, timeout)
	go func() {
		<-deadlineCtx.Done()
		if errors.Is(deadlineCtx.Err(), context.DeadlineExceeded) {
			// Also cancels the EVM; note cancellation is not necessarily
			// immediate.
			cancelling.Stop(errors.New("execution timeout"))
		}
	}()
	defer cancel()

	if err := execute(vm.Config{Tracer: cancelling}); err != nil {
		return nil, err
	}
	return cancelling.GetResult()
}

// evmCancellingTracer wraps a [Tracer] so that Stop also cancels the [vm.EVM]
// executing the traced transaction.
type evmCancellingTracer struct {
	Tracer

	mu      sync.Mutex
	env     *vm.EVM
	stopped bool
}

func (t *evmCancellingTracer) CaptureStart(env *vm.EVM, from, to common.Address, create bool, input []byte, gas uint64, value *big.Int) {
	t.mu.Lock()
	t.env = env
	stopped := t.stopped
	t.mu.Unlock()

	if stopped {
		env.Cancel()
	}
	t.Tracer.CaptureStart(env, from, to, create, input, gas, value)
}

func (t *evmCancellingTracer) Stop(err error) {
	t.mu.Lock()
	t.stopped = true
	env := t.env
	t.mu.Unlock()

	t.Tracer.Stop(err)
	if env != nil {
		env.Cancel()
	}
}
