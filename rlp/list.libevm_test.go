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
	"io"
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

func TestStructFieldHelpers(t *testing.T) {
	type foo struct {
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
	f := &[]uint64{50, 51}

	tests := []foo{
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
		{a, b, c, nil, nil, &[]uint64{}},
	}

	for _, obj := range tests {
		obj := obj
		t.Run("", func(t *testing.T) {
			t.Logf("\n%s", pretty.Sprint(obj))

			wantRLP, err := EncodeToBytes(obj)
			require.NoErrorf(t, err, "EncodeToBytes([actual struct])")

			t.Run("EncodeStructFields", func(t *testing.T) {
				var got bytes.Buffer
				err = EncodeStructFields(
					&got,
					[]any{obj.A, obj.B, obj.C},
					Optional(obj.D, obj.E, obj.F),
				)
				require.NoErrorf(t, err, "EncodeStructFields(..., [required], [optional])")

				assert.Equal(t, wantRLP, got.Bytes(), "EncodeToBytes() vs EncodeStructFields()")
			})

			t.Run("DecodeStructFields", func(t *testing.T) {
				s := NewStream(bytes.NewReader(wantRLP), 0)
				var got foo
				err := s.DecodeStructFields(
					[]any{&got.A, &got.B, &got.C},
					Optional(&got.D, &got.E, &got.F),
				)
				require.NoError(t, err, "Stream.DecodeStructFields(...)")

				var want foo
				err = DecodeBytes(wantRLP, &want)
				require.NoError(t, err, "DecodeBytes(...)")

				assert.Equal(t, want, got, "DecodeBytes(..., [original struct]) vs Stream.DecodeStructFields(...)")
			})
		})
	}
}

//nolint:testableexamples // Demonstrating code equivalence, not outputs.
func ExampleStream_DecodeStructFields() {
	type inner struct {
		X uint64
	}

	type outer struct {
		A uint64
		B *inner `rlp:"nil"`
		C *inner `rlp:"optional"`
	}

	val := outer{
		A: 42,
		B: &inner{X: 42},
		C: &inner{X: 99},
	}

	// Errors are dropped for brevity for the sake of the example only.

	_ = Encode(io.Discard, val)
	// is equivalent to
	_ = EncodeStructFields(
		io.Discard,
		[]any{val.A, val.B},
		Optional(val.C),
	)

	r := bytes.NewReader(nil /*arbitrary RLP buffer*/)
	var decoded outer
	_ = Decode(r, &decoded)
	// is equivalent to
	_ = NewStream(r, 0).DecodeStructFields(
		[]any{
			&val.A,
			Nillable(&val.B),
		},
		Optional(&val.C),
	)

	// Note the parallels between the arguments passed to
	// {En,De}codeStructFields() and that, when decoding an optional or
	// `rlp:"nil`-tagged field, a pointer to the _field_ is required even though
	// in this example it will be a `**inner`.
}

func TestNillable(t *testing.T) {
	type inner struct {
		X uint64
	}

	type outer struct {
		A *uint64   `rlp:"nil"`
		B *inner    `rlp:"nil"`
		C *[]uint64 `rlp:"nil"`
	}

	// Unlike the `rlp:"optional"` tag, there is no interplay between nil-tagged
	// fields so we don't need the Cartesian product of all possible
	// combinations.
	var tests []outer
	for _, a := range []*uint64{
		nil,
		common.PointerTo[uint64](0),
	} {
		tests = append(tests, outer{a, nil, nil})
	}
	for _, b := range []*inner{
		nil,
		{0},
	} {
		tests = append(tests, outer{nil, b, nil})
	}
	for _, c := range []*[]uint64{
		nil,
		{},
		{0},
	} {
		tests = append(tests, outer{nil, nil, c})
	}

	// When a Nillable encounters an empty list it MUST set the field to nil,
	// not just ignore it.
	corruptInitialValue := func() outer {
		return outer{common.PointerTo[uint64](42), &inner{42}, &[]uint64{42}}
	}

	for _, obj := range tests {
		obj := obj
		t.Run("", func(t *testing.T) {
			rlp, err := EncodeToBytes(obj)
			require.NoErrorf(t, err, "EncodeToBytes(%+v)", obj)
			t.Logf("%s => %#x", pretty.Sprint(obj), rlp)

			// Although this is an immediate inversion of the line above, it
			// provides us with the canonical RLP decoding, which our input
			// struct may not honour.
			want := corruptInitialValue()
			err = DecodeBytes(rlp, &want)
			require.NoErrorf(t, err, "DecodeBytes(%#x, %T)", rlp, &want)

			s := NewStream(bytes.NewReader(rlp), 0)
			got := corruptInitialValue()
			err = s.DecodeStructFields(
				[]any{
					Nillable(&got.A),
					Nillable(&got.B),
					Nillable(&got.C),
				},
				nil,
			)
			require.NoError(t, err, "Stream.DecodeStructFields(...)")

			assert.Equal(t, want, got, "DecodeBytes(...) vs Stream.DecodeStructFields([fields wrapped in Nillable()])")
		})
	}
}
