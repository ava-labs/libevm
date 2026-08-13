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

package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ava-labs/libevm/common"
)

// nilUnsafeExtra implements [StateAccountExtraCopier] with a method that
// dereferences its receiver, i.e. what a normal implementation looks like. It
// exists to prove that `cloneStateAccount` never calls Copy() on a nil payload.
type nilUnsafeExtra struct {
	Data []byte
}

var _ StateAccountExtraCopier[*nilUnsafeExtra] = (*nilUnsafeExtra)(nil)

func (e *nilUnsafeExtra) Copy() *nilUnsafeExtra {
	return &nilUnsafeExtra{Data: common.CopyBytes(e.Data)} // panics on a nil receiver
}

func TestStateAccountExtraCopier(t *testing.T) {
	TestOnlyClearRegisteredExtras()
	t.Cleanup(TestOnlyClearRegisteredExtras)
	payloads := RegisterExtras[
		NOOPHeaderHooks, *NOOPHeaderHooks,
		NOOPBlockBodyHooks, *NOOPBlockBodyHooks,
		*nilUnsafeExtra,
	]().StateAccount

	t.Run("nil payload", func(t *testing.T) {
		acct := NewEmptyStateAccount()
		require.Nil(t, payloads.Get(acct), "test setup: default payload is a nil pointer")

		// MUST NOT panic: Copy() would dereference the nil receiver.
		cp := acct.Copy()
		assert.Nil(t, payloads.Get(cp), "payload of copied account stays nil")
	})

	t.Run("non-nil payload is deep-copied", func(t *testing.T) {
		acct := NewEmptyStateAccount()
		payloads.Set(acct, &nilUnsafeExtra{Data: []byte{1, 2, 3}})

		cp := acct.Copy()
		orig, copied := payloads.Get(acct), payloads.Get(cp)
		require.NotSame(t, orig, copied, "payload pointer MUST NOT be shared")

		copied.Data[0] = 0xff
		assert.Equal(t, []byte{1, 2, 3}, orig.Data, "mutating the copy MUST NOT affect the original")
	})
}
