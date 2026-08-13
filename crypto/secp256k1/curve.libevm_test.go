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

//go:build !gofuzz && cgo
// +build !gofuzz,cgo

package secp256k1

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"math/big"
	"testing"
)

// TestUnmarshalValidatesPoint demonstrates that [BitCurve.Unmarshal] rejects
// points that aren't valid secp256k1 points. Without the check, a point on a
// different curve of the same form could reach [BitCurve.ScalarMult] and expose
// an invalid-curve attack on the scalar.
//
// `elliptic.Unmarshal()` is also exercised because it is what the ECIES
// decryption path calls. It currently performs its own equivalent validation for
// this curve, since dispatching to a curve's own parser requires both
// Unmarshal() and UnmarshalCompressed(); this test would catch the dispatch
// flipping if UnmarshalCompressed() were ever added.
func TestUnmarshalValidatesPoint(t *testing.T) {
	curve := S256()

	valid, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		t.Fatalf("ecdsa.GenerateKey(S256(), ...) error %v", err)
	}

	tests := []struct {
		name    string
		x, y    *big.Int
		wantNil bool
	}{
		{
			name: "valid point",
			x:    valid.X,
			y:    valid.Y,
		},
		{
			// y² = x³ + 7 has no solution at x=1 (8 is not a quadratic residue
			// mod P), so (1,1) is on some other curve of the same form.
			name:    "off-curve point",
			x:       big.NewInt(1),
			y:       big.NewInt(1),
			wantNil: true,
		},
		{
			name:    "all zeroes (point at infinity)",
			x:       big.NewInt(0),
			y:       big.NewInt(0),
			wantNil: true,
		},
		{
			// A coordinate MUST be reduced modulo P. Note that `valid.X+P`
			// isn't usable here as it wouldn't fit in the fixed-width encoding.
			name:    "x not reduced modulo P",
			x:       curve.P,
			y:       valid.Y,
			wantNil: true,
		},
		{
			name:    "y not reduced modulo P",
			x:       valid.X,
			y:       curve.P,
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := curve.Marshal(tt.x, tt.y)

			unmarshalers := map[string]func([]byte) (*big.Int, *big.Int){
				"BitCurve.Unmarshal": curve.Unmarshal,
				"elliptic.Unmarshal": func(b []byte) (*big.Int, *big.Int) {
					return elliptic.Unmarshal(curve, b)
				},
			}

			for name, unmarshal := range unmarshalers {
				gotX, gotY := unmarshal(buf)

				if tt.wantNil {
					if gotX != nil || gotY != nil {
						t.Errorf("%s(Marshal(%v, %v)) got (%v, %v); want (nil, nil)", name, tt.x, tt.y, gotX, gotY)
					}
					continue
				}
				if gotX.Cmp(tt.x) != 0 || gotY.Cmp(tt.y) != 0 {
					t.Errorf("%s(Marshal(x, y)) got (%v, %v); want (%v, %v)", name, gotX, gotY, tt.x, tt.y)
				}
			}
		})
	}
}
