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

package types

import (
	"io"

	"github.com/ethereum/go-ethereum/libevm/pseudo"
	"github.com/ethereum/go-ethereum/rlp"
)

type Extras[SA any] struct{}

func RegisterExtras[SA any](extras Extras[SA]) {
	if registeredExtras != nil {
		panic("re-registration of Extras")
	}
	registeredExtras = &extraConstructors{
		newStateAccount:   pseudo.NewConstructor[SA]().Zero,
		cloneStateAccount: extras.cloneStateAccount,
	}
}

func (e Extras[SA]) cloneStateAccount(s *StateAccountExtra) *StateAccountExtra {
	v := pseudo.MustNewValue[SA](s.t)
	return &StateAccountExtra{
		t: pseudo.From(v.Get()).Type,
	}
}

var registeredExtras *extraConstructors

type extraConstructors struct {
	newStateAccount   func() *pseudo.Type
	cloneStateAccount func(*StateAccountExtra) *StateAccountExtra
}

type StateAccountExtra struct {
	t *pseudo.Type
}

func (e *StateAccountExtra) clone() *StateAccountExtra {
	switch r := registeredExtras; {
	case r == nil, e == nil:
		return nil
	default:
		return r.cloneStateAccount(e)
	}
}

var _ interface {
	rlp.Encoder
	rlp.Decoder
} = (*StateAccountExtra)(nil)

func (e *StateAccountExtra) EncodeRLP(w io.Writer) error {
	switch r := registeredExtras; {
	case r == nil:
		return nil
	case e == nil:
		e = &StateAccountExtra{}
		fallthrough
	case e.t == nil:
		e.t = r.newStateAccount()
	}
	return e.t.EncodeRLP(w)
}

func (e *StateAccountExtra) DecodeRLP(s *rlp.Stream) error {
	// DO NOT MERGE without implementation
	return nil
}
