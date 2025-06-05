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
package core_test

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/require"

	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/libevm/core"
	"github.com/ava-labs/libevm/core/types"
	"github.com/ava-labs/libevm/crypto"
	"github.com/ava-labs/libevm/libevm"
	"github.com/ava-labs/libevm/libevm/ethtest"
	"github.com/ava-labs/libevm/libevm/hookstest"
	"github.com/ava-labs/libevm/params"
)

func TestCanExecuteTransaction(t *testing.T) {
	rng := ethtest.NewPseudoRand(42)
	account := rng.Address()
	slot := rng.Hash()

	makeErr := func(from common.Address, to *common.Address, val common.Hash) error {
		return fmt.Errorf("From: %v To: %v State: %v", from, to, val)
	}
	hooks := &hookstest.Stub{
		CanExecuteTransactionFn: func(from common.Address, to *common.Address, s libevm.StateReader) error {
			return makeErr(from, to, s.GetState(account, slot))
		},
	}
	hooks.Register(t)

	value := rng.Hash()

	state, evm := ethtest.NewZeroEVM(t)
	state.SetState(account, slot, value)
	msg := &core.Message{
		From: rng.Address(),
		To:   rng.AddressPtr(),
	}
	_, err := core.ApplyMessage(evm, msg, new(core.GasPool).AddGas(30e6))
	require.EqualError(t, err, makeErr(msg.From, msg.To, value).Error())
}

func TestMinimumGasConsumption(t *testing.T) {
	// All transactions will be basic transfers so consume [params.TxGas] by
	// default.
	tests := []struct {
		name                     string
		gasLimit, minConsumption uint64
		wantUsed                 uint64
	}{
		{
			name:           "consume_extra",
			gasLimit:       1e6,
			minConsumption: 5e5,
			wantUsed:       5e5,
		},
		{
			name:           "consume_extra",
			gasLimit:       1e6,
			minConsumption: 4e5,
			wantUsed:       4e5,
		},
		{
			name:           "no_extra_consumption",
			gasLimit:       50_000,
			minConsumption: params.TxGas - 1,
			wantUsed:       params.TxGas,
		},
		{
			name:           "zero_min",
			gasLimit:       50_000,
			minConsumption: 0,
			wantUsed:       params.TxGas,
		},
		{
			name:           "consume_extra_by_one",
			gasLimit:       1e6,
			minConsumption: params.TxGas + 1,
			wantUsed:       params.TxGas + 1,
		},
		{
			name:           "min_capped_at_limit",
			gasLimit:       1e6,
			minConsumption: 2e6,
			wantUsed:       1e6,
		},
	}

	const gasPrice = params.Wei

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hooks := &hookstest.Stub{
				MinimumGasConsumptionFn: func(limit uint64) uint64 {
					require.Equal(t, limit, tt.gasLimit)
					return tt.minConsumption
				},
			}
			hooks.Register(t)

			key, err := crypto.GenerateKey()
			require.NoError(t, err, "libevm/crypto.GenerateKey()")

			stateDB, evm := ethtest.NewZeroEVM(t)
			signer := types.LatestSigner(evm.ChainConfig())
			tx := types.MustSignNewTx(
				key, signer,
				&types.LegacyTx{
					GasPrice: big.NewInt(gasPrice),
					Gas:      tt.gasLimit,
					To:       &common.Address{},
					Value:    big.NewInt(0),
				},
			)
			msg, err := core.TransactionToMessage(tx, signer, big.NewInt(gasPrice))
			require.NoError(t, err, "core.TransactionToMessage(types.MustSignNewTx(...))")

			const startingBalance = 10 * params.Ether
			stateDB.SetNonce(msg.From, 0)
			stateDB.SetBalance(msg.From, uint256.NewInt(startingBalance))

			gotPool := core.GasPool(1e9) // modified when passed as pointer
			wantPool := gotPool - core.GasPool(tt.wantUsed)

			got, err := core.ApplyMessage(evm, msg, &gotPool)
			require.NoError(t, err, "core.ApplyMessage()")

			want := &core.ExecutionResult{
				UsedGas: tt.wantUsed,
			}
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("core.ApplyMessage(...) diff (-want +got):\n%s", diff)
			}
			if gotPool != wantPool {
				t.Errorf("After core.ApplyMessage(..., *%T); got %[1]T = %[1]d; want %d", gotPool, wantPool)
			}

			wantBalance := startingBalance - tt.wantUsed*gasPrice
			if got := stateDB.GetBalance(msg.From); !got.IsUint64() || got.Uint64() != wantBalance {
				t.Errorf("got remaining balance %s; want %d", got.String(), wantBalance)
			}
		})
	}
}
