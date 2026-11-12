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
	"pgregory.net/rapid"
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

var _ rapid.StateMachine = (*state[struct{}])(nil)

type state[T comparable] struct {
	sut    Value[T]
	gen    *rapid.Generator[T]
	filled bool
	with   T
}

func properties[T comparable](t *rapid.T) {
	s := &state[T]{
		sut: New[T](),
		gen: rapid.Make[T](),
	}
	t.Repeat(rapid.StateMachineActions(s))
}

func TestProperties(t *testing.T) {
	rapid.Check(t, properties[int])
	rapid.Check(t, properties[string])
	rapid.Check(t, properties[struct {
		x uint64
		y [8]byte
		z string
	}])
}

func FuzzProperties(f *testing.F) {
	f.Fuzz(rapid.MakeFuzz(properties[int]))
}

func (s *state[T]) Check(t *rapid.T) {
	_, ok := s.sut.TryPeek()
	if got, want := ok, s.filled; got != want {
		t.Fatalf("%T.TryPeek() got %t; want %t", s.sut, got, want)
	}
}

func (s *state[T]) Put(t *rapid.T) {
	if s.filled {
		t.Log("full")
		return
	}
	s.with = s.gen.Draw(t, "put")
	s.sut.Put(s.with)
	s.filled = true
}

func (s *state[T]) Peek(t *rapid.T) {
	s.testGet(t, s.sut.Peek)
}

func (s *state[T]) Take(t *rapid.T) {
	s.testGet(t, s.sut.Take)
	s.filled = false
}

func (s *state[T]) testGet(t *rapid.T, fn func() T) {
	if !s.filled {
		t.Log("empty")
		return
	}
	if got, want := fn(), s.with; got != want {
		t.Fatalf("got %v; want %v", got, want)
	}
}

func (s *state[T]) TryPeek(t *rapid.T) {
	s.testTry(t, s.sut.TryPeek)
}

func (s *state[T]) TryTake(t *rapid.T) {
	s.testTry(t, s.sut.TryTake)
	s.filled = false
}

func (s *state[T]) testTry(t *rapid.T, fn func() (T, bool)) {
	got, gotOK := fn()
	if want, wantOK := s.with, s.filled; gotOK != wantOK || (wantOK && got != want) {
		t.Fatalf("got (%v, %t); want (%v, %t)", got, gotOK, want, wantOK)
	}
}

func (s *state[T]) PeekCtx(t *rapid.T) {
	s.testCtx(t, s.sut.PeekCtx)
}

func (s *state[T]) TakeCtx(t *rapid.T) {
	if s.testCtx(t, s.sut.TakeCtx) == nil {
		s.filled = false
	}
}

func (s *state[T]) testCtx(t *rapid.T, fn func(context.Context) (T, error)) error {
	ctx, cancel := context.WithCancel(t.Context())
	if !s.filled || rapid.Bool().Draw(t, "cancel") {
		cancel()
	} else {
		defer cancel()
	}

	got, err := fn(ctx)
	switch {
	case s.filled && err == nil:
		if want := s.with; got != want {
			t.Fatalf("got %v; want %v", got, want)
		}
	default:
		if want := context.Canceled; err != want {
			t.Fatalf("got error %v; want %v", err, context.Canceled)
		}
	}
	return err
}

func (s *state[T]) Reset(t *rapid.T) {
	s.sut.Reset()
	s.filled = false
}
