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

package tests

import (
	"flag"
	"testing"
)

func TestMain(m *testing.M) {
	flag.BoolVar(
		&failInsteadOfSkip,
		"libevm.fail_instead_of_skip",
		false,
		"If true and test cases are missing then respective tests fail instead of skipping",
	)
	flag.Parse()

	m.Run()
}

var failInsteadOfSkip bool

// failOrSkip propagates its arguments to Skipf or Fatalf, depending on the
// value of [failInsteadOfSkip], defaulting to skipping for backwards
// compatibility. See [TestMain] for the respective flag.
func failOrSkip(tb testing.TB, format string, args ...any) {
	tb.Helper()
	fn := tb.Skipf
	if failInsteadOfSkip {
		fn = tb.Fatalf
	}
	fn(format, args...)
}
