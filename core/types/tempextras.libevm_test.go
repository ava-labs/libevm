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

package types

import (
	"testing"
)

func TestTempRegisteredExtras(t *testing.T) {
	TestOnlyClearRegisteredExtras()
	t.Cleanup(TestOnlyClearRegisteredExtras)

	type (
		primary struct {
			NOOPHeaderHooks
		}
		override struct {
			NOOPHeaderHooks
		}
	)

	RegisterExtras[primary, *primary, NOOPBlockBodyHooks, *NOOPBlockBodyHooks, bool]()
	testPrimaryExtras := func(t *testing.T) {
		t.Helper()
		assertHeaderHooksConcreteType[*primary](t)
	}

	t.Run("before_temp", testPrimaryExtras)
	t.Run("WithTempRegisteredExtras", func(t *testing.T) {
		WithTempRegisteredExtras(func(ExtraPayloads[*override, *NOOPBlockBodyHooks, bool]) {
			assertHeaderHooksConcreteType[*override](t)
		})
	})
	t.Run("after_temp", testPrimaryExtras)
}

func assertHeaderHooksConcreteType[WantT any](t *testing.T) {
	t.Helper()

	hdr := new(Header)
	switch got := hdr.hooks().(type) {
	case WantT:
	default:
		var want WantT
		t.Errorf("%T.hooks() got concrete type %T; want %T", hdr, got, want)
	}
}
