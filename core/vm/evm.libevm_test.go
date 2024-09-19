package vm

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/require"
)

type chainIDOverrider struct {
	chainID int64
}

func (o chainIDOverrider) OverrideNewEVMArgs(BlockContext, TxContext, StateDB, *params.ChainConfig, Config) (BlockContext, TxContext, StateDB, *params.ChainConfig, Config) {
	return BlockContext{}, TxContext{}, nil, &params.ChainConfig{ChainID: big.NewInt(o.chainID)}, Config{}
}

func TestOverrideNewEVMArgs(t *testing.T) {
	// The OverrideNewEVMArgs hook accepts and returns all arguments to
	// NewEVM(), in order. Here we lock in our assumption of that order. If this
	// breaks then the Hooks.OverrideNewEVMArgs() signature MUST be changed to
	// match.
	var _ func(BlockContext, TxContext, StateDB, *params.ChainConfig, Config) *EVM = NewEVM

	const chainID = 13579
	libevmHooks = nil
	RegisterHooks(chainIDOverrider{chainID: chainID})

	got := NewEVM(BlockContext{}, TxContext{}, nil, nil, Config{}).ChainConfig().ChainID
	require.Equal(t, big.NewInt(chainID), got)
}
