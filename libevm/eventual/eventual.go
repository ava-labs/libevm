// Copyright 2025-2026 the libevm authors.
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

// Package eventual provides a synchronisation primitive for delayed setting and
// getting of values.
package eventual

// A Value holds a value that is set at some unknown point in the future and
// used, possibly concurrently, by one or more peekers or a single taker
// (together, "getters"). The zero value is NOT ready for use.
//
// Although all methods are threadsafe, the returned `T` values might not be.
type Value[T any] struct {
	ch chan T
}

// New returns a new [Value].
func New[T any]() Value[T] {
	return Value[T]{
		ch: make(chan T, 1),
	}
}

// Put sets the value, unblocking any current and future getters. Put does not
// require a concurrent getter, however it is NOT possible to overwrite the value without an
// intervening call to [Value.Take], [Value.TryTake], or [Value.Reset].
func (v Value[T]) Put(x T) {
	v.ch <- x
}

// Peek returns the value after making it available for other getters. It blocks
// until a call to [Value.Put].
func (v Value[T]) Peek() T {
	x := <-v.ch
	v.ch <- x
	return x
}

// TryPeek is the non-blocking equivalent of [Value.Peek], returning true i.f.f.
// the respective call to [Value.Put] has already occurred.
func (v Value[T]) TryPeek() (T, bool) {
	select {
	case x := <-v.ch:
		v.ch <- x
		return x, true
	default:
		return v.zero(), false
	}
}

// Take returns the value and resets `v` to its default state as if immediately
// after construction. It blocks until a call to [Value.Put].
func (v Value[T]) Take() T {
	return <-v.ch
}

// TryTake is the non-blocking equivalent of [Value.Take], returning true i.f.f.
// the respective call to [Value.Put] has already occurred.
func (v Value[T]) TryTake() (T, bool) {
	select {
	case x := <-v.ch:
		return x, true
	default:
		return v.zero(), false
	}
}

// Reset is equivalent to [Value.TryTake] with the return arguments dropped. It
// exists for improved readability at the call site.
func (v Value[_]) Reset() {
	v.TryTake()
}

func (v Value[T]) zero() (z T) { return }
