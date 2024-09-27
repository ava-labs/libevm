package types

import (
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/require"
)

func TestStateAccountRLP(t *testing.T) {
	// RLP encodings that don't involve extra payloads were generated on raw
	// geth StateAccounts *before* any libevm modifications, thus locking in
	// default behaviour. Encodings that involve a boolean payload were
	// generated on ava-labs/coreth StateAccounts to guarantee equivalence.
	tests := []struct {
		name    string
		acc     *StateAccount
		wantHex string
	}{
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
		{
			name: "true boolean extra",
			acc: &StateAccount{
				Nonce:    0x444444,
				Balance:  uint256.NewInt(0x666666),
				Root:     common.Hash{},
				CodeHash: []byte{0xbb, 0xbb, 0xbb},
				Extra: func() *RLPPayload {
					p, _ := RLPPayloadOf(true) // not an error being dropped
					return p
				}(),
			},
			wantHex: `0xee8344444483666666a0000000000000000000000000000000000000000000000000000000000000000083bbbbbb01`,
		},
		{
			name: "false boolean extra",
			acc: &StateAccount{
				Nonce:    0x444444,
				Balance:  uint256.NewInt(0x666666),
				Root:     common.Hash{},
				CodeHash: []byte{0xbb, 0xbb, 0xbb},
				Extra: func() *RLPPayload {
					p, _ := RLPPayloadOf(false)
					return p
				}(),
			},
			wantHex: `0xee8344444483666666a0000000000000000000000000000000000000000000000000000000000000000083bbbbbb80`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := rlp.EncodeToBytes(tt.acc)
			require.NoError(t, err)
			t.Logf("got: %#x", got)

			tt.wantHex = strings.TrimPrefix(tt.wantHex, "0x")
			require.Equal(t, common.Hex2Bytes(tt.wantHex), got)
		})
	}
}
