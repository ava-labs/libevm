// Copyright 2024 the libevm authors.
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

package types_test

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "github.com/ava-labs/libevm/core/types"
	"github.com/ava-labs/libevm/libevm/ethtest"
	"github.com/ava-labs/libevm/rlp"
)

func TestHeaderRLPBackwardsCompatibility(t *testing.T) {
	// This is a deliberate change-detector test that locks in backwards
	// compatibility of RLP encoding.
	rng := ethtest.NewPseudoRand(42)

	const numExtraBytes = 16
	hdr := &Header{
		ParentHash:  rng.Hash(),
		UncleHash:   rng.Hash(),
		Coinbase:    rng.Address(),
		Root:        rng.Hash(),
		TxHash:      rng.Hash(),
		ReceiptHash: rng.Hash(),
		// Bloom populated below
		Difficulty: rng.Uint256().ToBig(),
		Number:     rng.BigUint64(),
		GasLimit:   rng.Uint64(),
		GasUsed:    rng.Uint64(),
		Time:       rng.Uint64(),
		Extra:      make([]byte, numExtraBytes), // populated below
		MixDigest:  rng.Hash(),
		// Nonce populated below

		BaseFee:          rng.BigUint64(),
		WithdrawalsHash:  rng.HashPtr(),
		BlobGasUsed:      rng.Uint64Ptr(),
		ExcessBlobGas:    rng.Uint64Ptr(),
		ParentBeaconRoot: rng.HashPtr(),
	}
	require.Equal(t, BloomByteLength, rng.Read(hdr.Bloom[:]))
	require.Equal(t, len(BlockNonce{}), rng.Read(hdr.Nonce[:]))
	require.Equal(t, numExtraBytes, rng.Read(hdr.Extra))

	// WARNING: changing this hex might break backwards compatibility of RLP
	// encoding (i.e. block hashes might change)!
	const wantHex = `f9029aa01a571e7e4d774caf46053201cfe0001b3c355ffcc93f510e671e8809741f0eeda0756095410506ec72a2c287fe83ebf68efb0be177e61acec1c985277e90e52087941bfc3bc193012ba58912c01fb35a3454831a8971a00bc9f064144eb5965c5e5d1020f9f90392e7e06ded9225966abc7c754b410e61a0d942eab201424f4320ec1e1ffa9390baf941629b9349977b5d48e0502dbb9386a035d9d550a9c113f78689b4c161c4605609bb57b83061914c42ad244daa7fc38eb90100718d155798390a6c6782181d1bac1dd64cd956332b008412ddc735f2994e297c8a088c6bb4c637542295ba3cbc3cd399c8127076f4d834d74d5b11a36b6d02e2fe3a583216aa4ccef052df9a96e7a454256bebabdfc38c429079f25913e0f1d7416b2f056c4a115fc757012b1757d2d69f0e5fb87c08605098d9031fa37cd0df6942c5a2da12a4424b978febf5479896165caf573cf82fb3aa10f6ebf6b62bef8ed36b8ea3d4b1ddb80c99afafa37cb8f3393eb6d802f5bc6c8cd6bcd168a7e0061a718218b848d945135b6dff228a4e66bade4717e6f4d318ac98fca12a053af6f98805a764fb5d523cb6f69029522cab9ced907cc75718f7e2c79154ef3fc7a04b31d39ae246d689f23176d679a62ff328f530407cbafd0146f45b2ed635282e88b36f6a5752feff5b881fc7fa9ef217f81d889f073433138e6ba58857515405d28f2a8e904bcda3066d382675f37dd1a18507b5fba02812f2701021506f27190adb52a1313f6d28c77d66ae1aa3d3d6757a762476f488294c7768cddd9ccf881b5da1b6a47970a3a0c8a2b7b2c44161190c82d5e1c8b55e05c7354f1e5f6512924c941fb3d93667dc889bc9df25654e163c88859405c51041475fa03a8c304a732153e20300c3482832d07b65f97958360da414cb438ce252aec6c2`
	want, err := hex.DecodeString(wantHex)
	require.NoError(t, err, "hex.DecodeString()")

	got, err := rlp.EncodeToBytes(hdr)
	require.NoErrorf(t, err, "rlp.EncodeToBytes(%T)", hdr)
	assert.Equalf(t, want, got, "rlp.EncodeToBytes(%T)", hdr)

	t.Run("ParseTree", func(t *testing.T) {
		got, err := rlp.ParseTree(got)
		require.NoErrorf(t, err, "rlp.ParseTree(rlp.EncodeToBytes(%T))", hdr)

		type (
			l = rlp.ListNode
			s = rlp.StringNode
		)
		u64Bytes := func(u uint64) []byte { return new(big.Int).SetUint64(u).Bytes() }
		want := l{
			s(hdr.ParentHash[:]),
			s(hdr.UncleHash[:]),
			s(hdr.Coinbase[:]),
			s(hdr.Root[:]),
			s(hdr.TxHash[:]),
			s(hdr.ReceiptHash[:]),
			s(hdr.Bloom[:]),
			s(hdr.Difficulty.Bytes()),
			s(hdr.Number.Bytes()),
			s(u64Bytes(hdr.GasLimit)),
			s(u64Bytes(hdr.GasUsed)),
			s(u64Bytes(hdr.Time)),
			s(hdr.Extra[:]),
			s(hdr.MixDigest[:]),
			s(hdr.Nonce[:]),
			s(hdr.BaseFee.Bytes()),
			s(hdr.WithdrawalsHash[:]),
			s(u64Bytes(*hdr.BlobGasUsed)),
			s(u64Bytes(*hdr.ExcessBlobGas)),
			s(hdr.ParentBeaconRoot[:]),
		}

		assert.Equal(t, want, got)
	})
}
