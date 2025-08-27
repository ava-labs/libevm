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

import "strconv"

var ErrUint16Range = &decError{"hex number > 16 bits"}

// DecodeUint16 decodes a hex string with 0x prefix as a quantity.
func DecodeUint16(input string) (uint16, error) {
	raw, err := checkNumber(input)
	if err != nil {
		return 0, err
	}
	dec, err := strconv.ParseUint(raw, 16, 16)
	if err != nil {
		err = mapError(err)
		if err == ErrUint64Range {
			return 0, ErrUint16Range
		}
	}
	return uint16(dec), err //nolint:gosec // G115 won't overflow uint16 as ParseUint uses 16 bits
}

// EncodeUint16 encodes i as a hex string with 0x prefix.
func EncodeUint16(i uint16) string {
	enc := make([]byte, 2, 6)
	copy(enc, "0x")
	return string(strconv.AppendUint(enc, uint64(i), 16))
}
