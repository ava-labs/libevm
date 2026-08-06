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

package types

import (
	"encoding/json"
	"io"

	"github.com/ava-labs/libevm/internal/libevm/pseudo"
	"github.com/ava-labs/libevm/rlp"
)

// HeaderHooks are required for all types registered with [RegisterExtras] for
// [Header] payloads.
type HeaderHooks interface {
	EncodeJSON(*Header) ([]byte, error)
	DecodeJSON(*Header, []byte) error
	EncodeRLP(*Header, io.Writer) error
	DecodeRLP(*Header, *rlp.Stream) error
	PostCopy(dst *Header)
	PostRPCMarshal(h *Header, marshalled map[string]any)
}

var _ interface {
	rlp.Encoder
	rlp.Decoder
	json.Marshaler
	json.Unmarshaler
} = (*Header)(nil)

// MarshalJSON implements the [json.Marshaler] interface.
func (h *Header) MarshalJSON() ([]byte, error) {
	return h.hooks().EncodeJSON(h)
}

// UnmarshalJSON implements the [json.Unmarshaler] interface.
func (h *Header) UnmarshalJSON(b []byte) error {
	return h.hooks().DecodeJSON(h, b)
}

// EncodeRLP implements the [rlp.Encoder] interface.
func (h *Header) EncodeRLP(w io.Writer) error {
	return h.hooks().EncodeRLP(h, w)
}

// DecodeRLP implements the [rlp.Decoder] interface.
func (h *Header) DecodeRLP(s *rlp.Stream) error {
	return h.hooks().DecodeRLP(h, s)
}

// NOOPHeaderHooks implements [HeaderHooks] such that they are equivalent to
// no type having been registered.
type NOOPHeaderHooks struct{}

var _ HeaderHooks = (*NOOPHeaderHooks)(nil)

func (*NOOPHeaderHooks) EncodeJSON(h *Header) ([]byte, error) {
	return h.marshalJSON()
}

func (*NOOPHeaderHooks) DecodeJSON(h *Header, b []byte) error {
	return h.unmarshalJSON(b)
}

func (*NOOPHeaderHooks) EncodeRLP(h *Header, w io.Writer) error {
	return h.encodeRLP(w)
}

func (*NOOPHeaderHooks) DecodeRLP(h *Header, s *rlp.Stream) error {
	type withoutMethods Header
	return s.Decode((*withoutMethods)(h))
}
func (*NOOPHeaderHooks) PostCopy(dst *Header) {}

func (*NOOPHeaderHooks) PostRPCMarshal(*Header, map[string]any) {}

var _ = []interface {
	rlp.Encoder
	rlp.Decoder
}{
	(*Body)(nil),
	(*extblock)(nil),
}

// EncodeRLP implements the [rlp.Encoder] interface.
func (b *Body) EncodeRLP(w io.Writer) error {
	return b.hooks().BodyRLPFieldsForEncoding(b).EncodeRLP(w)
}

// DecodeRLP implements the [rlp.Decoder] interface.
func (b *Body) DecodeRLP(s *rlp.Stream) error {
	return b.hooks().BodyRLPFieldPointersForDecoding(b).DecodeRLP(s)
}

// A [Block] is RLP-encoded as its [Header] followed by the fields of its
// [Body], hence the [extblock] {en,de}coding reuses the [BlockBodyHooks] body
// methods, prefixing the header to the required fields.

// body returns a [Body] carrying the fields of `b`. Although it is only a
// temporary carrier, it shares the [Block] extra payload and is therefore
// indistinguishable, from the perspective of the hooks, from a Body being
// {en,de}coded directly; in particular, hooks MAY access their own payload via
// the Body they are passed.
func (b *extblock) body() *Body {
	return &Body{
		Transactions: b.Txs,
		Uncles:       b.Uncles,
		Withdrawals:  b.Withdrawals,
		extra:        b.extra,
	}
}

func (b *extblock) EncodeRLP(w io.Writer) error {
	body := b.body()
	fields := body.hooks().BodyRLPFieldsForEncoding(body)
	fields.Required = append(
		[]any{b.Header},
		fields.Required...,
	)
	return fields.EncodeRLP(w)
}

func (b *extblock) DecodeRLP(s *rlp.Stream) error {
	// Unlike when encoding, the geth fields are the destination of the decoding
	// so are only copied from the [Body] afterwards. See [extblock.body] re the
	// extra payload.
	body := &Body{extra: b.extra}
	fields := body.hooks().BodyRLPFieldPointersForDecoding(body)
	fields.Required = append(
		[]any{&b.Header},
		fields.Required...,
	)
	if err := fields.DecodeRLP(s); err != nil {
		return err
	}
	b.Txs = body.Transactions
	b.Uncles = body.Uncles
	b.Withdrawals = body.Withdrawals
	return nil
}

// BlockBodyHooks are required for all types registered with [RegisterExtras]
// for [Block] and [Body] payloads. The same methods are used for both [Block]
// and [Body] {en,de}coding as a Block is encoded as its [Header] followed by
// the fields of its [Body].
type BlockBodyHooks interface {
	BodyRLPFieldsForEncoding(*Body) *rlp.Fields
	BodyRLPFieldPointersForDecoding(*Body) *rlp.Fields
	PostRPCMarshal(b *Block, marshalled map[string]any)
}

// NOOPBlockBodyHooks implements [BlockBodyHooks] such that they are equivalent
// to no type having been registered.
type NOOPBlockBodyHooks struct{}

var _ BlockBodyPayload[*NOOPBlockBodyHooks] = (*NOOPBlockBodyHooks)(nil)

func (NOOPBlockBodyHooks) Copy() *NOOPBlockBodyHooks { return &NOOPBlockBodyHooks{} }

// The RLP-related methods of [NOOPBlockBodyHooks] make assumptions about the
// struct fields and their order, which we lock in here as a change detector. If
// these break then they MUST be updated and the RLP methods reviewed + new
// backwards-compatibility tests added.
var (
	_ = &Body{
		[]*Transaction{}, []*Header{}, []*Withdrawal{}, // geth
		&pseudo.Type{}, // libevm
	}
	_ = extblock{
		&Header{}, []*Transaction{}, []*Header{}, []*Withdrawal{}, // geth
		&pseudo.Type{}, // libevm
	}
)

func (NOOPBlockBodyHooks) BodyRLPFieldsForEncoding(b *Body) *rlp.Fields {
	return &rlp.Fields{
		Required: []any{b.Transactions, b.Uncles},
		Optional: []any{b.Withdrawals},
	}
}

func (NOOPBlockBodyHooks) BodyRLPFieldPointersForDecoding(b *Body) *rlp.Fields {
	return &rlp.Fields{
		Required: []any{&b.Transactions, &b.Uncles},
		Optional: []any{&b.Withdrawals},
	}
}

func (NOOPBlockBodyHooks) PostRPCMarshal(*Block, map[string]any) {}
