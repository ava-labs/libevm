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
		newStateAccount: pseudo.NewConstructor[SA]().Zero,
	}
}

var registeredExtras *extraConstructors

type extraConstructors struct {
	newStateAccount func() *pseudo.Type
}

type StateAccountExtra struct {
	t *pseudo.Type
}

var _ interface {
	rlp.Encoder
	rlp.Decoder
} = (*StateAccountExtra)(nil)

func (p *StateAccountExtra) EncodeRLP(w io.Writer) error {
	switch r := registeredExtras; {
	case r == nil:
		return nil
	case p == nil:
		p = &StateAccountExtra{}
		fallthrough
	case p.t == nil:
		p.t = r.newStateAccount()
	}
	return p.t.EncodeRLP(w)
}

func (p *StateAccountExtra) DecodeRLP(s *rlp.Stream) error {
	// DO NOT MERGE without implementation
	return nil
}
