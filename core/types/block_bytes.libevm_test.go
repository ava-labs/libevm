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

package types_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "github.com/ava-labs/libevm/core/types"
	"github.com/ava-labs/libevm/libevm/ethtest"
	"github.com/ava-labs/libevm/rlp"
)

// referenceBlockBytes is the reference implementation against which
// [BlockBytes] is tested and benchmarked.
func referenceBlockBytes(headerBytes, bodyBytes []byte) ([]byte, error) {
	header := new(Header)
	if err := rlp.DecodeBytes(headerBytes, header); err != nil {
		return nil, err
	}
	body := new(Body)
	if err := rlp.DecodeBytes(bodyBytes, body); err != nil {
		return nil, err
	}
	block := NewBlockWithHeader(header).
		WithBody(*body).
		WithWithdrawals(body.Withdrawals)
	return rlp.EncodeToBytes(block)
}

// blockBodySize describes the [Body] contents of a test case.
type blockBodySize struct {
	txs, uncles, withdrawals int
}

func (s blockBodySize) String() string {
	return fmt.Sprintf("%d_txs_%d_uncles_%d_withdrawals", s.txs, s.uncles, s.withdrawals)
}

var blockBodySizes = []blockBodySize{
	{txs: 0, uncles: 0, withdrawals: 0},
	{txs: 1, uncles: 0, withdrawals: 0},
	{txs: 0, uncles: 1, withdrawals: 0},
	{txs: 0, uncles: 0, withdrawals: 1},
	{txs: 10, uncles: 0, withdrawals: 0},
	{txs: 10, uncles: 2, withdrawals: 4},
	{txs: 100, uncles: 0, withdrawals: 0},
	{txs: 100, uncles: 2, withdrawals: 16},
}

func newHeader(rng *ethtest.PseudoRand) *Header {
	return &Header{
		ParentHash:  rng.Hash(),
		UncleHash:   rng.Hash(),
		Coinbase:    rng.Address(),
		Root:        rng.Hash(),
		TxHash:      rng.Hash(),
		ReceiptHash: rng.Hash(),
		Bloom:       rng.Bloom(),
		Difficulty:  rng.BigUint64(),
		Number:      rng.BigUint64(),
		GasLimit:    rng.Uint64(),
		GasUsed:     rng.Uint64(),
		Time:        rng.Uint64(),
		Extra:       rng.Bytes(32),
		MixDigest:   rng.Hash(),
		Nonce:       rng.BlockNonce(),
		BaseFee:     rng.BigUint64(),
	}
}

// newTestHeaderAndBody returns a [Header] and a [Body] with pseudorandomly
// populated fields, the latter holding the number of items described by `size`.
func newBody(rng *ethtest.PseudoRand, size blockBodySize) *Body {
	body := &Body{
		Transactions: make([]*Transaction, size.txs),
		Uncles:       make([]*Header, size.uncles),
		Withdrawals:  make([]*Withdrawal, size.withdrawals),
	}
	for i := range size.txs {
		body.Transactions[i] = NewTx(&LegacyTx{
			Nonce:    rng.Uint64(),
			GasPrice: rng.BigUint64(),
			Gas:      rng.Uint64(),
			To:       rng.AddressPtr(),
			Value:    rng.BigUint64(),
			Data:     rng.Bytes(64),
		})
	}
	for i := range size.uncles {
		body.Uncles[i] = newHeader(rng)
	}
	for i := range size.withdrawals {
		body.Withdrawals[i] = &Withdrawal{
			Index:     rng.Uint64(),
			Validator: rng.Uint64(),
			Address:   rng.Address(),
			Amount:    rng.Uint64(),
		}
	}
	return body
}

func encodeRLP(tb testing.TB, x any) []byte {
	tb.Helper()
	b, err := rlp.EncodeToBytes(x)
	require.NoErrorf(tb, err, "rlp.EncodeToBytes(%T)", x)
	return b
}

func FuzzBlockBytes(f *testing.F) {
	rng := ethtest.NewPseudoRand(20250806)
	for _, size := range blockBodySizes {
		f.Add(
			encodeRLP(f, newHeader(rng)),
			encodeRLP(f, newBody(rng, size)),
		)
	}

	f.Fuzz(func(t *testing.T, headerBytes, bodyBytes []byte) {
		want, err := referenceBlockBytes(headerBytes, bodyBytes)
		if err != nil {
			t.Skip("invalid input bytes")
		}

		got, err := BlockBytes(headerBytes, bodyBytes)
		require.NoError(t, err, "BlockBytes()")
		assert.Equal(t, want, got, "referenceBlockBytes() == BlockBytes()")
	})
}

func BenchmarkBlockBytes(b *testing.B) {
	for _, size := range blockBodySizes {
		rng := ethtest.NewPseudoRand(2718281828)
		headerBytes := encodeRLP(b, newHeader(rng))
		bodyBytes := encodeRLP(b, newBody(rng, size))
		b.Run(size.String(), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = referenceBlockBytes(headerBytes, bodyBytes)
			}
		})
	}
}
