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

package ethapi

import (
	"errors"
	"math/big"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/libevm/consensus/beacon"
	"github.com/ava-labs/libevm/consensus/ethash"
	"github.com/ava-labs/libevm/core"
	"github.com/ava-labs/libevm/core/types"
	"github.com/ava-labs/libevm/core/vm"
	"github.com/ava-labs/libevm/params"
)

func TestCallInterceptor(t *testing.T) {
	addr := common.Address{'f', 'o', 'o'}
	const returnBufSize = 17 // arbitrary

	var got []byte
	retErr := errors.New("intercepted!")
	intercept := WithCallResultInterceptor(func(r *core.ExecutionResult) error {
		got = slices.Clone(r.ReturnData)
		return retErr
	})

	api := NewBlockChainAPI(
		newTestBackend(
			t,
			0,
			&core.Genesis{
				Config: params.TestChainConfig,
				Alloc: types.GenesisAlloc{
					addr: types.Account{
						Code: []byte{
							byte(vm.PUSH1), returnBufSize,
							byte(vm.PUSH1), 0,
							byte(vm.RETURN),
						},
					},
				},
			},
			beacon.New(ethash.NewFaker()),
			func(i int, b *core.BlockGen) {
				b.SetPoS()
			},
		),
		WithDefaultCallOptions(intercept),
	)

	_, err := api.Call(t.Context(), TransactionArgs{To: &addr}, nil, nil, nil)
	assert.Equalf(t, retErr, err, "%T.Call() error propagated from CallResultInterceptor", api)
	assert.Lenf(t, got, returnBufSize, "%T.ReturnData received by CallResultInterceptor", &core.ExecutionResult{})
}

func TestAccessListVerbatimTuples(t *testing.T) {
	var (
		from     = common.Address{'f', 'r', 'o', 'm'}
		to       = common.Address{'t', 'o'}
		verbatim = common.Address{'v', 'e', 'r', 'b'}
		other    = common.Address{'o', 't', 'h', 'e', 'r'}
	)
	// STATICCALL(gas, verbatim, 0, 0, 0, 0)
	code := append([]byte{
		byte(vm.PUSH1), 0, byte(vm.PUSH1), 0, byte(vm.PUSH1), 0, byte(vm.PUSH1), 0,
		byte(vm.PUSH20),
	}, verbatim.Bytes()...)
	code = append(code, byte(vm.GAS), byte(vm.STATICCALL), byte(vm.STOP))

	createAccessList := func(t *testing.T, caller types.AccessList, opts ...BlockChainAPIOption) types.AccessList {
		t.Helper()
		api := NewBlockChainAPI(
			newTestBackend(
				t,
				0,
				&core.Genesis{
					Config: params.TestChainConfig,
					Alloc: types.GenesisAlloc{
						from: {Balance: new(big.Int).Lsh(big.NewInt(1), 100)},
						to:   {Code: code},
					},
				},
				beacon.New(ethash.NewFaker()),
				func(i int, b *core.BlockGen) {
					b.SetPoS()
				},
			),
			opts...,
		)
		args := TransactionArgs{From: &from, To: &to}
		if caller != nil {
			args.AccessList = &caller
		}
		res, err := api.CreateAccessList(t.Context(), args, nil)
		require.NoErrorf(t, err, "%T.CreateAccessList()", api)
		require.Emptyf(t, res.Error, "%T.CreateAccessList() execution error", api)
		return *res.Accesslist
	}

	// Out of order and with duplicates, as storage keys that carry data may be,
	// and interleaved with another address.
	caller := func() types.AccessList {
		return types.AccessList{
			{Address: verbatim, StorageKeys: []common.Hash{{3}, {1}, {3}, {2}}},
			{Address: other, StorageKeys: []common.Hash{{9}}},
			{Address: verbatim, StorageKeys: []common.Hash{{5}, {5}}},
		}
	}
	opt := WithDefaultAccessListOptions(WithVerbatimAccessTuples(verbatim))

	t.Run("default", func(t *testing.T) {
		got := createAccessList(t, nil)
		assert.Equal(t, types.AccessList{{Address: verbatim, StorageKeys: []common.Hash{}}}, got, "generated tuple")

		got = createAccessList(t, caller())
		require.Len(t, got, 2)
		for _, tuple := range got {
			if tuple.Address == verbatim {
				assert.Len(t, tuple.StorageKeys, 4, "merged and deduplicated storage keys")
			}
		}
	})

	t.Run("verbatim", func(t *testing.T) {
		assert.Equal(t, types.AccessList{}, createAccessList(t, nil, opt), "generated tuple")

		want := caller()
		want = types.AccessList{want[1], want[0], want[2]}
		assert.Equal(t, want, createAccessList(t, caller(), opt), "caller tuples")
	})
}
