// Copyright 2024 the libevm authors.
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

// Package ethtest provides utility functions for use in testing
// Ethereum-related functionality.
package ethtest

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/libevm/core"
	"github.com/ava-labs/libevm/core/rawdb"
	"github.com/ava-labs/libevm/core/state"
	"github.com/ava-labs/libevm/core/types"
	"github.com/ava-labs/libevm/core/vm"
	"github.com/ava-labs/libevm/ethdb"
	"github.com/ava-labs/libevm/params"
)

// NewEmptyStateDB returns a fresh database from [rawdb.NewMemoryDatabase], a
// [state.Database] wrapping it, and a [state.StateDB] wrapping that, opened to
// [types.EmptyRootHash].
func NewEmptyStateDB(tb testing.TB) (ethdb.Database, state.Database, *state.StateDB) {
	tb.Helper()

	db := rawdb.NewMemoryDatabase()
	cache := state.NewDatabase(db)
	sdb, err := state.New(types.EmptyRootHash, cache, nil)
	require.NoError(tb, err, "state.New()")
	return db, cache, sdb
}

// NewZeroEVM returns a new EVM backed by a [rawdb.NewMemoryDatabase]; all other
// arguments to [vm.NewEVM] are the zero values of their respective types,
// except for the use of [core.CanTransfer] and [core.Transfer] instead of nil
// functions.
func NewZeroEVM(tb testing.TB, opts ...EVMOption) (*state.StateDB, *vm.EVM) {
	tb.Helper()

	_, _, sdb := NewEmptyStateDB(tb)

	args := &evmConstructorArgs{
		vm.BlockContext{
			CanTransfer: core.CanTransfer,
			Transfer:    core.Transfer,
		},
		vm.TxContext{},
		sdb,
		&params.ChainConfig{},
		vm.Config{},
	}
	for _, o := range opts {
		o.apply(args)
	}

	return sdb, vm.NewEVM(
		args.blockContext,
		args.txContext,
		args.stateDB,
		args.chainConfig,
		args.config,
	)
}

type evmConstructorArgs struct {
	blockContext vm.BlockContext
	txContext    vm.TxContext
	stateDB      vm.StateDB
	chainConfig  *params.ChainConfig
	config       vm.Config
}

// An EVMOption configures the EVM returned by [NewZeroEVM].
type EVMOption interface {
	apply(*evmConstructorArgs)
}

type funcOption func(*evmConstructorArgs)

var _ EVMOption = funcOption(nil)

func (f funcOption) apply(args *evmConstructorArgs) { f(args) }

// WithBlockContext overrides the default context.
func WithBlockContext(c vm.BlockContext) EVMOption {
	return funcOption(func(args *evmConstructorArgs) {
		args.blockContext = c
	})
}

// WithChainConfig overrides the default chain config. SHOULD be used with
// [WithBlockContext] to set a non-nil block number that activates forks defined
// in the config. Prefer [WithBlockNumberAndChainConfig]. A non-nil `Random`
// field is also required for post-Merge forks.
func WithChainConfig(c *params.ChainConfig) EVMOption {
	return funcOption(func(args *evmConstructorArgs) {
		args.chainConfig = c
	})
}

// WithBlockNumberAndChainConfig overrides the default block number and chain
// config. The [vm.BlockContext] used to set the block number has a non-nil
// `Random` field as this is part of the definition of a post-merge fork.
func WithBlockNumberAndChainConfig(n uint64, c *params.ChainConfig) EVMOption {
	hdr := &types.Header{
		Number:     new(big.Int).SetUint64(n),
		BaseFee:    big.NewInt(0),
		Difficulty: big.NewInt(0), // sets non-nil random
	}
	return funcOption(func(args *evmConstructorArgs) {
		args.blockContext = core.NewEVMBlockContext(hdr, nil, &common.Address{})
		args.chainConfig = c
	})
}

// WithAllEIPs is a convenience wrapper for [WithBlockNumberAndChainConfig] with
// block number 1 and [params.MergedTestChainConfig] as arguments.
func WithAllEIPs() EVMOption {
	return WithBlockNumberAndChainConfig(1, params.MergedTestChainConfig)
}
