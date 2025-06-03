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

package p256verify

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"

	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/libevm/core/vm"
	"github.com/ava-labs/libevm/libevm"
	"github.com/ava-labs/libevm/libevm/ethtest"
	"github.com/ava-labs/libevm/libevm/hookstest"
	"github.com/ava-labs/libevm/params"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ulerdoganTestCase is the test case from
// https://github.com/ulerdogan/go-ethereum/blob/cec0b058115282168c5afc5197de3f6b5479dc4a/core/vm/testdata/precompiles/p256Verify.json,
// copied under LGPL. See the respective commit for copyright and license
// information.
const ulerdoganTestCase = `4cee90eb86eaa050036147a12d49004b6b9c72bd725d39d4785011fe190f0b4da73bd4903f0ce3b639bbbf6e8e80d16931ff4bcf5993d58468e8fb19086e8cac36dbcd03009df8c59286b162af3bd7fcc0450c9aa81be5d10d312af6c66b1d604aebd3099c618202fcfe16ae7770b0c49ab5eadf74b754204a3bb6060e44eff37618b065f9832de4ca6ca971a7a1adc826d0f7c00181a5fb2ddf79ae00b4e10e`

func TestPrecompile(t *testing.T) {
	assert.Equal(t, params.P256VerifyGas, Precompile{}.RequiredGas(nil), "RequiredGas()")

	type testCase struct {
		name        string
		in          []byte
		wantSuccess bool
	}

	tests := []testCase{
		{
			name: "empty input",
		},
		{
			name: "input too short",
			in:   make([]byte, inputLen-1),
		},
		{
			name: "input too long",
			in:   make([]byte, inputLen+1),
		},
		{
			name: "pub key at infinity",
			in:   make([]byte, inputLen),
		},
		{
			name: "pub key not on curve",
			in:   []byte{inputLen - 1: 1},
		},
		{
			name:        "ulerdogan",
			in:          common.Hex2Bytes(ulerdoganTestCase),
			wantSuccess: true,
		},
	}

	for range 50 {
		priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		require.NoError(t, err, "ecdsa.GenerateKey(elliptic.P256(), crypto/rand.Reader)")

		for range 50 {
			var hash [32]byte
			_, err := rand.Read(hash[:])
			require.NoErrorf(t, err, "crypto/rand.Read(%T)", hash)

			in, err := Sign(priv, hash)
			require.NoErrorf(t, err, "Sign([P256 key], %#x)", hash)
			tests = append(tests, testCase{
				name:        "fuzz",
				in:          in,
				wantSuccess: true,
			})
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Precompile{}.Run(tt.in)
			require.NoError(t, err, "Run() always returns nil, even on verification failure")

			var want []byte
			if tt.wantSuccess {
				want = common.LeftPadBytes([]byte{1}, 32)
			}
			assert.Equal(t, want, got)
		})
	}
}

func BenchmarkPrecompile(b *testing.B) {
	in := common.Hex2Bytes(ulerdoganTestCase)
	var p Precompile

	for range b.N {
		p.Run(in) //nolint:errcheck // Always nil
	}
}

func TestViaEVM(t *testing.T) {
	addr := common.Address{42}
	hooks := hookstest.Stub{
		PrecompileOverrides: map[common.Address]libevm.PrecompiledContract{
			addr: Precompile{},
		},
	}
	hooks.Register(t)

	_, evm := ethtest.NewZeroEVM(t)
	in := common.Hex2Bytes(ulerdoganTestCase)

	got, _, err := evm.Call(vm.AccountRef{}, addr, in, 25000, uint256.NewInt(0))
	require.NoError(t, err)
	assert.Equal(t, []byte{31: 1}, got)
}
