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

package hexutil

import (
	"fmt"
	"reflect"
	"strconv"
)

var uint16T = reflect.TypeOf(Uint16(0))

// Uint16 marshals/unmarshals as a JSON string with 0x prefix.
// The zero value marshals as "0x0".
type Uint16 uint16

// MarshalText implements encoding.TextMarshaler.
func (b Uint16) MarshalText() ([]byte, error) {
	buf := make([]byte, 2, 6)
	copy(buf, `0x`)
	buf = strconv.AppendUint(buf, uint64(b), 16)
	return buf, nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (b *Uint16) UnmarshalJSON(input []byte) error {
	if !isString(input) {
		return errNonString(uint16T)
	}
	return wrapTypeError(b.UnmarshalText(input[1:len(input)-1]), uint16T)
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (b *Uint16) UnmarshalText(input []byte) error {
	raw, err := checkNumberText(input)
	if err != nil {
		return err
	}
	if len(raw) > 4 {
		return ErrUint16Range
	}
	var dec uint16
	for _, byte := range raw {
		nib := decodeNibble(byte)
		if nib == badNibble {
			return ErrSyntax
		}
		dec *= 16
		dec += uint16(nib)
	}

	*b = Uint16(dec)
	return nil
}

// String returns the hex encoding of b.
func (b Uint16) String() string {
	return EncodeUint16(uint16(b))
}

// ImplementsGraphQLType returns true if Uint16 implements the provided GraphQL type.
func (b Uint16) ImplementsGraphQLType(name string) bool { return name == "Int" }

// UnmarshalGraphQL unmarshals the provided GraphQL query data.
func (b *Uint16) UnmarshalGraphQL(input interface{}) error {
	var err error
	switch input := input.(type) {
	case string:
		return b.UnmarshalText([]byte(input))
	case int32:
		*b = Uint16(input)
	default:
		err = fmt.Errorf("unexpected type %T for Int", input)
	}
	return err
}
