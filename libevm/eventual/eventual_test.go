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
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TODO(arr4n) tparallel claims that the top-level test should also call
// `t.Parallel()` but this seems unsubstantiated. Is my understanding of
// parallel testing inside `t.Run()` incorrect (i.e. each run concurrently but
// the primary test is serialised) or is the linter flagging false positives?
//
//nolint:tparallel
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
		withCtx      func(Value[T], context.Context) (T, error)
		putEveryTime bool
	}{
		{
			method:       "Peek",
			blocking:     (Value[T]).Peek,
			nonBlocking:  (Value[T]).TryPeek,
			withCtx:      (Value[T]).PeekCtx,
			putEveryTime: false, // unnecessary and would block
		},
		{
			method:       "Take",
			blocking:     (Value[T]).Take,
			nonBlocking:  (Value[T]).TryTake,
			withCtx:      (Value[T]).TakeCtx,
			putEveryTime: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			t.Parallel()
			sut := New[T]()

			{
				_, ok := tt.nonBlocking(sut)
				assert.Falsef(t, ok, "Try%s() before Put()", tt.method)
			}

			{
				ctx, cancel := context.WithCancelCause(t.Context())
				errCause := errors.New("because")
				cancel(errCause)
				_, err := tt.withCtx(sut, ctx)
				//nolint:testifylint // Doesn't need to fail the whole test with require
				assert.ErrorIsf(t, err, errCause, "%sCtx() before Put() and context cancelled with %v", tt.method, errCause)
			}

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

			if tt.putEveryTime {
				sut.Put(val)
			}
			if got, err := tt.withCtx(sut, t.Context()); err != nil || got != val {
				t.Errorf("After Put(), %sCtx() got (%d, %v); want (%d, nil)", tt.method, got, err, val)
			}
		})
	}
}
