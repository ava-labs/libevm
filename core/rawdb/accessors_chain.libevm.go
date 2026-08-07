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
	"encoding/binary"
	"iter"

	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/libevm/ethdb"
	"github.com/ava-labs/libevm/rlp"
)

// A Header is an entry in the header keyspace: either a stored block
// header or a canonical number->hash mapping.
type Header struct {
	Number uint64
	Hash   common.Hash
	// RLP is the encoded block header stored against (Number, Hash). It is
	// nil i.f.f. the entry is the canonical number->hash mapping at Number,
	// in which case Hash is the canonical hash.
	//
	// The slice shares memory with the database iterator and is invalidated
	// when iteration advances; clone it to retain it.
	RLP rlp.RawValue
}

// Headers returns an iterator over the header keyspace, from the given block
// number onwards, yielding entries in key order along with any encountered
// errors. Iteration stops after the first error is yielded. Entries other
// than headers and canonical number->hash mappings (e.g. total difficulties)
// are skipped.
//
// Only the key-value store is read; data in the ancient (freezer) store is
// ignored.
func Headers(db ethdb.Iteratee, from uint64) iter.Seq2[Header, error] {
	return func(yield func(Header, error) bool) {
		it := db.NewIterator(headerPrefix, encodeBlockNumber(from))
		defer it.Release()

		var (
			hashKeyLength    = len(headerPrefix) + 8 + 1
			headerKeyLength  = len(headerPrefix) + 8 + common.HashLength
			numberOffset     = len(headerPrefix)
			hashSuffixOffset = numberOffset + 8
			hashOffset       = numberOffset + 8
		)
		for it.Next() {
			var (
				key = it.Key()
				h   Header
			)
			switch len(key) {
			case hashKeyLength: // number -> canonical hash
				if !bytes.Equal(key[hashSuffixOffset:], headerHashSuffix) {
					continue
				}
				h.Hash = common.BytesToHash(it.Value())
			case headerKeyLength: // number+hash -> header
				h.Hash = common.BytesToHash(key[hashOffset:])
				h.RLP = it.Value()
			default: // total-difficulty entries, etc.
				continue
			}
			h.Number = binary.BigEndian.Uint64(key[numberOffset:])
			if !yield(h, nil) {
				return
			}
		}
		if err := it.Error(); err != nil {
			yield(Header{}, err)
		}
	}
}

// A Body is a stored block body.
type Body struct {
	Number uint64
	Hash   common.Hash
	// RLP is the encoded block body stored against (Number, Hash).
	//
	// The slice shares memory with the database iterator and is invalidated
	// when iteration advances; clone it to retain it.
	RLP rlp.RawValue
}

// Bodies returns an iterator over the block-body keyspace, from the given
// block number onwards, yielding entries in key order along with any
// encountered errors. Iteration stops after the first error is yielded.
//
// Only the key-value store is read; data in the ancient (freezer) store is
// ignored.
func Bodies(db ethdb.Iteratee, from uint64) iter.Seq2[Body, error] {
	return func(yield func(Body, error) bool) {
		it := db.NewIterator(blockBodyPrefix, encodeBlockNumber(from))
		defer it.Release()

		var (
			keyLength    = len(blockBodyPrefix) + 8 + common.HashLength
			numberOffset = len(blockBodyPrefix)
			hashOffset   = numberOffset + 8
		)
		for it.Next() {
			key := it.Key()
			if len(key) != keyLength {
				continue
			}
			b := Body{
				Number: binary.BigEndian.Uint64(key[numberOffset:]),
				Hash:   common.BytesToHash(key[hashOffset:]),
				RLP:    it.Value(),
			}
			if !yield(b, nil) {
				return
			}
		}
		if err := it.Error(); err != nil {
			yield(Body{}, err)
		}
	}
}
