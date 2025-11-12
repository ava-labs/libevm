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
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ava-labs/libevm/core"
	"github.com/ava-labs/libevm/core/rawdb"
	"github.com/ava-labs/libevm/core/state"
	"github.com/ava-labs/libevm/core/types"
	"github.com/ava-labs/libevm/core/vm"
	"github.com/ava-labs/libevm/ethdb"
	"github.com/ava-labs/libevm/libevm/options"
	"github.com/ava-labs/libevm/params"
	"github.com/ava-labs/libevm/triedb"
)

// NewEmptyStateDB returns an empty, memory-backed, state database.
func NewEmptyStateDB(tb testing.TB, disk ethdb.Database) *state.StateDB {
	tb.Helper()
	sdb, err := state.New(types.EmptyRootHash, state.NewDatabase(disk), nil)
	require.NoError(tb, err, "state.New()")
	return sdb
}

// NewZeroEVM returns a new EVM backed by a [rawdb.NewMemoryDatabase]; all other
// arguments to [vm.NewEVM] are the zero values of their respective types,
// except for the use of [core.CanTransfer] and [core.Transfer] instead of nil
// functions.
func NewZeroEVM(tb testing.TB, opts ...EVMOption) (*state.StateDB, *vm.EVM) {
	tb.Helper()

	args := &evmConstructorArgs{
		vm.BlockContext{
			CanTransfer: core.CanTransfer,
			Transfer:    core.Transfer,
		},
		vm.TxContext{},
		&params.ChainConfig{},
		vm.Config{},
		&core.Genesis{},
	}
	args = options.ApplyTo(args, opts...)

	sdb := args.stateDB(tb)

	return sdb, vm.NewEVM(
		args.blockContext,
		args.txContext,
		sdb,
		args.chainConfig,
		args.config,
	)
}

type evmConstructorArgs struct {
	blockContext vm.BlockContext
	txContext    vm.TxContext
	chainConfig  *params.ChainConfig
	config       vm.Config
	genesis      *core.Genesis
}

func (args *evmConstructorArgs) stateDB(tb testing.TB) *state.StateDB {
	tb.Helper()

	disk := rawdb.NewMemoryDatabase()
	if args.genesis == nil {
		return NewEmptyStateDB(tb, disk)
	}

	args.genesis.Config = args.chainConfig
	tdb := triedb.NewDatabase(disk, nil)
	_, root, err := core.SetupGenesisBlock(disk, tdb, args.genesis)
	require.NoError(tb, err, "core.SetupGenesisBlock()")
	require.NoError(tb, tdb.Commit(root, false), "%T.Commit([genesis root])", tdb)

	cache := state.NewDatabase(disk)
	sdb, err := state.New(root, cache, nil)
	require.NoErrorf(tb, err, "state.New(%#x, ...)", root)
	return sdb
}

// An EVMOption configures the EVM returned by [NewZeroEVM].
type EVMOption = options.Option[evmConstructorArgs]

// WithBlockContext overrides the default context.
func WithBlockContext(c vm.BlockContext) EVMOption {
	return options.Func[evmConstructorArgs](func(args *evmConstructorArgs) {
		args.blockContext = c
	})
}

// WithBlockContext overrides the default context.
func WithChainConfig(c *params.ChainConfig) EVMOption {
	return options.Func[evmConstructorArgs](func(args *evmConstructorArgs) {
		args.chainConfig = c
	})
}

// WithGenesis overrides the default, empty genesis. The [params.ChainConfig]
// will be ignored; use [WithChainConfig] if necessary.
func WithGenesis(g *core.Genesis) EVMOption {
	return options.Func[evmConstructorArgs](func(args *evmConstructorArgs) {
		args.genesis = g
	})
}
