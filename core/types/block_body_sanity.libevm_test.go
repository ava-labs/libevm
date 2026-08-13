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
	"errors"
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/libevm/rlp"
)

// maxSanitizedBodyExtra mirrors the 100KB cap that [Header.SanityCheck] applies
// to Header.Extra.
const maxSanitizedBodyExtra = 100 * 1024

var errBodyExtraTooLarge = errors.New("too large block body extra")

// sanitizedBodyHooks is a [BlockBodyHooks] implementation that also implements
// [BlockBodySanitizer], i.e. what a downstream chain carrying body extras is
// expected to do to bound them.
type sanitizedBodyHooks struct {
	Data []byte
	NOOPBlockBodyHooks
}

var (
	_ BlockBodyPayload[*sanitizedBodyHooks] = (*sanitizedBodyHooks)(nil)
	_ BlockBodySanitizer                    = (*sanitizedBodyHooks)(nil)
)

func (h *sanitizedBodyHooks) Copy() *sanitizedBodyHooks {
	return &sanitizedBodyHooks{Data: common.CopyBytes(h.Data)}
}

func (h *sanitizedBodyHooks) BlockRLPFieldsForEncoding(*BlockRLPProxy) *rlp.Fields {
	return &rlp.Fields{Required: []any{h.Data}}
}

func (h *sanitizedBodyHooks) BodySanityCheck(*Block) error {
	if n := len(h.Data); n > maxSanitizedBodyExtra {
		return fmt.Errorf("%w: size %d", errBodyExtraTooLarge, n)
	}
	return nil
}

// unsanitizedBodyHooks carries body extras but, like [NOOPBlockBodyHooks], does
// NOT implement [BlockBodySanitizer].
type unsanitizedBodyHooks struct {
	Data []byte
	NOOPBlockBodyHooks
}

var _ BlockBodyPayload[*unsanitizedBodyHooks] = (*unsanitizedBodyHooks)(nil)

func (h *unsanitizedBodyHooks) Copy() *unsanitizedBodyHooks {
	return &unsanitizedBodyHooks{Data: common.CopyBytes(h.Data)}
}

func TestBodySanityCheck(t *testing.T) {
	t.Run("no registered extras", func(t *testing.T) {
		TestOnlyClearRegisteredExtras()
		t.Cleanup(TestOnlyClearRegisteredExtras)

		// [NOOPBlockBodyHooks] isn't a [BlockBodySanitizer], so this MUST be a
		// noop rather than a panic on the type assertion.
		b := NewBlockWithHeader(&Header{Number: big.NewInt(1)})
		assert.NoError(t, b.BodySanityCheck(), "BodySanityCheck() with no registered extras")
		assert.NoError(t, b.SanityCheck(), "SanityCheck() with no registered extras")
	})

	t.Run("registered sanitizer", func(t *testing.T) {
		sizes := []struct {
			name    string
			size    int
			wantErr bool
		}{
			{name: "empty", size: 0},
			{name: "at limit", size: maxSanitizedBodyExtra},
			{name: "one over limit", size: maxSanitizedBodyExtra + 1, wantErr: true},
			{name: "far over limit", size: 10 * maxSanitizedBodyExtra, wantErr: true},
		}

		for _, s := range sizes {
			t.Run(s.name, func(t *testing.T) {
				TestOnlyClearRegisteredExtras()
				t.Cleanup(TestOnlyClearRegisteredExtras)
				extras := RegisterExtras[
					NOOPHeaderHooks, *NOOPHeaderHooks,
					sanitizedBodyHooks, *sanitizedBodyHooks,
					bool,
				]()

				b := NewBlockWithHeader(&Header{Number: big.NewInt(1)})
				extras.Block.Set(b, &sanitizedBodyHooks{Data: make([]byte, s.size)})

				err := b.BodySanityCheck()
				if s.wantErr {
					require.ErrorIsf(t, err, errBodyExtraTooLarge, "BodySanityCheck() with a %d-byte body extra", s.size)
				} else {
					require.NoErrorf(t, err, "BodySanityCheck() with a %d-byte body extra", s.size)
				}

				// [Block.SanityCheck] is the p2p entry point (via
				// NewBlockPacket.sanityCheck) so it MUST propagate the body
				// check, not just the header one.
				err = b.SanityCheck()
				if s.wantErr {
					require.ErrorIsf(t, err, errBodyExtraTooLarge, "SanityCheck() with a %d-byte body extra", s.size)
				} else {
					require.NoErrorf(t, err, "SanityCheck() with a %d-byte body extra", s.size)
				}
			})
		}
	})

	t.Run("registered extras without sanitizer", func(t *testing.T) {
		TestOnlyClearRegisteredExtras()
		t.Cleanup(TestOnlyClearRegisteredExtras)
		extras := RegisterExtras[
			NOOPHeaderHooks, *NOOPHeaderHooks,
			unsanitizedBodyHooks, *unsanitizedBodyHooks,
			bool,
		]()

		b := NewBlockWithHeader(&Header{Number: big.NewInt(1)})
		extras.Block.Set(b, &unsanitizedBodyHooks{Data: make([]byte, 100*maxSanitizedBodyExtra)})

		// This documents the deliberate limitation of the opt-in design: a chain
		// that registers body extras but doesn't implement [BlockBodySanitizer]
		// gets NO bound. If a default cap is ever introduced, this expectation
		// MUST change.
		assert.NoError(t, b.BodySanityCheck(), "BodySanityCheck() with a registered non-sanitizing payload")
	})

	t.Run("header error takes precedence", func(t *testing.T) {
		TestOnlyClearRegisteredExtras()
		t.Cleanup(TestOnlyClearRegisteredExtras)
		extras := RegisterExtras[
			NOOPHeaderHooks, *NOOPHeaderHooks,
			sanitizedBodyHooks, *sanitizedBodyHooks,
			bool,
		]()

		// A block number that doesn't fit in a uint64 fails [Header.SanityCheck].
		b := NewBlockWithHeader(&Header{Number: new(big.Int).Lsh(big.NewInt(1), 64)})
		extras.Block.Set(b, &sanitizedBodyHooks{Data: make([]byte, maxSanitizedBodyExtra+1)})

		err := b.SanityCheck()
		require.Error(t, err, "SanityCheck() with both an invalid header and an oversized body extra")
		assert.NotErrorIs(t, err, errBodyExtraTooLarge, "SanityCheck() reports the header failure first")
	})
}
