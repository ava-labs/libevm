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

package state_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/libevm/core/rawdb"
	"github.com/ava-labs/libevm/core/state"
	"github.com/ava-labs/libevm/core/types"
	"github.com/ava-labs/libevm/libevm/ethtest"
)

// accountExtra is a [types.StateAccount] extra payload holding a reference type,
// which is only deep-copyable because it implements
// [types.StateAccountExtraCopier].
type accountExtra struct {
	Data *[]byte // MUST be exported
}

var _ types.StateAccountExtraCopier[*accountExtra] = (*accountExtra)(nil)

func (a *accountExtra) Copy() *accountExtra {
	if a.Data == nil {
		return &accountExtra{}
	}
	buf := common.CopyBytes(*a.Data)
	return &accountExtra{Data: &buf}
}

// TestExtraIsolatedByStateDBCopy demonstrates that a [types.StateAccount] extra
// payload isn't shared between a [state.StateDB] and its copy. Note the
// difference to the equivalent sub-test of TestGetSetExtra, which copies the
// StateDB *before* the account is loaded and therefore doesn't exercise
// `stateObject.deepCopy()` at all.
//
// The registered setter mutates the `*types.StateAccountExtra` in place, so a
// shallow copy of the [types.StateAccount] — whose Extra field is a pointer —
// leaks writes across StateDB instances. As StateDB.Copy() is used for miner
// snapshots, `eth_call` and `eth_estimateGas`, that would corrupt state.
func TestExtraIsolatedByStateDBCopy(t *testing.T) {
	types.TestOnlyClearRegisteredExtras()
	t.Cleanup(types.TestOnlyClearRegisteredExtras)
	payloads := types.RegisterExtras[
		types.NOOPHeaderHooks, *types.NOOPHeaderHooks,
		types.NOOPBlockBodyHooks, *types.NOOPBlockBodyHooks,
		*accountExtra,
	]().StateAccount

	rng := ethtest.NewPseudoRand(1234)
	addr := rng.Address()

	origBuf := rng.Bytes(8)
	origExtra := &accountExtra{Data: &origBuf}
	newBuf := rng.Bytes(8)
	newExtra := &accountExtra{Data: &newBuf}
	require.NotEqual(t, origExtra, newExtra, "test setup: distinct extra payloads")

	sdb, err := state.New(types.EmptyRootHash, state.NewDatabase(rawdb.NewMemoryDatabase()), nil)
	require.NoError(t, err, "state.New()")

	sdb.CreateAccount(addr)
	// This is what makes the test meaningful: the account is in the StateDB's
	// map of state objects, so Copy() routes it through
	// `stateObject.deepCopy()`.
	state.SetExtra(sdb, payloads, addr, origExtra)
	require.Equal(t, origExtra, state.GetExtra(sdb, payloads, addr), "GetExtra() immediately after SetExtra()")

	t.Run("write to copy", func(t *testing.T) {
		cp := sdb.Copy()
		require.Equal(t, origExtra, state.GetExtra(cp, payloads, addr), "GetExtra([copy]) before writing")

		state.SetExtra(cp, payloads, addr, newExtra)

		assert.Equal(t, newExtra, state.GetExtra(cp, payloads, addr), "GetExtra([copy]) after SetExtra([copy])")
		assert.Equal(t, origExtra, state.GetExtra(sdb, payloads, addr), "GetExtra([original]) MUST be unaffected by SetExtra([copy])")
	})

	// Mutating the payload *through* its pointer, rather than replacing it via
	// SetExtra(), is only isolated if the registered `SA` implements
	// [types.StateAccountExtraCopier]; `*accountExtra` does, below.
	t.Run("mutate through payload pointer", func(t *testing.T) {
		cp := sdb.Copy()

		origPayload := state.GetExtra(sdb, payloads, addr)
		copyPayload := state.GetExtra(cp, payloads, addr)
		require.NotSame(t, origPayload, copyPayload, "GetExtra([original]) vs GetExtra([copy]) MUST NOT return the same payload pointer")
		require.NotSame(t, origPayload.Data, copyPayload.Data, "payload's own reference field MUST NOT be shared")

		before := common.CopyBytes(*origPayload.Data)
		(*copyPayload.Data)[0] ^= 0xff

		assert.Equal(t, before, *origPayload.Data, "mutating the copy's payload in place MUST NOT be visible on the original")
		assert.NotEqual(t, before, *copyPayload.Data, "test setup: the mutation actually changed the copy")
	})

	t.Run("write to original", func(t *testing.T) {
		cp := sdb.Copy()

		state.SetExtra(sdb, payloads, addr, newExtra)
		t.Cleanup(func() { state.SetExtra(sdb, payloads, addr, origExtra) })

		assert.Equal(t, newExtra, state.GetExtra(sdb, payloads, addr), "GetExtra([original]) after SetExtra([original])")
		assert.Equal(t, origExtra, state.GetExtra(cp, payloads, addr), "GetExtra([copy]) MUST be unaffected by SetExtra([original])")
	})
}
