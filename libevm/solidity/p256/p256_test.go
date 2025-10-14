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

package p256_test

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"

	"github.com/holiman/uint256"
	"github.com/stretchr/testify/require"

	"github.com/ava-labs/libevm/accounts/abi"
	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/libevm/core/vm"
	"github.com/ava-labs/libevm/libevm"
	"github.com/ava-labs/libevm/libevm/ethtest"
	"github.com/ava-labs/libevm/libevm/hookstest"

	_ "embed"
)

//go:generate solc --output-dir ./ --overwrite --abi --bin P256Proxy.sol

var (
	//go:embed P256Proxy.bin
	proxyBinHex string
	//go:embed P256Proxy.abi
	proxyABIJSON []byte
)

func TestP256(t *testing.T) {
	stub := hookstest.Stub{
		PrecompileOverrides: map[common.Address]libevm.PrecompiledContract{
			common.BytesToAddress([]byte{1, 0}): &vm.P256Verify{},
		},
	}
	stub.Register(t)

	_, evm := ethtest.NewZeroEVM(t, ethtest.WithPUSH0Enabled()...)
	sdb := evm.StateDB
	eoa := common.Address{'e', 'o', 'a'}
	sdb.CreateAccount(eoa)
	sdb.AddBalance(eoa, new(uint256.Int).SetAllOne())

	caller := vm.AccountRef(eoa)
	_, proxy, _, err := evm.Create(caller, common.Hex2Bytes(proxyBinHex), 30e6, uint256.NewInt(0))
	require.NoErrorf(t, err, "%T.Create([P256Proxy])", evm)

	proxyABI, err := abi.JSON(bytes.NewReader(proxyABIJSON))
	require.NoError(t, err, "abi.JSON([P256Proxy])")

	call := func(t *testing.T, digest, r, s, x, y common.Hash) []byte {
		t.Helper()
		data, err := proxyABI.Pack("verify", digest, r, s, x, y)
		require.NoErrorf(t, err, "%T.Pack(%q)", proxyABI, "verify")

		got, _, err := evm.StaticCall(caller, proxy, data, 30e6)
		require.NoError(t, err, "evm.Call(P256Proxy.verify(...))")
		return got
	}

	for range 100 {
		t.Run("", func(t *testing.T) {
			var digest common.Hash
			_, err = rand.Read(digest[:])
			require.NoError(t, err, "crypto/rand.Read()")

			key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
			require.NoError(t, err, "ecdsa.GenerateKey(elliptic.P256(), crypto/rand.Reader)")
			bigR, bigS, err := ecdsa.Sign(rand.Reader, key, digest[:])
			require.NoError(t, err, "ecdsa.Sign(...)")

			var r, s, x, y common.Hash
			bigR.FillBytes(r[:])
			bigS.FillBytes(s[:])
			key.X.FillBytes(x[:])
			key.Y.FillBytes(y[:])

			t.Run("valid", func(t *testing.T) {
				require.Equal(t, []byte{31: 1}, call(t, digest, r, s, x, y))
			})
			t.Run("invalid", func(t *testing.T) {
				digest := digest
				digest[0]++
				require.Equal(t, make([]byte, 32), call(t, digest, r, s, x, y))
			})
		})
	}
}
