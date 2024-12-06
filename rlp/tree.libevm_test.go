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

package rlp

import (
	"bytes"
	"encoding/hex"
	"math/rand"
	"runtime"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/require"
)

func TestParseTreeNoErrorOnUpstreamTests(t *testing.T) {
	for _, tt := range encTests {
		t.Run("", func(t *testing.T) {
			if tt.error != "" {
				return
			}
			switch tt.val.(type) {
			// These return arbitrary byte slices, not real RLP.
			case *testEncoder:
				return
			case struct{ E testEncoderValueMethod }:
				return
			case undecodableEncoder:
				return
			}

			buf, err := hex.DecodeString(tt.output)
			require.NoErrorf(t, err, "hex.DecodeString(%T.output)", tt)

			_, err = ParseTree(buf)
			require.NoErrorf(t, err, "ParseTree(%T => %#x)", tt.val, buf)
		})
	}
}

func TestParseTree(t *testing.T) {
	t.Parallel()

	type test struct {
		fuzzed   bool
		fuzzSeed int64
		value    any
		want     ItemNode
	}
	tests := []test{{
		value: []byte(nil),
		want:  StringNode{},
	}}

	for i := 0; i < 3e3; i++ {
		seed := int64(i)
		rng := rand.New(rand.NewSource(seed)) //nolint:gosec // Not security; reproducible (deterministic) fuzzing is useful
		val, node := randomItemNode(rng)
		tests = append(tests, test{
			fuzzed:   true,
			fuzzSeed: seed,
			value:    val,
			want:     node,
		})
	}

	throttle := make(chan struct{}, runtime.GOMAXPROCS(0))
	for _, tt := range tests {
		tt := tt
		t.Run("", func(t *testing.T) {
			t.Parallel()
			throttle <- struct{}{}
			defer func() { <-throttle }()

			if tt.fuzzed {
				t.Logf("Fuzzing seed: %d", tt.fuzzSeed)
			}

			buf := bytes.NewBuffer(nil)
			require.NoError(t, Encode(buf, tt.value))
			t.Logf("RLP encoding of %T: %#x\nValue: %x", tt.value, buf.Bytes(), tt.value)

			got, err := ParseTree(buf.Bytes())
			require.NoError(t, err)

			if diff := cmp.Diff(tt.want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("%s", diff)
			}
		})
	}
}

func randomItemNode(rng *rand.Rand) (any, ItemNode) {
	switch rng.Intn(3) {
	case 0:
		v := byte(rng.Intn(128))
		if v == 0 {
			// https://ethereum.org/en/developers/docs/data-structures-and-encoding/rlp/#:~:text=thus%20making%20the%20integer%20value%20zero%20equivalent%20to%20the%20empty%20byte%20array
			return v, StringNode{}
		}
		return v, ByteNode(v)

	case 1:
		v := make([]byte, rng.Intn(110)) // >55 is a "long" string so ~half of each
		rng.Read(v)                      //nolint:gosec // Documented as returning nil error
		v = bytes.TrimLeft(v, "\x00")

		var node ItemNode
		switch n := len(v); {
		case n == 1 && v[0] < 128:
			node = ByteNode(v[0])
		default:
			node = StringNode(v)
		}
		return v, node

	default:
		var (
			vals []any
			list ListNode
		)
		for i, n := 0, rng.Intn(5); i < n; i++ {
			v, item := randomItemNode(rng)
			vals = append(vals, v)
			list = append(list, item)
		}
		return vals, list
	}
}
