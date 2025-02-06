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

// EncodeStructFields encodes the `required` and `optional` slices to `w`,
// concatenated as a single list, as if they were fields in a struct. The
// optional "fields" are treated identically to those tagged with
// `rlp:"optional"`.
//
// See the example for [Stream.DecodeStructFields].
func EncodeStructFields(w io.Writer, required, optional []any) error {
	includeOptional, err := optionalFieldInclusionFlags(optional)
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

		for i, v := range optional {
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

var errUnsupportedOptionalFieldType = errors.New("unsupported optional field type")

// optionalFieldInclusionFlags returns a slice of booleans, the same length as
// `vals`, indicating whether or not the respective optional value MUST be
// written to a list. A value must be written if it or any later value is
// non-nil; the returned slice is therefore monotonic non-increasing from true
// to false.
func optionalFieldInclusionFlags(vals []any) ([]bool, error) {
	flags := make([]bool, len(vals))
	var include bool
	for i := len(vals) - 1; i >= 0; i-- {
		switch v := reflect.ValueOf(vals[i]); v.Kind() {
		case reflect.Slice, reflect.Pointer:
			include = include || !v.IsNil()
		default:
			return nil, fmt.Errorf("%w: %T", errUnsupportedOptionalFieldType, vals[i])
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
// fields, be they `required` or `optional`, MUST be pointers and all `optional`
// fields MUST be provided in case they are present in the RLP being decoded.
//
// Typically, the arguments to this function mirror those passed to
// [EncodeStructFields] except for being pointers. See the example.
func (s *Stream) DecodeStructFields(required, optional []any) error {
	return s.FromList(func() error {
		for _, v := range required {
			if err := s.Decode(v); err != nil {
				return err
			}
		}

		for _, v := range optional {
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
