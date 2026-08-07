// Copyright 2024-2025 the libevm authors.
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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "github.com/ava-labs/libevm/core/types"
	"github.com/ava-labs/libevm/crypto"
	"github.com/ava-labs/libevm/internal/libevm/pseudo"
	"github.com/ava-labs/libevm/libevm/ethtest"
	"github.com/ava-labs/libevm/rlp"
	"github.com/ava-labs/libevm/trie"
)

type stubHeaderHooks struct {
	suffix                                   []byte
	gotRawJSONToUnmarshal, gotRawRLPToDecode []byte
	setHeaderToOnUnmarshalOrDecode           Header
	accessor                                 pseudo.Accessor[*Header, *stubHeaderHooks]
	toCopy                                   *stubHeaderHooks

	errMarshal, errUnmarshal, errEncode, errDecode error

	NOOPHeaderHooks
}

func fakeHeaderJSON(h *Header, suffix []byte) []byte {
	return []byte(fmt.Sprintf(`"%#x:%#x"`, h.ParentHash, suffix))
}

func fakeHeaderRLP(h *Header, suffix []byte) []byte {
	return append(crypto.Keccak256(h.ParentHash[:]), suffix...)
}

func (hh *stubHeaderHooks) EncodeJSON(h *Header) ([]byte, error) {
	return fakeHeaderJSON(h, hh.suffix), hh.errMarshal
}

func (hh *stubHeaderHooks) DecodeJSON(h *Header, b []byte) error {
	hh.gotRawJSONToUnmarshal = b
	*h = hh.setHeaderToOnUnmarshalOrDecode
	return hh.errUnmarshal
}

func (hh *stubHeaderHooks) EncodeRLP(h *Header, w io.Writer) error {
	if _, err := w.Write(fakeHeaderRLP(h, hh.suffix)); err != nil {
		return err
	}
	return hh.errEncode
}

func (hh *stubHeaderHooks) DecodeRLP(h *Header, s *rlp.Stream) error {
	r, err := s.Raw()
	if err != nil {
		return err
	}
	hh.gotRawRLPToDecode = r
	*h = hh.setHeaderToOnUnmarshalOrDecode
	return hh.errDecode
}

func (hh *stubHeaderHooks) PostCopy(dst *Header) {
	hh.accessor.Set(dst, hh.toCopy)
}

func TestHeaderHooks(t *testing.T) {
	TestOnlyClearRegisteredExtras()
	defer TestOnlyClearRegisteredExtras()

	extras := RegisterExtras[
		stubHeaderHooks, *stubHeaderHooks,
		NOOPBlockBodyHooks, *NOOPBlockBodyHooks,
		struct{},
	]()
	rng := ethtest.NewPseudoRand(13579)

	suffix := rng.Bytes(8)
	hdr := &Header{
		ParentHash: rng.Hash(),
	}
	extras.Header.Get(hdr).suffix = append([]byte{}, suffix...)

	t.Run("MarshalJSON", func(t *testing.T) {
		got, err := json.Marshal(hdr)
		require.NoError(t, err, "json.Marshal(%T)", hdr)
		assert.Equal(t, fakeHeaderJSON(hdr, suffix), got)
	})

	t.Run("UnmarshalJSON", func(t *testing.T) {
		hdr := new(Header)
		stub := &stubHeaderHooks{
			setHeaderToOnUnmarshalOrDecode: Header{
				Extra: []byte("can you solve this puzzle? 0xbda01b6cf56c303bd3f581599c0d5c0b"),
			},
		}
		extras.Header.Set(hdr, stub)

		input := fmt.Sprintf("%q", "hello, JSON world")
		err := json.Unmarshal([]byte(input), hdr)
		require.NoErrorf(t, err, "json.Unmarshal()")

		assert.Equal(t, input, string(stub.gotRawJSONToUnmarshal), "raw JSON received by hook")
		assert.Equal(t, &stub.setHeaderToOnUnmarshalOrDecode, hdr, "%T after JSON unmarshalling with hook", hdr)
	})

	t.Run("EncodeRLP", func(t *testing.T) {
		got, err := rlp.EncodeToBytes(hdr)
		require.NoError(t, err, "rlp.EncodeToBytes(%T)", hdr)
		assert.Equal(t, fakeHeaderRLP(hdr, suffix), got)
	})

	t.Run("DecodeRLP", func(t *testing.T) {
		input, err := rlp.EncodeToBytes(rng.Bytes(8))
		require.NoError(t, err)

		hdr := new(Header)
		stub := &stubHeaderHooks{
			setHeaderToOnUnmarshalOrDecode: Header{
				Extra: []byte("arr4n was here"),
			},
		}
		extras.Header.Set(hdr, stub)
		err = rlp.DecodeBytes(input, hdr)
		require.NoErrorf(t, err, "rlp.DecodeBytes(%#x)", input)

		assert.Equal(t, input, stub.gotRawRLPToDecode, "raw RLP received by hooks")
		assert.Equalf(t, &stub.setHeaderToOnUnmarshalOrDecode, hdr, "%T after RLP decoding with hook", hdr)
	})

	t.Run("PostCopy", func(t *testing.T) {
		hdr := new(Header)
		stub := &stubHeaderHooks{
			accessor: extras.Header,
			toCopy: &stubHeaderHooks{
				suffix: []byte("copied"),
			},
		}
		extras.Header.Set(hdr, stub)

		got := extras.Header.Get(CopyHeader(hdr))
		assert.Equal(t, stub.toCopy, got)
	})

	t.Run("error_propagation", func(t *testing.T) {
		errMarshal := errors.New("whoops")
		errUnmarshal := errors.New("is it broken?")
		errEncode := errors.New("uh oh")
		errDecode := errors.New("something bad happened")

		hdr := new(Header)
		setStub := func() {
			extras.Header.Set(hdr, &stubHeaderHooks{
				errMarshal:   errMarshal,
				errUnmarshal: errUnmarshal,
				errEncode:    errEncode,
				errDecode:    errDecode,
			})
		}

		setStub()
		// The { } blocks are defensive, avoiding accidentally having the wrong
		// error checked in a future refactor. The verbosity is acceptable for
		// clarity in tests.
		{
			_, err := json.Marshal(hdr)
			assert.ErrorIs(t, err, errMarshal, "via json.Marshal()") //nolint:testifylint // require is inappropriate here as we wish to keep going
		}
		{
			err := json.Unmarshal([]byte("{}"), hdr)
			assert.Equal(t, errUnmarshal, err, "via json.Unmarshal()")
		}

		setStub() // [stubHeaderHooks] completely overrides the Header
		{
			err := rlp.Encode(io.Discard, hdr)
			assert.Equal(t, errEncode, err, "via rlp.Encode()")
		}
		{
			err := rlp.DecodeBytes([]byte{0}, hdr)
			assert.Equal(t, errDecode, err, "via rlp.DecodeBytes()")
		}
	})
}

type blockPayload struct {
	NOOPBlockBodyHooks
	x int
}

func (p *blockPayload) Copy() *blockPayload {
	return &blockPayload{x: p.x}
}

func TestBlockWithX(t *testing.T) {
	TestOnlyClearRegisteredExtras()
	t.Cleanup(TestOnlyClearRegisteredExtras)

	extras := RegisterExtras[
		NOOPHeaderHooks, *NOOPHeaderHooks,
		blockPayload, *blockPayload,
		struct{},
	]()

	typ := reflect.TypeOf(&Block{})
	for i := 0; i < typ.NumMethod(); i++ {
		method := typ.Method(i).Name
		if method == "Withdrawals" || !strings.HasPrefix(method, "With") {
			continue
		}

		block := NewBlockWithHeader(&Header{})
		const initialPayload = int(42)
		payload := &blockPayload{
			x: initialPayload,
		}
		extras.Block.Set(block, payload)

		t.Run(method, func(t *testing.T) {
			var newBlock *Block

			switch method {
			case "WithBody":
				var body Body
				extras.Body.Set(&body, payload)
				newBlock = block.WithBody(body)
			case "WithSeal":
				newBlock = block.WithSeal(&Header{})
			case "WithWithdrawals":
				newBlock = block.WithWithdrawals(nil)
			default:
				t.Fatalf("method call not implemented: %s", method)
			}

			payload.x++
			// This specifically uses `require` instead of `assert` because a
			// failure here invalidates the next test, which demonstrates a deep
			// copy.
			require.Equalf(t, initialPayload+1, extras.Block.Get(block).x, "%T payload %T after modification via pointer", block, payload)

			switch got := extras.Block.Get(newBlock); got.x {
			case initialPayload: // expected
			case 0:
				t.Errorf("%T payload %T got zero value; the payload was probably not copied, resulting in a default being created", newBlock, got)
			case initialPayload + 1:
				t.Errorf("%T payload %T got same value as modified original; the payload was probably shallow copied", newBlock, got)
			default:
				t.Errorf("%T payload %T got %d, want %d; this is unexpected even as an error so you're on your own here", newBlock, got, got.x, initialPayload)
			}
		})
	}
}

// bodyPayload is a [types.BlockBodyHooks] implementation carrying an extra
// field that is {en,de}coded as if it was a regular RLP field of both the
// [types.Block] and [types.Body].
type bodyPayload struct {
	Data []byte

	NOOPBlockBodyHooks
}

func (p *bodyPayload) Copy() *bodyPayload {
	return &bodyPayload{
		Data: slices.Clone(p.Data),
	}
}

var bodyPayloads pseudo.Accessor[*Body, *bodyPayload]

func (*bodyPayload) BodyRLPFieldsForEncoding(b *Body) *rlp.Fields {
	// Rather than using the receiver directly, we access it through b. This
	// demonstrates that the hooks can access their own payload via the
	// [types.Body] they are passed.
	p := bodyPayloads.Get(b)
	return &rlp.Fields{
		Required: []any{b.Transactions, b.Uncles, p.Data},
		Optional: []any{b.Withdrawals},
	}
}

func (*bodyPayload) BodyRLPFieldPointersForDecoding(b *Body) *rlp.Fields {
	// See above comment on why we access the receiver through b rather than
	// directly.
	p := bodyPayloads.Get(b)
	return &rlp.Fields{
		Required: []any{&b.Transactions, &b.Uncles, &p.Data},
		Optional: []any{&b.Withdrawals},
	}
}

// TestBodyExtraRoundTrip demonstrates that the extra from the method receiver
// is the same as the extra from the argument for [types.BlockBodyHooks]
// functions.
func TestBodyExtraRoundTrip(t *testing.T) {
	TestOnlyClearRegisteredExtras()
	t.Cleanup(TestOnlyClearRegisteredExtras)

	extras := RegisterExtras[
		NOOPHeaderHooks, *NOOPHeaderHooks,
		bodyPayload, *bodyPayload,
		struct{},
	]()
	bodyPayloads = extras.Body

	rng := ethtest.NewPseudoRand(142857)
	wantBlock := NewBlock(
		newHeader(rng),
		[]*Transaction{
			newTx(rng),
		},
		[]*Header{
			newHeader(rng),
		},
		nil, // Receipts don't meaningfully impact the RLP of the block.
		trie.NewStackTrie(nil),
	)
	want := extras.Block.Get(wantBlock)
	want.Data = rng.Bytes(8)

	b, err := rlp.EncodeToBytes(wantBlock)
	require.NoErrorf(t, err, "rlp.EncodeToBytes(%T)", wantBlock)

	gotBlock := new(Block)
	require.NoErrorf(t, rlp.DecodeBytes(b, gotBlock), "rlp.DecodeBytes(rlp.EncodeToBytes(%T), %T)", wantBlock, gotBlock)
	got := extras.Block.Get(gotBlock)
	assert.Equalf(t, want, got, "%T payload after RLP round trip", got)
}

// newHeader returns a [Header] with randomly populated fields.
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

// newTx returns a [Transaction] with randomly populated fields.
func newTx(rng *ethtest.PseudoRand) *Transaction {
	return NewTx(&LegacyTx{
		Nonce:    rng.Uint64(),
		GasPrice: rng.BigUint64(),
		Gas:      rng.Uint64(),
		To:       rng.AddressPtr(),
		Value:    rng.BigUint64(),
		Data:     rng.Bytes(64),
	})
}

// bodySize describes the number of items in the [Body] of a test case.
type bodySize struct {
	txs, uncles, withdrawals int
}

// bodySizes are the [Body] shapes covered by [FuzzBlockBytes] seeds and by
// [BenchmarkBlockBytes]. They cover every combination of empty and non-empty
// fields, the last of which is optional in RLP.
var bodySizes = []bodySize{
	{txs: 0, uncles: 0, withdrawals: 0},
	{txs: 1, uncles: 0, withdrawals: 0},
	{txs: 0, uncles: 1, withdrawals: 0},
	{txs: 0, uncles: 0, withdrawals: 1},
	{txs: 10, uncles: 0, withdrawals: 0},
	{txs: 10, uncles: 2, withdrawals: 4},
	{txs: 100, uncles: 0, withdrawals: 0},
	{txs: 100, uncles: 2, withdrawals: 16},
}

// newBody returns a [Body] with randomly populated fields, holding the number
// of items described by `size`.
func newBody(rng *ethtest.PseudoRand, size bodySize) *Body {
	body := &Body{}
	for range size.txs {
		body.Transactions = append(body.Transactions, newTx(rng))
	}
	for range size.uncles {
		body.Uncles = append(body.Uncles, newHeader(rng))
	}
	for range size.withdrawals {
		body.Withdrawals = append(body.Withdrawals, &Withdrawal{
			Index:     rng.Uint64(),
			Validator: rng.Uint64(),
			Address:   rng.Address(),
			Amount:    rng.Uint64(),
		})
	}
	return body
}

// encodeRLP RLP-encodes `v`, failing the test if it can't be encoded.
func encodeRLP(tb testing.TB, v any) []byte {
	tb.Helper()
	b, err := rlp.EncodeToBytes(v)
	require.NoErrorf(tb, err, "rlp.EncodeToBytes(%T)", v)
	return b
}

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

// FuzzBlockBytes demonstrates that [BlockBytes] is equivalent to
// [referenceBlockBytes] for all inputs that the latter accepts. The seed corpus
// covers every shape in [bodySizes].
func FuzzBlockBytes(f *testing.F) {
	rng := ethtest.NewPseudoRand(20250806)
	for _, size := range bodySizes {
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
	for _, size := range bodySizes {
		rng := ethtest.NewPseudoRand(2718281828)
		headerBytes := encodeRLP(b, newHeader(rng))
		bodyBytes := encodeRLP(b, newBody(rng, size))
		b.Run(fmt.Sprintf("%d_txs_%d_uncles_%d_withdrawals", size.txs, size.uncles, size.withdrawals), func(b *testing.B) {
			for b.Loop() {
				_, _ = BlockBytes(headerBytes, bodyBytes)
			}
		})
	}
}
