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
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/libevm/rlp"
)

// cChainBodyExtras carries the same additional fields as the ava-labs/coreth
// [Body] and implements [BodyHooks] to achieve equivalent RLP {en,de}coding.
type cChainBodyExtras struct {
	Version uint32
	ExtData *[]byte
}

var _ BodyHooks = (*cChainBodyExtras)(nil)

func (e *cChainBodyExtras) AppendRLPFields(b rlp.EncoderBuffer, _ bool) error {
	b.WriteUint64(uint64(e.Version))
	if e.ExtData != nil {
		b.WriteString(string(*e.ExtData))
	} else {
		b.WriteString("")
	}
	return nil
}

func (e *cChainBodyExtras) DecodeExtraRLPFields(s *rlp.Stream) error {
	if err := s.Decode(&e.Version); err != nil {
		return err
	}

	buf, err := s.Bytes()
	if err != nil {
		return err
	}
	if len(buf) > 0 {
		e.ExtData = &buf
	} else {
		// Respect the `rlp:"nil"` field tag.
		e.ExtData = nil
	}

	return nil
}

func TestBodyRLPCChainCompat(t *testing.T) {
	// The inputs to this test were used to generate the expected RLP with
	// ava-labs/coreth. This serves as both an example of how to use [BodyHooks]
	// and a test of compatibility.

	t.Cleanup(func() {
		todoRegisteredBodyHooks = NOOPBodyHooks{}
	})

	to := common.HexToAddress(`decafc0ffeebad`)
	body := &Body{
		Transactions: []*Transaction{
			NewTx(&LegacyTx{
				Nonce: 42,
				To:    &to,
			}),
		},
		Uncles: []*Header{ /* RLP encoding differs in ava-labs/coreth */ },
	}

	const version = 314159
	tests := []struct {
		name  string
		extra *cChainBodyExtras
		// WARNING: changing these values might break backwards compatibility of
		// RLP encoding!
		wantRLPHex string
	}{
		{
			extra: &cChainBodyExtras{
				Version: version,
			},
			wantRLPHex: `e5dedd2a80809400000000000000000000000000decafc0ffeebad8080808080c08304cb2f80`,
		},
		{
			extra: &cChainBodyExtras{
				Version: version,
				ExtData: &[]byte{1, 4, 2, 8, 5, 7},
			},
			wantRLPHex: `ebdedd2a80809400000000000000000000000000decafc0ffeebad8080808080c08304cb2f86010402080507`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wantRLP, err := hex.DecodeString(tt.wantRLPHex)
			require.NoError(t, err)

			t.Run("Encode", func(t *testing.T) {
				todoRegisteredBodyHooks = tt.extra
				got, err := rlp.EncodeToBytes(body)
				require.NoError(t, err)
				assert.Equal(t, wantRLP, got)
			})

			t.Run("Decode", func(t *testing.T) {
				var extra cChainBodyExtras
				todoRegisteredBodyHooks = &extra

				got := new(Body)
				err := rlp.DecodeBytes(wantRLP, got)
				require.NoError(t, err)
				// DO NOT MERGE without asserting `got` with [cmp.Diff]
				assert.Equal(t, tt.extra, &extra)
			})
		})
	}
}
