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

// Package errs provides a mechanism for [testing error semantics] through
// unique identifiers, instead of depending on error messages that may result in
// change-detector tests.
//
// [testing error semantics]: https://google.github.io/styleguide/go/decisions#test-error-semantics
package errs

import (
	"errors"
	"fmt"
)

// An ID is a distinct numeric identifier for an error. Any two errors with the
// same ID will result in [errors.Is] returning true, regardless of their
// messages.
type ID int

// An identifier performs ID comparison, for embedding in all error types to
// provide their Is() method.
type identifier struct {
	id ID
}

func (id identifier) errorID() ID { return id.id }

func (id identifier) Is(target error) bool {
	t, ok := target.(interface{ errorID() ID })
	if !ok {
		return false
	}
	return t.errorID() == id.errorID()
}

func (id ID) asIdentifier() identifier { return identifier{id} }

// WithID returns a new error with the ID and message.
func WithID(id ID, msg string) error {
	return noWrap{errors.New(msg), id.asIdentifier()}
}

type noWrap struct {
	error
	identifier
}

// WithIDf is the formatted equivalent of [WithID], supporting the same
// wrapping semantics as [fmt.Errorf].
func WithIDf(id ID, format string, a ...any) error {
	switch err := fmt.Errorf(format, a...).(type) {
	case singleWrapper:
		return single{err, id.asIdentifier()}
	case multiWrapper:
		return multi{err, id.asIdentifier()}
	default:
		return noWrap{err, id.asIdentifier()}
	}
}

type singleWrapper interface {
	error
	Unwrap() error
}

type single struct {
	singleWrapper
	identifier
}

type multiWrapper interface {
	error
	Unwrap() []error
}

type multi struct {
	multiWrapper
	identifier
}
