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

package rlp

import (
	"errors"
	"fmt"
	"io"
	"reflect"
)

// InList is a convenience wrapper, calling `fn` between calls to
// [EncoderBuffer.List] and [EncoderBuffer.ListEnd]. If `fn` returns an error,
// it is propagated directly.
func (b EncoderBuffer) InList(fn func() error) error {
	l := b.List()
	if err := fn(); err != nil {
		return err
	}
	b.ListEnd(l)
	return nil
}

// EncodeListToBuffer is equivalent to [Encode], writing the RLP encoding of
// each element to `b`, except that it wraps the writes inside a call to
// [EncoderBuffer.InList].
func EncodeListToBuffer[T any](b EncoderBuffer, vals []T) error {
	return b.InList(func() error {
		for _, v := range vals {
			if err := Encode(b, v); err != nil {
				return err
			}
		}
		return nil
	})
}

// EncodeStructFields encodes the required and optional slices to `w`,
// concatenated as a single list, as if they were fields in a struct. The
// optional "fields", which MAY be nil, are treated identically to those tagged
// with `rlp:"optional"`.
//
// See the example for [Stream.DecodeStructFields].
func EncodeStructFields(w io.Writer, required []any, opt *OptionalFields) error {
	includeOptional, err := opt.inclusionFlags()
	if err != nil {
		return err
	}

	b := NewEncoderBuffer(w)
	err = b.InList(func() error {
		for _, v := range required {
			if err := Encode(b, v); err != nil {
				return err
			}
		}

		for i, v := range opt.vals() {
			if !includeOptional[i] {
				return nil
			}
			if err := Encode(b, v); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return b.Flush()
}

// Optional returns the `vals` as [OptionalFields]; see the type's documentation
// for the resulting behaviour.
func Optional(vals ...any) *OptionalFields {
	return &OptionalFields{vals}
}

// OptionalFields are treated by [EncodeStructFields] and
// [Stream.DecodeStructFields] as if they were tagged with `rlp:"optional"`.
type OptionalFields struct {
	// Note that the [OptionalFields] type exists primarily to improve
	// readability at the call sites of [EncodeStructFields] and
	// [Stream.DecodeStructFields]. While an `[]any` slice would suffice, it
	// results in ambiguous usage of field functionality.

	v []any
}

// vals is a convenience wrapper, returning o.v, but allowing for a nil
// receiver, in which case it returns a nil slice.
func (o *OptionalFields) vals() []any {
	if o == nil {
		return nil
	}
	return o.v
}

var errUnsupportedOptionalFieldType = errors.New("unsupported optional field type")

// inclusionFlags returns a slice of booleans, the same length as `fs`,
// indicating whether or not the respective field MUST be written to a list. A
// field must be written if it or any later field value is non-nil; the returned
// slice is therefore monotonic non-increasing from true to false.
func (o *OptionalFields) inclusionFlags() ([]bool, error) {
	if o == nil {
		return nil, nil
	}

	flags := make([]bool, len(o.v))
	var include bool
	for i := len(o.v) - 1; i >= 0; i-- {
		switch v := reflect.ValueOf(o.v[i]); v.Kind() {
		case reflect.Slice, reflect.Pointer:
			include = include || !v.IsNil()
		default:
			return nil, fmt.Errorf("%w: %T", errUnsupportedOptionalFieldType, o.v[i])
		}
		flags[i] = include
	}
	return flags, nil
}

// FromList is a convenience wrapper, calling `fn` between calls to
// [Stream.List] and [Stream.ListEnd]. If `fn` returns an error, it is
// propagated directly.
func (s *Stream) FromList(fn func() error) error {
	if _, err := s.List(); err != nil {
		return err
	}
	if err := fn(); err != nil {
		return err
	}
	return s.ListEnd()
}

// DecodeList assumes that the next item in `s` is a list and decodes every item
// in said list to a `*T`.
//
// The returned slice is guaranteed to be non-nil, even if the list is empty.
// This is in keeping with other behaviour in this package and it is therefore
// the responsibility of callers to respect `rlp:"nil"` struct tags.
func DecodeList[T any](s *Stream) ([]*T, error) {
	vals := []*T{}
	err := s.FromList(func() error {
		for s.MoreDataInList() {
			var v T
			if err := s.Decode(&v); err != nil {
				return err
			}
			vals = append(vals, &v)
		}
		return nil
	})
	return vals, err
}

// DecodeStructFields is the inverse of [EncodeStructFields]. All destination
// fields, be they required or optional, MUST be pointers and all optional
// fields MUST be provided in case they are present in the RLP being decoded. If
// no optional fields exist, the argument MAY be nil.
//
// Typically, the arguments to this function mirror those passed to
// [EncodeStructFields] except for being pointers. See the example.
func (s *Stream) DecodeStructFields(required []any, opt *OptionalFields) error {
	return s.FromList(func() error {
		for _, v := range required {
			if err := s.Decode(v); err != nil {
				return err
			}
		}

		for _, v := range opt.vals() {
			if !s.MoreDataInList() {
				return nil
			}
			if err := s.Decode(v); err != nil {
				return err
			}
		}
		return nil
	})
}

// Nillable wraps `field` to mirror the behaviour of an `rlp:"nil"` tag; i.e. if
// a zero-sized RLP item is decoded into the returned Decoder then it is dropped
// and `*field` is set to nil, otherwise the RLP item is decoded directly into
// `field`. The return argument is intended for use with
// [Stream.DecodeStructFields].
func Nillable[T any](field **T) Decoder {
	return &nillable[T]{field}
}

type nillable[T any] struct{ v **T }

func (n *nillable[T]) DecodeRLP(s *Stream) error {
	_, size, err := s.Kind()
	if err != nil {
		return err
	}
	if size > 0 {
		return s.Decode(n.v)
	}
	*n.v = nil
	_, err = s.Raw() // consume the item
	return err
}
