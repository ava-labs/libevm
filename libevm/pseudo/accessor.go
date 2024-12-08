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

package pseudo

// An Accessor provides access to T values held in other types.
type Accessor[From any, T any] struct {
	get func(From) *Type
	set func(From, *Type)
}

// NewAccessor constructs a new [Accessor]. The `get` function MUST return a
// [Type] holding a T.
func NewAccessor[From any, T any](get func(From) *Type, set func(From, *Type)) Accessor[From, T] {
	return Accessor[From, T]{get, set}
}

// Get returns the T held by the C.
func (a Accessor[C, T]) Get(from C) T {
	return MustNewValue[T](a.get(from)).Get()
}

// Get returns a pointer to the T held by the C, which is guaranteed to be
// non-nil. However, if T is itself a pointer, no guarantees are provided.
func (a Accessor[C, T]) GetPointer(from C) *T {
	return MustPointerTo[T](a.get(from)).Value.Get()
}

// Set sets the T carried by the C.
func (a Accessor[C, T]) Set(on C, val T) {
	a.set(on, From(val).Type)
}
