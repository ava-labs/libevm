package types

import (
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/libevm/pseudo"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/require"
)

func TestStateAccountRLP(t *testing.T) {
	// RLP encodings that don't involve extra payloads were generated on raw
	// geth StateAccounts *before* any libevm modifications, thus locking in
	// default behaviour. Encodings that involve a boolean payload were
	// generated on ava-labs/coreth StateAccounts to guarantee equivalence.

	type test struct {
		name     string
		register func()
		acc      *StateAccount
		wantHex  string
	}

	explicitFalseBoolean := test{
		name: "explicit false-boolean extra",
		register: func() {
			RegisterExtras(Extras[bool]{})
		},
		acc: &StateAccount{
			Nonce:    0x444444,
			Balance:  uint256.NewInt(0x666666),
			Root:     common.Hash{},
			CodeHash: []byte{0xbb, 0xbb, 0xbb},
			Extra: &StateAccountExtra{
				t: pseudo.From(false).Type,
			},
		},
		wantHex: `0xee8344444483666666a0000000000000000000000000000000000000000000000000000000000000000083bbbbbb80`,
	}

	// The vanilla geth code won't set payloads so we need to ensure that the
	// zero-value encoding is used instead of the null-value default as when
	// no type is registered.
	implicitFalseBoolean := explicitFalseBoolean
	implicitFalseBoolean.name = "implicit false-boolean extra as zero-value of registered type"
	// Clearing the Extra makes the `false` value implicit and due only to the
	// fact that we register `bool`. Most importantly, note that `wantHex`
	// remains identical.
	implicitFalseBoolean.acc.Extra = nil

	tests := []test{
		explicitFalseBoolean,
		implicitFalseBoolean,
		{
			name: "true-boolean extra",
			register: func() {
				RegisterExtras(Extras[bool]{})
			},
			acc: &StateAccount{
				Nonce:    0x444444,
				Balance:  uint256.NewInt(0x666666),
				Root:     common.Hash{},
				CodeHash: []byte{0xbb, 0xbb, 0xbb},
				Extra: &StateAccountExtra{
					t: pseudo.From(true).Type,
				},
			},
			wantHex: `0xee8344444483666666a0000000000000000000000000000000000000000000000000000000000000000083bbbbbb01`,
		},
		{
			name: "vanilla geth account",
			acc: &StateAccount{
				Nonce:    0xcccccc,
				Balance:  uint256.NewInt(0x555555),
				Root:     common.MaxHash,
				CodeHash: []byte{0x77, 0x77, 0x77},
			},
			wantHex: `0xed83cccccc83555555a0ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff83777777`,
		},
		{
			name: "vanilla geth account",
			acc: &StateAccount{
				Nonce:    0x444444,
				Balance:  uint256.NewInt(0x666666),
				Root:     common.Hash{},
				CodeHash: []byte{0xbb, 0xbb, 0xbb},
			},
			wantHex: `0xed8344444483666666a0000000000000000000000000000000000000000000000000000000000000000083bbbbbb`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.register != nil {
				registeredExtras = nil
				tt.register()
				t.Cleanup(func() {
					registeredExtras = nil
				})
			}

			got, err := rlp.EncodeToBytes(tt.acc)
			require.NoError(t, err)
			t.Logf("got: %#x", got)

			tt.wantHex = strings.TrimPrefix(tt.wantHex, "0x")
			require.Equal(t, common.Hex2Bytes(tt.wantHex), got)
		})
	}
}
