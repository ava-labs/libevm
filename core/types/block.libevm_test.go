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
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "github.com/ava-labs/libevm/core/types"
	"github.com/ava-labs/libevm/crypto"
	"github.com/ava-labs/libevm/libevm/ethtest"
	"github.com/ava-labs/libevm/rlp"
)

type stubHeaderHooks struct {
	suffix                                   []byte
	gotRawJSONToUnmarshal, gotRawRLPToDecode []byte
	setHeaderToOnUnmarshalOrDecode           Header

	errMarshal, errUnmarshal, errEncode, errDecode error
}

func fakeHeaderJSON(h *Header, suffix []byte) []byte {
	return []byte(fmt.Sprintf(`"%#x:%#x"`, h.ParentHash, suffix))
}

func fakeHeaderRLP(h *Header, suffix []byte) []byte {
	return append(crypto.Keccak256(h.ParentHash[:]), suffix...)
}

func (hh *stubHeaderHooks) MarshalJSON(h *Header) ([]byte, error) { //nolint:govet
	return fakeHeaderJSON(h, hh.suffix), hh.errMarshal
}

func (hh *stubHeaderHooks) UnmarshalJSON(h *Header, b []byte) error { //nolint:govet
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
