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
	"bytes"
	"testing"

	"github.com/kr/pretty"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ava-labs/libevm/common"
)

func TestEncodeListToBuffer(t *testing.T) {
	vals := []uint{1, 2, 3, 4, 5}

	want, err := EncodeToBytes(vals)
	require.NoErrorf(t, err, "EncodeToBytes(%T{%[1]v})", vals)

	var got bytes.Buffer
	buf := NewEncoderBuffer(&got)
	err = EncodeListToBuffer(buf, vals)
	require.NoErrorf(t, err, "EncodeListToBuffer(..., %T{%[1]v})", vals)
	require.NoErrorf(t, buf.Flush(), "%T.Flush()", buf)

	assert.Equal(t, want, got.Bytes(), "EncodeListToBuffer(..., %T{%[1]v})", vals)
}

func TestDecodeList(t *testing.T) {
	vals := []uint{0, 1, 42, 314159}

	rlp, err := EncodeToBytes(vals)
	require.NoErrorf(t, err, "EncodeToBytes(%T{%[1]v})", vals)

	s := NewStream(bytes.NewReader(rlp), 0)
	got, err := DecodeList[uint](s)
	require.NoErrorf(t, err, "DecodeList[%T]()", vals[0])

	require.Equal(t, len(vals), len(got), "number of values returned by DecodeList()")
	for i, gotPtr := range got {
		assert.Equalf(t, vals[i], *gotPtr, "DecodeList()[%d]", i)
	}
}

func TestEncodeStructFields(t *testing.T) {
	type goldStandard struct {
		A uint64
		B uint64
		C *uint64
		D *uint64   `rlp:"optional"`
		E []uint64  `rlp:"optional"`
		F *[]uint64 `rlp:"optional"`
	}

	const (
		a uint64 = iota
		b
		cVal
		dVal
	)
	c := common.PointerTo(cVal)
	d := common.PointerTo(dVal)
	e := []uint64{40, 41}
	f := common.PointerTo([]uint64{50, 51})

	tests := []goldStandard{
		{a, b, c, d, e, f},       // 000 (which of d/e/f are nil)
		{a, b, c, d, e, nil},     // 001
		{a, b, c, d, nil, f},     // 010
		{a, b, c, d, nil, nil},   // 011
		{a, b, c, nil, e, f},     // 100
		{a, b, c, nil, e, nil},   // 101
		{a, b, c, nil, nil, f},   // 110
		{a, b, c, nil, nil, nil}, // 111
		// Empty and nil slices are treated differently when optional
		{a, b, c, nil, []uint64{}, nil},
	}

	for _, obj := range tests {
		obj := obj
		t.Run("", func(t *testing.T) {
			t.Logf("\n%s", pretty.Sprint(obj))

			want, err := EncodeToBytes(obj)
			require.NoErrorf(t, err, "EncodeToBytes([actual struct])")

			var got bytes.Buffer
			err = EncodeStructFields(
				&got,
				[]any{obj.A, obj.B, obj.C},
				[]any{obj.D, obj.E, obj.F},
			)
			require.NoErrorf(t, err, "EncodeStructFields(..., [required], [optional])")

			assert.Equal(t, want, got.Bytes(), "EncodeToBytes() vs EncodeStructFields()")
		})
	}
}
