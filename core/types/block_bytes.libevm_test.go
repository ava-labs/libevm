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
	"bytes"
	"fmt"
	"math/big"
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

type blockBytesImpl struct {
	name string
	fn   func(headerBytes, bodyBytes []byte) ([]byte, error)
}

// blockBytesImpls are all of the ways of converting header + body bytes into
// block bytes, all of which MUST be equivalent.
var blockBytesImpls = []blockBytesImpl{
	{"BlockBytes", BlockBytes},
	{"reencode", referenceBlockBytes},
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

// newTestHeaderAndBody returns a [Header] and a [Body] with pseudorandomly
// populated fields, the latter holding the number of items described by `size`.
func newTestHeaderAndBody(rng *ethtest.PseudoRand, size blockBodySize) (*Header, *Body) {
	newHeader := func() *Header {
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

	body := new(Body)
	for i := 0; i < size.txs; i++ {
		body.Transactions = append(body.Transactions, NewTx(&LegacyTx{
			Nonce:    rng.Uint64(),
			GasPrice: rng.BigUint64(),
			Gas:      rng.Uint64(),
			To:       rng.AddressPtr(),
			Value:    rng.BigUint64(),
			Data:     rng.Bytes(64),
		}))
	}
	for i := 0; i < size.uncles; i++ {
		body.Uncles = append(body.Uncles, newHeader())
	}
	for i := 0; i < size.withdrawals; i++ {
		body.Withdrawals = append(body.Withdrawals, &Withdrawal{
			Index:     rng.Uint64(),
			Validator: rng.Uint64(),
			Address:   rng.Address(),
			Amount:    rng.Uint64(),
		})
	}
	return newHeader(), body
}

// newTestHeaderAndBodyBytes returns the RLP encodings of [newTestHeaderAndBody].
func newTestHeaderAndBodyBytes(tb testing.TB, rng *ethtest.PseudoRand, size blockBodySize) (headerBytes, bodyBytes []byte) {
	tb.Helper()
	header, body := newTestHeaderAndBody(rng, size)
	return mustEncodeRLP(tb, header), mustEncodeRLP(tb, body)
}

func mustEncodeRLP(tb testing.TB, x any) []byte {
	tb.Helper()
	b, err := rlp.EncodeToBytes(x)
	require.NoErrorf(tb, err, "rlp.EncodeToBytes(%T)", x)
	return b
}

func TestBlockBytes(t *testing.T) {
	rng := ethtest.NewPseudoRand(0xC0FFEE)

	t.Run("equivalence", func(t *testing.T) {
		for _, size := range blockBodySizes {
			t.Run(size.String(), func(t *testing.T) {
				headerBytes, bodyBytes := newTestHeaderAndBodyBytes(t, rng, size)
				want, err := referenceBlockBytes(headerBytes, bodyBytes)
				require.NoError(t, err, "referenceBlockBytes()")

				for _, impl := range blockBytesImpls {
					got, err := impl.fn(headerBytes, bodyBytes)
					require.NoErrorf(t, err, "%s()", impl.name)
					assert.Equalf(t, want, got, "%s()", impl.name)
				}
			})
		}
	})

	t.Run("with_registered_extras", func(t *testing.T) {
		// The extra payload is part of the body's RLP so [BlockBytes] MUST
		// propagate it without any knowledge of it.
		TestOnlyClearRegisteredExtras()
		t.Cleanup(TestOnlyClearRegisteredExtras)
		extras := RegisterExtras[
			NOOPHeaderHooks, *NOOPHeaderHooks,
			rlpBodyPayload, *rlpBodyPayload,
			struct{},
		]()
		rlpBodyPayloads = extras.Body

		header, body := newTestHeaderAndBody(rng, blockBodySize{2, 1, 1})
		extras.Body.Set(body, &rlpBodyPayload{
			Version: 314159,
			Data:    rng.Bytes(8),
		})

		block := NewBlockWithHeader(header).
			WithBody(*body).
			WithWithdrawals(body.Withdrawals)
		want := mustEncodeRLP(t, block)

		got, err := BlockBytes(mustEncodeRLP(t, header), mustEncodeRLP(t, body))
		require.NoError(t, err, "BlockBytes()")
		assert.Equal(t, want, got)
	})

	t.Run("invalid_body", func(t *testing.T) {
		// Only the body's outermost list is parsed, so only it can be rejected.
		headerBytes, bodyBytes := newTestHeaderAndBodyBytes(t, rng, blockBodySize{1, 1, 1})

		tests := []struct {
			name      string
			bodyBytes []byte
		}{
			{"empty", nil},
			{"not_a_list", []byte{0x80}},
			{"truncated", bodyBytes[:len(bodyBytes)-1]},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := BlockBytes(headerBytes, tt.bodyBytes)
				assert.Error(t, err, "BlockBytes()")
			})
		}
	})

	t.Run("unvalidated_arguments", func(t *testing.T) {
		// [BlockBytes] documents that it doesn't validate its arguments beyond
		// the body's outermost list; these cases pin that behaviour, all of
		// which the decode-based implementation would instead reject.
		headerBytes, bodyBytes := newTestHeaderAndBodyBytes(t, rng, blockBodySize{1, 1, 1})

		tests := []struct {
			name                   string
			headerBytes, bodyBytes []byte
		}{
			{
				name:        "header_not_a_list",
				headerBytes: []byte{0x80},
				bodyBytes:   bodyBytes,
			},
			{
				name:        "trailing_bytes_after_header",
				headerBytes: append(bytes.Clone(headerBytes), 0),
				bodyBytes:   bodyBytes,
			},
			{
				name:        "trailing_bytes_after_body",
				headerBytes: headerBytes,
				bodyBytes:   append(bytes.Clone(bodyBytes), 0),
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := referenceBlockBytes(tt.headerBytes, tt.bodyBytes)
				require.Error(t, err, "referenceBlockBytes() [demonstrating that the argument is invalid]")

				got, err := BlockBytes(tt.headerBytes, tt.bodyBytes)
				require.NoError(t, err, "BlockBytes()")
				// The output is only guaranteed to be a well-formed RLP list.
				_, rest, err := rlp.SplitList(got)
				require.NoErrorf(t, err, "rlp.SplitList(BlockBytes(...) = %#x)", got)
				assert.Emptyf(t, rest, "bytes after RLP list returned by BlockBytes()")
			})
		}
	})
}

func FuzzBlockBytes(f *testing.F) {
	rng := ethtest.NewPseudoRand(20250806)
	for _, size := range blockBodySizes {
		headerBytes, bodyBytes := newTestHeaderAndBodyBytes(f, rng, size)
		f.Add(headerBytes, bodyBytes)
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

// FuzzBlockBytesOfEncodedFields complements [FuzzBlockBytes], which has to skip
// all inputs that aren't canonical RLP. Fuzzing the *fields* of a [Header] and a
// [Body] instead of their encodings guarantees canonical inputs so every one of
// them is compared against the reference implementation. The unbounded fields
// allow the fuzzer to explore all sizes of RLP length prefix.
func FuzzBlockBytesOfEncodedFields(f *testing.F) {
	f.Add([]byte("header extra"), []byte("tx data"), uint8(0), uint8(0), uint8(0), false, uint64(0))
	f.Add([]byte{}, []byte{}, uint8(3), uint8(2), uint8(4), true, ^uint64(0))

	f.Fuzz(func(
		t *testing.T,
		headerExtra, txData []byte,
		numTxs, numUncles, numWithdrawals uint8,
		emptyWithdrawals bool,
		scalar uint64,
	) {
		newHeader := func(extra []byte) *Header {
			return &Header{
				Extra:      extra,
				Difficulty: new(big.Int).SetUint64(scalar),
				Number:     new(big.Int).SetUint64(scalar),
				GasLimit:   scalar,
				Nonce:      EncodeNonce(scalar),
			}
		}

		body := new(Body)
		for i := 0; i < int(numTxs%4); i++ {
			body.Transactions = append(body.Transactions, NewTx(&LegacyTx{
				Nonce: scalar + uint64(i),
				Data:  txData,
			}))
		}
		for i := 0; i < int(numUncles%3); i++ {
			body.Uncles = append(body.Uncles, newHeader(headerExtra))
		}
		// A nil [Body.Withdrawals] is absent from the RLP, an empty but non-nil
		// one is included as an empty list, and both MUST be propagated.
		if emptyWithdrawals {
			body.Withdrawals = []*Withdrawal{}
		}
		for i := 0; i < int(numWithdrawals%5); i++ {
			body.Withdrawals = append(body.Withdrawals, &Withdrawal{
				Index:  scalar + uint64(i),
				Amount: scalar,
			})
		}

		headerBytes := mustEncodeRLP(t, newHeader(headerExtra))
		bodyBytes := mustEncodeRLP(t, body)

		want, err := referenceBlockBytes(headerBytes, bodyBytes)
		require.NoError(t, err, "referenceBlockBytes()")

		got, err := BlockBytes(headerBytes, bodyBytes)
		require.NoError(t, err, "BlockBytes()")
		assert.Equal(t, want, got, "BlockBytes() vs decoding, combining, and re-encoding")
	})
}

// skipIfNonCanonicalRLP skips the test if re-encoding `decoded`, which was
// decoded from `rlpBytes`, doesn't reproduce `rlpBytes` exactly.
func skipIfNonCanonicalRLP(t *testing.T, rlpBytes []byte, decoded any) {
	t.Helper()
	canonical, err := rlp.EncodeToBytes(decoded)
	if err != nil {
		t.Skipf("rlp.EncodeToBytes(%T) error %v", decoded, err)
	}
	if !bytes.Equal(rlpBytes, canonical) {
		t.Skipf("non-canonical RLP input for %T", decoded)
	}
}

func BenchmarkBlockBytes(b *testing.B) {
	for _, size := range blockBodySizes {
		rng := ethtest.NewPseudoRand(2718281828)
		headerBytes, bodyBytes := newTestHeaderAndBodyBytes(b, rng, size)
		b.Run(size.String(), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = BlockBytes(headerBytes, bodyBytes)
			}
		})
	}
}
