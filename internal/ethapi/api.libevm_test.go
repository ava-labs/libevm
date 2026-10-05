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
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"

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

	api := NewBlockChainAPI(newTestBackend(
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
		}),
	)

	var got []byte
	retErr := errors.New("intercepted!")
	opt := WithCallResultInterceptor(func(r *core.ExecutionResult) error {
		got = slices.Clone(r.ReturnData)
		return retErr
	})

	_, err := api.Call(t.Context(), TransactionArgs{To: &addr}, nil, nil, nil, opt)
	assert.Equalf(t, retErr, err, "%T.Call() error propagated from CallResultInterceptor", api)
	assert.Lenf(t, got, returnBufSize, "%T.ReturnData received by CallResultInterceptor", &core.ExecutionResult{})
}
