// Copyright 2025 the libevm authors.
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

package parallel

import (
	"crypto/sha256"
	"encoding/binary"
	"math/rand/v2"
	"testing"

	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/libevm/core"
	"github.com/ava-labs/libevm/core/types"
	"github.com/ava-labs/libevm/trie"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m, goleak.IgnoreCurrent())
}

type shaHandler struct {
	addr common.Address
	gas  uint64
}

func (h *shaHandler) Gas(tx *types.Transaction) (uint64, bool) {
	if to := tx.To(); to == nil || *to != h.addr {
		return 0, false
	}
	return h.gas, true
}

func (*shaHandler) Process(i int, tx *types.Transaction) [sha256.Size]byte {
	return sha256.Sum256(tx.Data())
}

func TestProcessor(t *testing.T) {
	handler := &shaHandler{
		addr: common.Address{'s', 'h', 'a', 2, 5, 6},
		gas:  1e6,
	}
	p := New(handler, 8)
	t.Cleanup(p.Close)

	type blockParams struct {
		numTxs                              int
		sendToAddrEvery, sufficientGasEvery int
	}

	// Each set of params is effectively a test case, but they are all run on
	// the same [Processor].
	params := []blockParams{
		{
			numTxs: 0,
		},
		{
			numTxs:             500,
			sendToAddrEvery:    7,
			sufficientGasEvery: 5,
		},
		{
			numTxs:             1_000,
			sendToAddrEvery:    7,
			sufficientGasEvery: 5,
		},
		{
			numTxs:             1_000,
			sendToAddrEvery:    11,
			sufficientGasEvery: 3,
		},
		{
			numTxs:             100,
			sendToAddrEvery:    1,
			sufficientGasEvery: 1,
		},
		{
			numTxs: 0,
		},
	}

	rng := rand.New(rand.NewPCG(0, 0))
	for range 100 {
		params = append(params, blockParams{
			numTxs:             rng.IntN(1000),
			sendToAddrEvery:    1 + rng.IntN(30),
			sufficientGasEvery: 1 + rng.IntN(30),
		})
	}

	for _, tc := range params {
		t.Run("", func(t *testing.T) {
			t.Logf("%+v", tc)

			txs := make(types.Transactions, tc.numTxs)
			wantProcessed := make([]bool, tc.numTxs)
			for i := range len(txs) {
				var (
					to       common.Address
					extraGas uint64
				)

				wantProcessed[i] = true
				if i%tc.sendToAddrEvery == 0 {
					to = handler.addr
				} else {
					wantProcessed[i] = false
				}
				if i%tc.sufficientGasEvery == 0 {
					extraGas = handler.gas
				} else {
					wantProcessed[i] = false
				}

				data := binary.BigEndian.AppendUint64(nil, uint64(i))
				gas, err := core.IntrinsicGas(data, nil, false, true, true, true)
				require.NoError(t, err, "core.IntrinsicGas(%#x, nil, false, true, true, true)", data)

				txs[i] = types.NewTx(&types.LegacyTx{
					To:   &to,
					Data: data,
					Gas:  gas + extraGas,
				})
			}

			block := types.NewBlock(&types.Header{}, txs, nil, nil, trie.NewStackTrie(nil))
			require.NoError(t, p.StartBlock(block), "BeforeBlock()")
			defer p.FinishBlock(block)

			for i, tx := range txs {
				wantOK := wantProcessed[i]

				var want [sha256.Size]byte
				if wantOK {
					want = handler.Process(i, tx)
				}

				got, gotOK := p.Result(i)
				if got != want || gotOK != wantOK {
					t.Errorf("Result(%d) got (%#x, %t); want (%#x, %t)", i, got, gotOK, want, wantOK)
				}
			}
		})

		if t.Failed() {
			break
		}
	}
}
