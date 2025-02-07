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

package types_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ava-labs/libevm/common"
	. "github.com/ava-labs/libevm/core/types"
	"github.com/ava-labs/libevm/libevm/ethtest"
	"github.com/ava-labs/libevm/libevm/pseudo"
	"github.com/ava-labs/libevm/rlp"
)

type stubHeaderHooks struct {
	suffix                 []byte
	gotRawJSONToUnmarshal  []byte
	setHeaderToOnUnmarshal Header
	accessor               pseudo.Accessor[*Header, *stubHeaderHooks]
	toCopy                 *stubHeaderHooks

	errMarshal, errUnmarshal error
}

func fakeHeaderJSON(h *Header, suffix []byte) []byte {
	return []byte(fmt.Sprintf(`"%#x:%#x"`, h.ParentHash, suffix))
}

func (hh *stubHeaderHooks) MarshalJSON(h *Header) ([]byte, error) { //nolint:govet
	return fakeHeaderJSON(h, hh.suffix), hh.errMarshal
}

func (hh *stubHeaderHooks) UnmarshalJSON(h *Header, b []byte) error { //nolint:govet
	hh.gotRawJSONToUnmarshal = b
	*h = hh.setHeaderToOnUnmarshal
	return hh.errUnmarshal
}

func directEncodeHeaderRLP(tb testing.TB, h *Header, extraPayloadSuffix []byte) []byte {
	tb.Helper()

	// The encoded type mirrors the fields returned by
	// [stubHeaderHooks.RLPFieldsForEncoding].
	buf, err := rlp.EncodeToBytes(struct {
		ParentHash common.Hash
		Suffix     []byte
	}{h.ParentHash, extraPayloadSuffix})

	require.NoError(tb, err)
	return buf
}

func (hh *stubHeaderHooks) RLPFieldsForEncoding(h *Header) *rlp.Fields {
	return &rlp.Fields{Required: []any{h.ParentHash, hh.suffix}}
}

func (hh *stubHeaderHooks) RLPFieldPointersForDecoding(h *Header) *rlp.Fields {
	return &rlp.Fields{Required: []any{&h.ParentHash, &hh.suffix}}
}

func (hh *stubHeaderHooks) PostCopy(dst *Header) {
	hh.accessor.Set(dst, hh.toCopy)
}

func TestHeaderHooks(t *testing.T) {
	TestOnlyClearRegisteredExtras()
	defer TestOnlyClearRegisteredExtras()

	extras := RegisterExtras[stubHeaderHooks, *stubHeaderHooks, struct{}]()
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
			setHeaderToOnUnmarshal: Header{
				Extra: []byte("can you solve this puzzle? 0xbda01b6cf56c303bd3f581599c0d5c0b"),
			},
		}
		extras.Header.Set(hdr, stub)

		input := fmt.Sprintf("%q", "hello, JSON world")
		err := json.Unmarshal([]byte(input), hdr)
		require.NoErrorf(t, err, "json.Unmarshal()")

		assert.Equal(t, input, string(stub.gotRawJSONToUnmarshal), "raw JSON received by hook")
		assert.Equal(t, &stub.setHeaderToOnUnmarshal, hdr, "%T after JSON unmarshalling with hook", hdr)
	})

	t.Run("EncodeRLP", func(t *testing.T) {
		got, err := rlp.EncodeToBytes(hdr)
		require.NoError(t, err, "rlp.EncodeToBytes(%T)", hdr)
		assert.Equal(t, directEncodeHeaderRLP(t, hdr, suffix), got)
	})

	t.Run("DecodeRLP", func(t *testing.T) {
		input := directEncodeHeaderRLP(t, hdr, suffix)

		got := new(Header)
		stub := &stubHeaderHooks{}
		extras.Header.Set(got, stub)
		err := rlp.DecodeBytes(input, got)
		require.NoErrorf(t, err, "rlp.DecodeBytes(%#x)", input)

		assert.Equalf(t, hdr.ParentHash, got.ParentHash, "RLP-decoded %T.ParentHash", hdr)
		assert.Equalf(t, suffix, stub.suffix, "RLP-decoded %T.suffix", stub)
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

		hdr := new(Header)
		setStub := func() {
			extras.Header.Set(hdr, &stubHeaderHooks{
				errMarshal:   errMarshal,
				errUnmarshal: errUnmarshal,
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
	})
}
