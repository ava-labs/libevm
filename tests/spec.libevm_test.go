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

	"github.com/ava-labs/libevm/common"
)

func TestMain(m *testing.M) {
	flag.BoolVar(
		&fatalIfNoSpecTestDir,
		"libevm.fail_without_execution_spec_tests",
		false,
		"If true and no execution spec tests are found then respective tests will fail",
	)
	flag.Parse()

	m.Run()
}

var fatalIfNoSpecTestDir bool

// executionSpecTestDirExists is equivalent to [common.FileExist] except that if
// it were to return false and [fatalIfNoSpecTestDir] is true then it results in
// a fatal error. See the flag in [TestMain].
//
// Without this, the block and state execution spec tests fail silently when not
// present. This resulted in them not being run on libevm for over a year.
func executionSpecTestDirExists(t *testing.T, dirPath string) bool {
	t.Helper()
	if common.FileExist(dirPath) {
		return true
	}
	if fatalIfNoSpecTestDir {
		t.Fatalf("directory %q does not exist", dirPath)
	}
	return false
}
