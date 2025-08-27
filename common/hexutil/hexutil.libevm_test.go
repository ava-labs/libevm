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

import "testing"

var (
	encodeUint16Tests = []marshalTest{
		{uint16(0), "0x0"},
		{uint16(1), "0x1"},
		{uint16(0xff), "0xff"},
		{uint16(0x1122), "0x1122"},
	}

	decodeUint16Tests = []unmarshalTest{
		// invalid
		{input: `0`, wantErr: ErrMissingPrefix},
		{input: `0x`, wantErr: ErrEmptyNumber},
		{input: `0x01`, wantErr: ErrLeadingZero},
		{input: `0xfffff`, wantErr: ErrUint16Range},
		{input: `0xz1`, wantErr: ErrSyntax},
		// valid
		{input: `0x0`, want: uint16(0)},
		{input: `0x2`, want: uint16(0x2)},
		{input: `0x2F2`, want: uint16(0x2f2)},
		{input: `0X2F2`, want: uint16(0x2f2)},
		{input: `0xff`, want: uint16(0xff)},
		{input: `0x12af`, want: uint16(0x12af)},
		{input: `0xbbb`, want: uint16(0xbbb)},
		{input: `0xffff`, want: uint16(0xffff)},
	}
)

func TestEncodeUint16(t *testing.T) {
	for _, test := range encodeUint16Tests {
		input, ok := test.input.(uint16)
		if !ok {
			t.Errorf("input %v: not a uint16", test.input)
		}
		enc := EncodeUint16(input)
		if enc != test.want {
			t.Errorf("input %x: wrong encoding %s", test.input, enc)
		}
	}
}

func TestDecodeUint16(t *testing.T) {
	for _, test := range decodeUint16Tests {
		dec, err := DecodeUint16(test.input)
		if !checkError(t, test.input, err, test.wantErr) {
			continue
		}
		want, ok := test.want.(uint16)
		if !ok {
			t.Errorf("want %v: not a uint16", test.want)
		}
		if dec != want {
			t.Errorf("input %s: value mismatch: got %x, want %x", test.input, dec, want)
			continue
		}
	}
}
