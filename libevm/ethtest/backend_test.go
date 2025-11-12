package ethtest

import (
	"math/big"
	"testing"

	"github.com/ava-labs/libevm/accounts/abi/bind"
	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/libevm/core"
	"github.com/ava-labs/libevm/core/types"
	"github.com/ava-labs/libevm/crypto"
	"github.com/ava-labs/libevm/params"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/require"
)

// The above copyright and licensing exclude the original WETH9 contract and
// compiled artefacts, which are licensed under the following:
//
// Copyright (C) 2015, 2016, 2017 Dapphub
//
// This program is free software: you can redistribute it and/or modify it under
// the terms of the GNU General Public License as published by the Free Software
// Foundation, either version 3 of the License, or (at your option) any later
// version.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
// FOR A PARTICULAR PURPOSE.  See the GNU General Public License for more
// details.
//
// You should have received a copy of the GNU General Public License along with
// this program. If not, see <http://www.gnu.org/licenses/>.

//go:generate go run ../../cmd/abigen --abi weth.abi --bin weth.bin --pkg ethtest --type weth --out weth_test.go

func TestMinimalBackend(t *testing.T) {
	key, err := crypto.GenerateKey()
	require.NoError(t, err)
	eoa := crypto.PubkeyToAddress(key.PublicKey)

	backend, signer := NewMinimalBackend(
		t,
		WithChainConfig(params.AllDevChainProtocolChanges),
		WithGenesis(&core.Genesis{
			Alloc: types.GenesisAlloc{
				eoa: {
					Balance: new(uint256.Int).SetAllOne().ToBig(),
				},
			},
		}),
	)

	opts := &bind.TransactOpts{
		From: eoa,
		Signer: func(_ common.Address, tx *types.Transaction) (*types.Transaction, error) {
			return types.SignTx(tx, signer, key)
		},
		Value: big.NewInt(42),
	}

	_, _, weth, err := DeployWeth(opts, backend)
	require.NoError(t, err)

	got := backend.ResultOf(t)(weth.Deposit(opts))
	require.False(t, got.Failed())

	bal, err := weth.BalanceOf(nil, eoa)
	require.NoError(t, err)
	require.True(t, bal.Cmp(opts.Value) == 0)
}

func ExampleMinimalBackend_ResultOf() {

}
