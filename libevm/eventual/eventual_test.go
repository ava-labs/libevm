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

package eventual

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestValue(t *testing.T) {
	t.Run("Reset", func(t *testing.T) {
		v := New[int]()
		v.Put(0)
		v.Reset()
		v.Put(0) // would block without intervening Reset()
		v.Reset()
		v.Reset() // idempotent
	})

	type T = int

	tests := []struct {
		method       string
		blocking     func(Value[T]) T
		nonBlocking  func(Value[T]) (T, bool)
		putEveryTime bool
	}{
		{
			method:       "Peek",
			blocking:     (Value[T]).Peek,
			nonBlocking:  (Value[T]).TryPeek,
			putEveryTime: false, // unnecessary and would block
		},
		{
			method:       "Take",
			blocking:     (Value[T]).Take,
			nonBlocking:  (Value[T]).TryTake,
			putEveryTime: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			t.Parallel()
			sut := New[T]()

			_, ok := tt.nonBlocking(sut)
			assert.Falsef(t, ok, "Try%s() before Put()", tt.method)

			const val = 42

			unblocked := make(chan struct{})
			go func() {
				assert.Equalf(t, val, tt.blocking(sut), "%s() called before Put()", tt.method)
				close(unblocked)
			}()

			select {
			case <-unblocked:
				t.Errorf("%s() unblocked before Put()", tt.method)
			case <-time.After(200 * time.Millisecond):
				// TODO(arr4n): change this to use synctest when at Go 1.25
			}

			sut.Put(val)
			<-unblocked

			if tt.putEveryTime {
				sut.Put(val)
			}
			assert.Equal(t, val, tt.blocking(sut), "%s() called after Put()", tt.method)

			if tt.putEveryTime {
				sut.Put(val)
			}
			if got, ok := tt.nonBlocking(sut); !ok || got != val {
				t.Errorf("After Put(), Try%s() got (%d, %t); want (%d, true)", tt.method, got, ok, val)
			}
		})
	}
}
