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

func (b EncoderBuffer) InList(fn func() error) error {
	l := b.List()
	if err := fn(); err != nil {
		return err
	}
	b.ListEnd(l)
	return nil
}

func EncodeListToBuffer[T any](b EncoderBuffer, vals []T) error {
	return b.InList(func() error {
		for _, v := range vals {
			if err := Encode(b, v); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Stream) FromList(fn func() error) error {
	if _, err := s.List(); err != nil {
		return err
	}
	if err := fn(); err != nil {
		return err
	}
	return s.ListEnd()
}

func DecodeList[T any](s *Stream) ([]*T, error) {
	// From the package-level documentation:
	//
	// > Note that package rlp never leaves a pointer-type struct field as nil
	// > unless one of the "nil" struct tags is present.
	//
	// We therefore return a non-nil pointer to maintain said invariant as it
	// makes use of this function easier.
	vals := make([]*T, 0)
	err := s.FromList(func() error {
		for s.MoreDataInList() {
			var v T
			if err := s.Decode(&v); err != nil {
				return err
			}
			vals = append(vals, &v)
		}
		return nil
	})
	return vals, err
}
