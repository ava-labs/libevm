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
	"encoding/json"
	"testing"
)

var unmarshalUint16Tests = []unmarshalTest{
	// invalid encoding
	{input: "", wantErr: errJSONEOF},
	{input: "null", wantErr: errNonString(uint16T)},
	{input: "10", wantErr: errNonString(uint16T)},
	{input: `"0"`, wantErr: wrapTypeError(ErrMissingPrefix, uint16T)},
	{input: `"0x"`, wantErr: wrapTypeError(ErrEmptyNumber, uint16T)},
	{input: `"0x01"`, wantErr: wrapTypeError(ErrLeadingZero, uint16T)},
	{input: `"0x10000"`, wantErr: wrapTypeError(ErrUint16Range, uint16T)},
	{input: `"0xx"`, wantErr: wrapTypeError(ErrSyntax, uint16T)},
	{input: `"0xz1"`, wantErr: wrapTypeError(ErrSyntax, uint16T)},

	// valid encoding
	{input: `""`, want: uint16(0)},
	{input: `"0x0"`, want: uint16(0)},
	{input: `"0x2"`, want: uint16(0x2)},
	{input: `"0x2F2"`, want: uint16(0x2f2)},
	{input: `"0X2F2"`, want: uint16(0x2f2)},
	{input: `"0x1122"`, want: uint16(0x1122)},
	{input: `"0xbbb"`, want: uint16(0xbbb)},
	{input: `"0xffff"`, want: uint16(0xffff)},
}

func TestUnmarshalUint16(t *testing.T) {
	for _, test := range unmarshalUint16Tests {
		var v Uint16
		err := json.Unmarshal([]byte(test.input), &v)
		if !checkError(t, test.input, err, test.wantErr) {
			continue
		}
		if uint16(v) != test.want.(uint16) {
			t.Errorf("input %s: value mismatch: got %d, want %d", test.input, v, test.want)
			continue
		}
	}
}

func BenchmarkUnmarshalUint16(b *testing.B) {
	input := []byte(`"0x1234"`)
	for i := 0; i < b.N; i++ {
		var v Uint16
		v.UnmarshalJSON(input)
	}
}

func TestMarshalUint16(t *testing.T) {
	tests := []struct {
		input uint16
		want  string
	}{
		{0, "0x0"},
		{1, "0x1"},
		{0xff, "0xff"},
		{0x1122, "0x1122"},
		{0xffff, "0xffff"},
	}

	for _, test := range tests {
		out, err := json.Marshal(Uint16(test.input))
		if err != nil {
			t.Errorf("%d: %v", test.input, err)
			continue
		}
		if want := `"` + test.want + `"`; string(out) != want {
			t.Errorf("%d: MarshalJSON output mismatch: got %q, want %q", test.input, out, want)
			continue
		}
		if out := Uint16(test.input).String(); out != test.want {
			t.Errorf("%d: String mismatch: got %q, want %q", test.input, out, test.want)
			continue
		}
	}
}
