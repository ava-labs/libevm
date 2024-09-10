package vm

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/libevm"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/rand"
)

// precompileOverrides is a [params.RulesHooks] that overrides precompiles from
// a map of predefined addresses.
type precompileOverrides struct {
	contracts        map[common.Address]PrecompiledContract
	params.NOOPHooks // all other hooks
}

func (o precompileOverrides) PrecompileOverride(a common.Address) (libevm.PrecompiledContract, bool) {
	c, ok := o.contracts[a]
	return c, ok
}

// A precompileStub is a [PrecompiledContract] that always returns the same
// values.
type precompileStub struct {
	requiredGas uint64
	returnData  []byte
}

func (s *precompileStub) RequiredGas([]byte) uint64  { return s.requiredGas }
func (s *precompileStub) Run([]byte) ([]byte, error) { return s.returnData, nil }

func TestPrecompileOverride(t *testing.T) {
	type test struct {
		name        string
		addr        common.Address
		requiredGas uint64
		stubData    []byte
	}

	const gasLimit = uint64(1e7)

	tests := []test{
		{
			name:        "arbitrary values",
			addr:        common.Address{'p', 'r', 'e', 'c', 'o', 'm', 'p', 'i', 'l', 'e'},
			requiredGas: 314159,
			stubData:    []byte("the return data"),
		},
	}

	rng := rand.New(rand.NewSource(42))
	for _, addr := range PrecompiledAddressesCancun {
		tests = append(tests, test{
			name:        fmt.Sprintf("existing precompile %v", addr),
			addr:        addr,
			requiredGas: rng.Uint64n(gasLimit),
			stubData:    addr[:],
		})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			precompile := &precompileStub{
				requiredGas: tt.requiredGas,
				returnData:  tt.stubData,
			}

			params.TestOnlyClearRegisteredExtras()
			params.RegisterExtras(params.Extras[params.NOOPHooks, precompileOverrides]{
				NewRules: func(_ *params.ChainConfig, _ *params.Rules, _ *params.NOOPHooks, blockNum *big.Int, isMerge bool, timestamp uint64) *precompileOverrides {
					return &precompileOverrides{
						contracts: map[common.Address]PrecompiledContract{
							tt.addr: precompile,
						},
					}
				},
			})

			t.Run(fmt.Sprintf("%T.Call([overridden precompile address = %v])", &EVM{}, tt.addr), func(t *testing.T) {
				gotData, gotGasLeft, err := newEVM(t).Call(AccountRef{}, tt.addr, nil, gasLimit, uint256.NewInt(0))
				require.NoError(t, err)
				assert.Equal(t, tt.stubData, gotData, "contract's return data")
				assert.Equal(t, gasLimit-tt.requiredGas, gotGasLeft, "gas left")
			})
		})
	}
}

func newEVM(t *testing.T) *EVM {
	t.Helper()

	sdb, err := state.New(common.Hash{}, state.NewDatabase(rawdb.NewMemoryDatabase()), nil)
	require.NoError(t, err, "state.New()")

	return NewEVM(
		BlockContext{
			Transfer: func(_ StateDB, _, _ common.Address, _ *uint256.Int) {},
		},
		TxContext{},
		sdb,
		&params.ChainConfig{},
		Config{},
	)
}
