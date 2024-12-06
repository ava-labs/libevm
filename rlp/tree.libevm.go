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
	"encoding/binary"
	"errors"
	"io"
)

// An ItemNode is a parsed RLP item as part of a tree, which may have only a
// root node. Nodes contain only their unpacked values, not their length- and
// type-denoting tags.
type ItemNode interface {
	rlpItem()
}

var _ = []ItemNode{ListNode(nil), StringNode(nil), ByteNode(0)}

// A ListNode is a slice of RLP items. It is the ItemNode equivalent of [List].
type ListNode []ItemNode

// A StringNode is an RLP [ItemNode] holding an arbitrary byte slice. It is the
// ItemNode equivalent of [String].
type StringNode []byte

// An ByteNode is an RLP [ItemNode] representing an unsigned integer <= 127. It
// is the ItemNode equivalent of [Byte].
//
// [ParseTree] will only return an ByteNode if the value is in the range [0,127]
// but an ByteNode MAY be outside of this range for the purpose of re-encoding.
type ByteNode byte

func (ListNode) rlpItem()   {}
func (StringNode) rlpItem() {}
func (ByteNode) rlpItem()   {}

var (
	errConcatenated  = errors.New("concatenated items outside of list")
	errTrailingBytes = errors.New("trailing bytes after parsing")
	errTooLong       = errors.New("parsing >8 big-endian bytes")
)

// ParseTree parses the RLP-encoded buffer and returns one of the concrete
// ItemNode types. All [StringNode] instances will be backed by the same memory
// as the argument received by ParseTree.
func ParseTree(rlp []byte) (ItemNode, error) {
	return parse(rlp, false /*inList*/)
}

// parseList is a convenience wrapper around [slicer.short] and [slicer.long],
// returning their return buffer as a [ListNode].
func parseList(str []byte, err error) (ItemNode, error) {
	if err != nil {
		return nil, err
	}
	return parse(str, true)
}

func parse(rlp []byte, inList bool) (ItemNode, error) {
	buf := &slicer{buf: rlp, i: 0}
	var items []ItemNode

	for eof := false; !eof; {
		switch tag, err := buf.byte(); {
		case err == io.EOF:
			eof = true

		case err != nil:
			// Impossible but being defensive in case of a future refactor.
			return nil, err

		case tag <= 0x7f:
			items = append(items, ByteNode(tag))

		case tag <= 0xb7:
			str, err := buf.short(tag, 0x80)
			if err != nil {
				return nil, err
			}
			items = append(items, StringNode(str))

		case tag <= 0xbf:
			str, err := buf.long(tag, 0xb7)
			if err != nil {
				return nil, err
			}
			items = append(items, StringNode(str))

		case tag <= 0xf7:
			list, err := parseList(buf.short(tag, 0xc0))
			if err != nil {
				return nil, err
			}
			items = append(items, list)

		default:
			list, err := parseList(buf.long(tag, 0xf7))
			if err != nil {
				return nil, err
			}
			items = append(items, list)
		}

		if !inList && len(items) > 1 {
			return nil, errConcatenated
		}
	}

	if n := buf.left(); n > 0 {
		return nil, errTrailingBytes
	}
	if inList {
		return ListNode(items), nil
	}
	return items[0], nil
}

// A slicer is a byte-slice reader that returns slices backed by the same memory
// as its buffer.
type slicer struct {
	buf []byte
	i   uint64
}

func (s *slicer) len() uint64 {
	return uint64(len(s.buf))
}

func (s *slicer) left() uint64 {
	return s.len() - s.i
}

// next returns the next `n` bytes.
func (s *slicer) next(n uint64) ([]byte, error) {
	if n > s.left() {
		return nil, io.EOF
	}
	b := s.buf[s.i : s.i+n]
	s.i += n
	return b, nil
}

func (s *slicer) byte() (byte, error) {
	b, err := s.next(1)
	if err != nil {
		return 0, err
	}
	return b[0], nil
}

// short returns the bytes encoding either a string or a list of <=55 bytes.
func (s *slicer) short(tag, base byte) ([]byte, error) {
	return s.next(uint64(tag - base))
}

// long returns the bytes encoding either a string or a list of >55 bytes, first
// reading the length.
func (s *slicer) long(tag, base byte) ([]byte, error) {
	n, err := s.bigEndian(uint64(tag - base))
	if err != nil {
		return nil, err
	}
	return s.next(n)
}

// bigEndian returns the next `nBytes` bytes interpreted as a big-endian uint64.
func (s *slicer) bigEndian(nBytes uint64) (uint64, error) {
	if nBytes > 8 {
		return 0, errTooLong
	}
	buf, err := s.next(nBytes)
	if err != nil {
		return 0, err
	}

	var padded [8]byte
	copy(padded[8-len(buf):], buf)
	return binary.BigEndian.Uint64(padded[:]), nil
}
