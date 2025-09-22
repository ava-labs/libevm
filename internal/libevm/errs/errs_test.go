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

package errs

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIs(t *testing.T) {
	ids := []ID{0, 42}
	errsByID := make(map[ID][]error)
	for _, id := range ids {
		errsByID[id] = []error{
			WithID(id, "WithID()"),
			WithIDf(id, "WithIDf() no wrapping"),
			WithIDf(id, "WithIDf() wrap one %w", errors.New("x")),
			WithIDf(id, "WithIDf() wrap multi %w + %w", errors.New("x"), errors.New("y")),
		}
	}

	unidentified := []error{
		errors.New("errors.New()"),
		fmt.Errorf("fmt.Errorf()"),
	}

	for id, errs := range errsByID {
		for _, err := range errs {
			for targetID, targets := range errsByID {
				want := id == targetID
				for _, target := range targets {
					assert.Equalf(t, want, errors.Is(err, target), "errors.Is(%v [ID %d], %v [ID %d])", err, id, target, targetID)
				}
			}

			for _, target := range unidentified {
				assert.NotErrorIsf(t, err, target, "error ID %d", id)
			}
		}
	}
}

func Example() {
	id42 := WithID(42, "hello")
	alsoWithID42 := WithIDf(42, "%s", "world")
	unidentified := errors.New("hello")

	fmt.Println(errors.Is(id42, alsoWithID42))
	fmt.Println(errors.Is(id42, unidentified))

	// Output:
	// true
	// false
}
