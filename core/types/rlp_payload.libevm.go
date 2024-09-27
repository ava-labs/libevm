package types

import (
	"io"

	"github.com/ethereum/go-ethereum/libevm/pseudo"
	"github.com/ethereum/go-ethereum/rlp"
)

type RLPPayload struct {
	t *pseudo.Type
}

func NewRLPPayload[T any]() (*RLPPayload, *pseudo.Value[T]) {
	var x T
	return RLPPayloadOf(x)
}

func RLPPayloadOf[T any](x T) (*RLPPayload, *pseudo.Value[T]) {
	p := pseudo.From(x)
	return &RLPPayload{p.Type}, p.Value
}

var _ interface {
	rlp.Encoder
	rlp.Decoder
} = (*RLPPayload)(nil)

func (p *RLPPayload) EncodeRLP(w io.Writer) error {
	if p == nil || p.t == nil {
		return nil
	}
	return p.t.EncodeRLP(w)
}

func (p *RLPPayload) DecodeRLP(s *rlp.Stream) error {
	// DO NOT MERGE without implementation
	return nil
}
