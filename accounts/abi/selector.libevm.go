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

package abi

import (
	"encoding/binary"
	"fmt"
)

// SelectorByteLen is the number of bytes in an ABI function selector.
const SelectorByteLen = 4

// A Selector is an ABI function selector. It is a uint32 instead of a [4]byte
// to allow for simpler hex literals.
type Selector uint32

// String returns a hex encoding of `s`.
func (s Selector) String() string {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(s))
	return fmt.Sprintf("%#x", b)
}

// ExtractSelector returns the first 4 bytes of the slice as a Selector. It
// assumes that its input is of sufficient length.
func ExtractSelector(data []byte) Selector {
	return Selector(binary.BigEndian.Uint32(data[:4]))
}

// Selector returns the Method's ID as a Selector.
func (m Method) Selector() Selector {
	return ExtractSelector(m.ID)
}

// MethodsBySelector maps 4-byte ABI selectors to the corresponding method
// representation. The key MUST be equivalent to the value's [Method.Selector].
type MethodsBySelector map[Selector]Method

// BySelector returns the the [Method]s keyed by their Selectors.
func (a *ABI) MethodsBySelector() MethodsBySelector {
	ms := make(MethodsBySelector)
	for _, m := range a.Methods {
		ms[m.Selector()] = m
	}
	return ms
}

// FindSelector extracts the Selector from `data` and, if it exists in `m`,
// returns it. The returned boolean functions as for regular map lookups. Unlike
// [ExtractSelector], FindSelector confirms that `data` has at least 4 bytes,
// treating invalid inputs as not found.
func (m MethodsBySelector) FindSelector(data []byte) (Selector, bool) {
	if len(data) < SelectorByteLen {
		return 0, false
	}
	sel := ExtractSelector(data)
	if _, ok := m[sel]; !ok {
		return 0, false
	}
	return sel, true
}
