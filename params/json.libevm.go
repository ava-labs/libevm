package params

import (
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum/libevm/pseudo"
)

var _ interface {
	json.Marshaler
	json.Unmarshaler
} = (*ChainConfig)(nil)

// chainConfigWithoutMethods avoids infinite recurion into
// [ChainConfig.UnmarshalJSON].
type chainConfigWithoutMethods ChainConfig

// chainConfigWithExportedExtra supports JSON (un)marshalling of a [ChainConfig]
// while exposing the `extra` field as the "extra" JSON key.
type chainConfigWithExportedExtra struct {
	*chainConfigWithoutMethods              // embedded to achieve regular JSON unmarshalling
	Extra                      *pseudo.Type `json:"extra"` // `c.extra` is otherwise unexported
}

// UnmarshalJSON implements the [json.Unmarshaler] interface.
func (c *ChainConfig) UnmarshalJSON(data []byte) error {
	extras := registeredExtras

	if extras != nil && !extras.reuseJSONRoot {
		return c.unmarshalJSONWithExtra(data)
	}

	if err := json.Unmarshal(data, (*chainConfigWithoutMethods)(c)); err != nil {
		return err
	}
	if extras == nil {
		return nil
	}

	// Invariants if here:
	// - reg.reuseJSONRoot == true
	// - Non-extra ChainConfig fields already unmarshalled

	c.extra = extras.chainConfig.NilPointer()
	if err := json.Unmarshal(data, c.extra); err != nil {
		c.extra = nil
		return err
	}
	return nil
}

// unmarshalJSONWithExtra unmarshals JSON under the assumption that the
// registered [Extras] payload is in the JSON "extra" key. All other
// unmarshalling is performed as if no [Extras] were registered.
func (c *ChainConfig) unmarshalJSONWithExtra(data []byte) error {
	cc := &chainConfigWithExportedExtra{
		chainConfigWithoutMethods: (*chainConfigWithoutMethods)(c),
		Extra:                     registeredExtras.chainConfig.NilPointer(),
	}
	if err := json.Unmarshal(data, cc); err != nil {
		return err
	}
	c.extra = cc.Extra
	return nil
}

// MarshalJSON implements the [json.Marshaler] interface.
func (c *ChainConfig) MarshalJSON() ([]byte, error) {
	switch extras := registeredExtras; {
	case extras == nil:
		return json.Marshal((*chainConfigWithoutMethods)(c))

	case !extras.reuseJSONRoot:
		return c.marshalJSONWithExtra()

	default:
		// The inverse of reusing the JSON root is merging two JSON buffers,
		// which isn't supported by the native package. So we use
		// map[string]json.RawMessage intermediates.
		geth, err := toJSONRawMessages((*chainConfigWithoutMethods)(c))
		if err != nil {
			return nil, err
		}
		extra, err := toJSONRawMessages(c.extra)
		if err != nil {
			return nil, err
		}

		for k, v := range extra {
			if _, ok := geth[k]; ok {
				return nil, fmt.Errorf("duplicate JSON key %q in both %T and registered extra", k, c)
			}
			geth[k] = v
		}
		return json.Marshal(geth)
	}
}

// marshalJSONWithExtra is the inverse of unmarshalJSONWithExtra().
func (c *ChainConfig) marshalJSONWithExtra() ([]byte, error) {
	cc := &chainConfigWithExportedExtra{
		chainConfigWithoutMethods: (*chainConfigWithoutMethods)(c),
		Extra:                     c.extra,
	}
	return json.Marshal(cc)
}

func toJSONRawMessages(v any) (map[string]json.RawMessage, error) {
	buf, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	msgs := make(map[string]json.RawMessage)
	if err := json.Unmarshal(buf, &msgs); err != nil {
		return nil, err
	}
	return msgs, nil
}
