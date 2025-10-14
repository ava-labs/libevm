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

//go:generate solc --output-dir ./ --overwrite --abi --bin P256SmokeTest.sol

var (
	//go:embed P256SmokeTest.bin
	proxyBinHex string
	//go:embed P256SmokeTest.abi
	proxyABIJSON []byte
)

func TestP256(t *testing.T) {
	stub := hookstest.Stub{
		PrecompileOverrides: map[common.Address]libevm.PrecompiledContract{
			common.BytesToAddress([]byte{1, 0}): &vm.P256Verify{},
		},
	}
	stub.Register(t)

	proxyABI, err := abi.JSON(bytes.NewReader(proxyABIJSON))
	require.NoError(t, err, "abi.JSON([P256Proxy])")
	pack := func(t *testing.T, method string, in p256Input) []byte {
		t.Helper()
		buf, err := proxyABI.Pack(method, in.digest, in.r, in.s, in.x, in.y)
		require.NoError(t, err, "%T.Pack(%q, ...)", proxyABI, method)
		return buf
	}

	sdb, evm := ethtest.NewZeroEVM(t, ethtest.WithPUSH0Enabled()...)
	eoa := common.Address{'e', 'o', 'a'}
	sdb.CreateAccount(eoa)
	sdb.AddBalance(eoa, new(uint256.Int).SetAllOne())
	caller := vm.AccountRef(eoa)

	creationCode := append(common.Hex2Bytes(proxyBinHex), pack(t, "", randomP256Input(t))...)
	_, proxy, _, err := evm.Create(caller, creationCode, 30e6, uint256.NewInt(0))
	require.NoErrorf(t, err, "%T.Create([P256Proxy])", evm)

	call := func(t *testing.T, in p256Input) []byte {
		t.Helper()
		got, _, err := evm.StaticCall(caller, proxy, pack(t, "verify", in), 30e6)
		require.NoError(t, err, "evm.Call(P256Proxy.verify(...))")
		return got
	}

	for range 100 {
		t.Run("", func(t *testing.T) {
			in := randomP256Input(t)
			t.Run("valid", func(t *testing.T) {
				require.Equal(t, []byte{31: 1}, call(t, in))
			})
			t.Run("invalid", func(t *testing.T) {
				in.digest[0]++
				require.Equal(t, make([]byte, 32), call(t, in))
			})
		})
	}
}

type p256Input struct {
	digest, r, s, x, y common.Hash
}

func randomP256Input(t *testing.T) p256Input {
	t.Helper()

	var out p256Input

	_, err := rand.Read(out.digest[:])
	require.NoError(t, err, "crypto/rand.Read()")

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err, "ecdsa.GenerateKey(elliptic.P256(), crypto/rand.Reader)")
	bigR, bigS, err := ecdsa.Sign(rand.Reader, key, out.digest[:])
	require.NoError(t, err, "ecdsa.Sign(...)")

	bigR.FillBytes(out.r[:])
	bigS.FillBytes(out.s[:])
	key.X.FillBytes(out.x[:])
	key.Y.FillBytes(out.y[:])

	return out
}
