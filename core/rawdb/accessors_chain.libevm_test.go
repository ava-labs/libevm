// Copyright 2026 the libevm authors.
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

package rawdb

import (
	"bytes"
	"math/big"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/libevm/core/types"
	"github.com/ava-labs/libevm/ethdb"
	"github.com/ava-labs/libevm/rlp"
)

// newTestBlock returns a block at the given height, with contents varied by
// height and canonicity so that all stored entries are distinguishable.
func newTestBlock(n uint64, canonical bool) *types.Block {
	hdr := &types.Header{
		Number: new(big.Int).SetUint64(n),
		Extra:  []byte{byte(n), byte(1)},
	}
	if canonical {
		hdr.Extra[1] = 0
	}

	var body types.Body
	if n%2 == 0 {
		nonce := n
		if !canonical {
			nonce += 1 << 32
		}
		body.Transactions = types.Transactions{types.NewTx(&types.LegacyTx{
			Nonce:    nonce,
			GasPrice: big.NewInt(1),
			Gas:      21_000,
			To:       &common.Address{},
			Value:    big.NewInt(2),
			V:        big.NewInt(3),
			R:        big.NewInt(4),
			S:        big.NewInt(5),
		})}
	}
	if n%3 == 0 {
		body.Uncles = []*types.Header{{Number: new(big.Int).SetUint64(n + 1)}}
	}

	return types.NewBlockWithHeader(hdr).WithBody(body)
}

// A keyedEntry pairs an expected iterator entry with the database key it is
// stored against, from which the expected yield order is derived.
type keyedEntry[E any] struct {
	key   []byte
	entry E
}

func sortedEntries[E any](keyed []keyedEntry[E]) []E {
	slices.SortFunc(keyed, func(a, b keyedEntry[E]) int {
		return bytes.Compare(a.key, b.key)
	})
	var entries []E
	for _, k := range keyed {
		entries = append(entries, k.entry)
	}
	return entries
}

// writeTestChain populates the database with canonical and sibling blocks at
// heights [0, last], returning the expected [Headers] and [Bodies] entries in
// yield order.
func writeTestChain(t *testing.T, db ethdb.KeyValueWriter, last uint64) ([]Header, []Body) {
	t.Helper()

	var (
		headers []keyedEntry[Header]
		bodies  []keyedEntry[Body]
	)
	for n := uint64(0); n <= last; n++ {
		canon := newTestBlock(n, true)
		for _, b := range []*types.Block{canon, newTestBlock(n, false)} {
			WriteBlock(db, b)
			// Total-difficulty entries extend header keys and MUST be skipped.
			WriteTd(db, b.Hash(), n, big.NewInt(1))

			hdrRLP, err := rlp.EncodeToBytes(b.Header())
			require.NoErrorf(t, err, "rlp.EncodeToBytes(header %d)", n)
			headers = append(headers, keyedEntry[Header]{
				key:   headerKey(n, b.Hash()),
				entry: Header{Number: n, Hash: b.Hash(), RLP: hdrRLP},
			})

			bodyRLP, err := rlp.EncodeToBytes(b.Body())
			require.NoErrorf(t, err, "rlp.EncodeToBytes(body %d)", n)
			bodies = append(bodies, keyedEntry[Body]{
				key:   blockBodyKey(n, b.Hash()),
				entry: Body{Number: n, Hash: b.Hash(), RLP: bodyRLP},
			})
		}

		WriteCanonicalHash(db, canon.Hash(), n)
		headers = append(headers, keyedEntry[Header]{
			key:   headerHashKey(n),
			entry: Header{Number: n, Hash: canon.Hash()},
		})
	}
	return sortedEntries(headers), sortedEntries(bodies)
}

func TestHeaders(t *testing.T) {
	db := NewMemoryDatabase()
	const last = 10
	want, _ := writeTestChain(t, db, last)
	const perHeight = 3 // two headers + one canonical mapping

	collect := func(from uint64, limit int) []Header {
		var got []Header
		for e, err := range Headers(db, from) {
			require.NoError(t, err)
			e.RLP = bytes.Clone(e.RLP)
			got = append(got, e)
			if len(got) == limit {
				break
			}
		}
		return got
	}

	tests := []struct {
		name string
		from uint64
		want []Header
	}{
		{
			name: "all",
			from: 0,
			want: want,
		},
		{
			name: "from middle",
			from: 7,
			want: want[7*perHeight:],
		},
		{
			name: "beyond latest",
			from: last + 1,
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, collect(tt.from, -1))
		})
	}

	t.Run("early break", func(t *testing.T) {
		assert.Equal(t, want[:4], collect(0, 4))
	})
}

func TestBodies(t *testing.T) {
	db := NewMemoryDatabase()
	const last = 10
	_, want := writeTestChain(t, db, last)
	const perHeight = 2 // canonical + sibling

	collect := func(from uint64, limit int) []Body {
		var got []Body
		for e, err := range Bodies(db, from) {
			require.NoError(t, err)
			e.RLP = bytes.Clone(e.RLP)
			got = append(got, e)
			if len(got) == limit {
				break
			}
		}
		return got
	}

	tests := []struct {
		name string
		from uint64
		want []Body
	}{
		{
			name: "all",
			from: 0,
			want: want,
		},
		{
			name: "from middle",
			from: 7,
			want: want[7*perHeight:],
		},
		{
			name: "beyond latest",
			from: last + 1,
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, collect(tt.from, -1))
		})
	}

	t.Run("early break", func(t *testing.T) {
		assert.Equal(t, want[:3], collect(0, 3))
	})
}
