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

// Package p256verify implements an EVM precompile to verify P256 ECDSA
// signatures, as described in RIP-7212.
package p256verify

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"math/big"
)

// Precompile implements ECDSA verification on the P256 curve, as defined by
// [RIP-7212].
//
// [RIP-7212]: https://github.com/ethereum/RIPs/blob/1f55794f65caa4c4bb2b8d9bda7d713b8c734157/RIPS/rip-7212.md
type Precompile struct{}

// RequiredGas always returns 3450.
func (Precompile) RequiredGas([]byte) uint64 { return 3450 }

const inputLen = 160

type input [inputLen]byte

// Run parses and verifies the signature. On success it returns a 32-byte
// big-endian representation of the number 1, otherwise it returns an empty
// slice. The returned error is always nil.
func (Precompile) Run(sig []byte) ([]byte, error) {
	if len(sig) != inputLen || !(*input)(sig).verify() {
		return nil, nil
	}
	return []byte{31: 1}, nil
}

func (in *input) verify() bool {
	key, ok := in.pubkey()
	if !ok {
		return false
	}
	return ecdsa.Verify(key, in.word(0), in.bigWord(1), in.bigWord(2))
}

func (in *input) pubkey() (*ecdsa.PublicKey, bool) {
	x := in.bigWord(3)
	y := in.bigWord(4)
	if x.Sign() == 0 && y.Sign() == 0 {
		return nil, false
	}

	curve := elliptic.P256()
	if !curve.IsOnCurve(x, y) {
		return nil, false
	}
	return &ecdsa.PublicKey{
		Curve: curve,
		X:     x,
		Y:     y,
	}, true
}

func (in *input) word(index int) []byte {
	s := index * 32
	return in[s : s+32]
}

func (in *input) bigWord(index int) *big.Int {
	return new(big.Int).SetBytes(in.word(index))
}

// Sign signs `hash` with the private key, using [rand.Reader] as the first
// argument to [ecdsa.Sign]. It returns a signature payload constructed with
// [Pack], which can therefore be passed directly to the precompile.
func Sign(priv *ecdsa.PrivateKey, hash [32]byte) ([]byte, error) {
	r, s, err := ecdsa.Sign(rand.Reader, priv, hash[:])
	if err != nil {
		return nil, err
	}
	return Pack(hash, r, s, &priv.PublicKey), nil
}

// Pack packs the arguments into a byte slice compatible with [Precompile.Run].
// It does NOT perform any validation on its inputs and therefore may panic if,
// for example, a [big.Int] with >256 bits is received. Keys and signatures
// generated with [elliptic.GenerateKey] and [ecdsa.Sign] are valid inputs.
func Pack(hash [32]byte, r, s *big.Int, key *ecdsa.PublicKey) []byte {
	var in input
	copy(in.word(0), hash[:])
	r.FillBytes(in.word(1))
	s.FillBytes(in.word(2))
	key.X.FillBytes(in.word(3))
	key.Y.FillBytes(in.word(4))
	return in[:]
}
