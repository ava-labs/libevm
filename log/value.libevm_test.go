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

package log

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTypeOf(t *testing.T) {
	type foo struct{}

	tests := map[any]string{
		nil:         "<nil>",
		int(0):      "int",
		int(1):      "int",
		uint(0):     "uint",
		foo{}:       "log.foo",
		(*foo)(nil): "*log.foo",
	}

	for in, want := range tests {
		got := TypeOf(in).LogValue()
		assert.Equalf(t, want, got.String(), "TypeOf(%T(%[1]v))", in, in)
	}
}
