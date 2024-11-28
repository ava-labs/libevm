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
package main

import (
	"context"
	"math/big"
	"testing"

	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ava-labs/libevm/accounts/abi/bind"
	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/libevm/core/types"
	"github.com/ava-labs/libevm/core/vm"
	"github.com/ava-labs/libevm/crypto"
	"github.com/ava-labs/libevm/eth/ethconfig"
	"github.com/ava-labs/libevm/ethclient/simulated"
	"github.com/ava-labs/libevm/libevm"
	"github.com/ava-labs/libevm/libevm/ethtest"
	"github.com/ava-labs/libevm/libevm/hookstest"
	"github.com/ava-labs/libevm/libevm/precompilegen/testprecompile"
	"github.com/ava-labs/libevm/node"
	"github.com/ava-labs/libevm/params"
)

//go:generate solc -o ./ --overwrite --abi --bin Test.sol
//go:generate go run . -in IPrecompile.abi -out ./testprecompile/generated.go -package testprecompile
//go:generate go run ../../cmd/abigen --abi PrecompileTest.abi --bin PrecompileTest.bin --pkg main --out ./abigen.gen_test.go --type PrecompileTest

func successfulTxReceipt(ctx context.Context, tb testing.TB, client bind.DeployBackend, tx *types.Transaction) *types.Receipt {
	tb.Helper()
	r, err := bind.WaitMined(ctx, client, tx)
	require.NoErrorf(tb, err, "bind.WaitMined(tx %#x)", tx.Hash())
	require.Equalf(tb, uint64(1), r.Status, "%T.Status", r)
	return r
}

func TestGeneratedPrecompile(t *testing.T) {
	ctx := context.Background()
	rng := ethtest.NewPseudoRand(424242)
	precompile := rng.Address()

	hooks := &hookstest.Stub{
		PrecompileOverrides: map[common.Address]libevm.PrecompiledContract{
			precompile: testprecompile.New(contract{}),
		},
	}
	extras := hookstest.Register(t, params.Extras[*hookstest.Stub, *hookstest.Stub]{
		NewRules: func(_ *params.ChainConfig, r *params.Rules, _ *hookstest.Stub, blockNum *big.Int, isMerge bool, timestamp uint64) *hookstest.Stub {
			r.IsCancun = true // enable PUSH0
			return hooks
		},
	})

	key := rng.UnsafePrivateKey(t)
	eoa := crypto.PubkeyToAddress(key.PublicKey)

	sim := simulated.NewBackend(
		types.GenesisAlloc{
			eoa: types.Account{
				Balance: new(uint256.Int).Not(uint256.NewInt(0)).ToBig(),
			},
		},
		func(nodeConf *node.Config, ethConf *ethconfig.Config) {
			ethConf.Genesis.GasLimit = 30e6
			extras.SetOnChainConfig(ethConf.Genesis.Config, hooks)
		},
	)
	defer sim.Close()

	txOpts, err := bind.NewKeyedTransactorWithChainID(key, big.NewInt(1337))
	require.NoError(t, err, "bind.NewKeyedTransactorWithChainID(..., 1337)")
	txOpts.GasLimit = 30e6

	client := sim.Client()
	_, tx, test, err := DeployPrecompileTest(txOpts, client, precompile)
	require.NoError(t, err, "DeployPrecompileTest(...)")
	sim.Commit()
	successfulTxReceipt(ctx, t, client, tx)
	suite := &PrecompileTestSession{
		Contract:     test,
		TransactOpts: *txOpts,
	}

	tests := []struct {
		transact        func() (*types.Transaction, error)
		wantCalledEvent string
	}{
		{
			transact: func() (*types.Transaction, error) {
				return suite.Echo(rng.BigUint64())
			},
			wantCalledEvent: "Echo(uint256)",
		},
		{
			transact: func() (*types.Transaction, error) {
				return suite.Echo0("hello world")
			},
			wantCalledEvent: "Echo(string)",
		},
		{
			transact: func() (*types.Transaction, error) {
				return suite.Extract(IPrecompileWrapper{
					Val: rng.BigUint64(),
				})
			},
			wantCalledEvent: "Extract(...)",
		},
		{
			transact: func() (*types.Transaction, error) {
				return suite.HashPacked(rng.BigUint64(), [2]byte{42, 42}, rng.Address())
			},
			wantCalledEvent: "HashPacked(...)",
		},
		{
			transact: func() (*types.Transaction, error) {
				return suite.Self()
			},
			wantCalledEvent: "Self()",
		},
		{
			transact: func() (*types.Transaction, error) {
				return suite.RevertWith(rng.Bytes(8))
			},
			wantCalledEvent: "RevertWith(...)",
		},
		{
			transact: func() (*types.Transaction, error) {
				return suite.EchoingFallback(rng.Bytes(8))
			},
			wantCalledEvent: "EchoingFallback(...)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.wantCalledEvent, func(t *testing.T) {
			tx, err := tt.transact()
			require.NoError(t, err, "send tx")
			sim.Commit()

			rcpt := successfulTxReceipt(ctx, t, client, tx)
			require.Equalf(t, uint64(1), rcpt.Status, "%T.Status (i.e. transaction included)", rcpt)

			require.Lenf(t, rcpt.Logs, 1, "%T.Logs", rcpt)
			called, err := test.ParseCalled(*rcpt.Logs[0])
			require.NoErrorf(t, err, "%T.ParseCalled(...)", test)
			assert.Equal(t, tt.wantCalledEvent, called.Arg0, "function name emitted with `Called` event")
		})
	}
}

type contract struct{}

var _ testprecompile.Contract = contract{}

func (contract) Fallback(env vm.PrecompileEnvironment, callData []byte) ([]byte, uint64, error) {
	// Note the test-suite assumption of the fallback's behaviour:
	var _ = (*PrecompileTest).EchoingFallback
	return callData, 0, nil
}

func (contract) Echo(env vm.PrecompileEnvironment, x *big.Int) (*big.Int, error) {
	return x, nil
}

func (contract) Echo0(env vm.PrecompileEnvironment, x string) (string, error) {
	return x, nil
}

func (contract) Extract(env vm.PrecompileEnvironment, x struct {
	Val *big.Int "json:\"val\""
}) (*big.Int, error) {
	return x.Val, nil
}

func (contract) HashPacked(env vm.PrecompileEnvironment, x *big.Int, y [2]byte, z common.Address) (hash [32]byte, _ error) {
	copy(
		hash[:],
		crypto.Keccak256(
			uint256.MustFromBig(x).PaddedBytes(32),
			y[:],
			z.Bytes(),
		),
	)
	return hash, nil
}

func (contract) RevertWith(env vm.PrecompileEnvironment, x []byte) error {
	return vm.RevertError(x)
}

func (contract) Self(env vm.PrecompileEnvironment) (common.Address, error) {
	return env.Addresses().Self, nil
}
