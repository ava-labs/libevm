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

package params

import (
	"encoding/json"
	"fmt"

	"github.com/ava-labs/libevm/libevm/pseudo"
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
	switch reg := registeredExtras; {
	case reg.Registered() && !reg.Get().reuseJSONRoot:
		return c.unmarshalJSONWithExtra(data)

	case reg.Registered() && reg.Get().reuseJSONRoot: // although the latter is redundant, it's clearer
		c.extra = reg.Get().newChainConfig()
		if err := UnmarshalChainConfigJSON(data, nil, c.extra); err != nil {
			c.extra = nil
			return err
		}
		fallthrough // Important! We've only unmarshalled the extra field.
	default: // reg == nil
		return UnmarshalChainConfigJSON[struct{}](data, c, nil)
	}
}

// unmarshalJSONWithExtra unmarshals JSON under the assumption that the
// registered [Extras] payload is in the JSON "extra" key. All other
// unmarshalling is performed as if no [Extras] were registered.
func (c *ChainConfig) unmarshalJSONWithExtra(data []byte) error {
	cc := &chainConfigWithExportedExtra{
		chainConfigWithoutMethods: (*chainConfigWithoutMethods)(c),
		Extra:                     registeredExtras.Get().newChainConfig(),
	}
	if err := json.Unmarshal(data, cc); err != nil {
		return err
	}
	c.extra = cc.Extra
	return nil
}

// MarshalJSON implements the [json.Marshaler] interface.
func (c *ChainConfig) MarshalJSON() ([]byte, error) {
	switch reg := registeredExtras; {
	case !reg.Registered():
		return json.Marshal((*chainConfigWithoutMethods)(c))

	case !reg.Get().reuseJSONRoot:
		return c.marshalJSONWithExtra()

	default: // reg.reuseJSONRoot == true
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

// UnmarshalChainConfigJSON JSON decodes the given data according to its arguments:
// - if only `extra` is nil, the data is decoded into `config` and the "extra" JSON
// key field is ignored.
// - if only `config` is nil, the data is decoded into `extra` only.
// - if both `config` and `extra` are non-nil, the data is first decoded into `config`
// and then the "extra" JSON field is decoded into `extra.`
// - if both `config` and `extra` are nil pointers, an error is returned.
func UnmarshalChainConfigJSON[T any](data []byte, config *ChainConfig, extra *T) (err error) {
	switch {
	case config == nil && extra == nil:
		return fmt.Errorf("chain config and extra config are both nil")
	case extra == nil:
		// non-registered extra call from ChainConfig.UnmarshalJSON
		// we only want to decode to the chain config, ignoring the
		// "extra" JSON key.
		err = json.Unmarshal(data, (*chainConfigWithoutMethods)(config))
		if err != nil {
			return fmt.Errorf("decoding chain config without extra: %s", err)
		}
		return nil
	case config == nil:
		// decode the data to the extra argument only.
		// this originates from the registered extra + re-use JSON
		// root call from ChainConfig.UnmarshalJSON.
		err = json.Unmarshal(data, extra)
		if err != nil {
			extra = nil
			return fmt.Errorf("decoding chain config to %T: %s", extra, err)
		}
		return nil
	default:
		// Decode the data separately to the chain config and the extra.
		err = json.Unmarshal(data, (*chainConfigWithoutMethods)(config))
		if err != nil {
			return fmt.Errorf("decoding root chain config: %s", err)
		}

		jsonExtra := struct {
			Extra *T `json:"extra"`
		}{
			Extra: extra,
		}
		err = json.Unmarshal(data, &jsonExtra)
		if err != nil {
			return fmt.Errorf("decoding extra config to %T: %s",
				extra, err)
		}
		return nil
	}
}
