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
	extra := (*struct{})(nil)
	const reuseJSONRoot = false
	return UnmarshalChainConfigJSON(data, c, extra, reuseJSONRoot)
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

// UnmarshalChainConfigJSON JSON decodes the given data:
//   - if the extra is NOT registered AND the `extra` argument is nil, the chain config
//     is decoded in `config` and the "extra" JSON field is ignored.
//   - if the extra is NOT registered AND the `extra` argument is NOT nil, the chain config
//     is decoded in `config` and the "extra" JSON field is decoded in `extra`.
//   - if the extra is registered, whether the root JSON is re-used for the extra or not,
//     the chain config is decoded only in the `config` argument.
func UnmarshalChainConfigJSON[T any](data []byte, config *ChainConfig, extra *T, reuseJSONRoot bool) (err error) {
	if !registeredExtras.Registered() {
		err = json.Unmarshal(data, (*chainConfigWithoutMethods)(config))
		switch {
		case err != nil:
			return fmt.Errorf("decoding root chain config: %s", err)
		case extra == nil: // ignore the "extra" JSON key
		case reuseJSONRoot:
			err = json.Unmarshal(data, (*chainConfigWithoutMethods)(config))
			if err != nil {
				return fmt.Errorf("decoding chain config: %s", err)
			}

			err = json.Unmarshal(data, extra)
			if err != nil {
				return fmt.Errorf("decoding extra config to %T: %s", config.extra, err)
			}
		default:
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
		}
		return nil
	}

	chainConfigWithoutMethods := (*chainConfigWithoutMethods)(config)

	// registered extra and extra config is in the "extra" JSON field.
	if !registeredExtras.Get().reuseJSONRoot {
		configWrapper := &chainConfigWithExportedExtra{
			chainConfigWithoutMethods: chainConfigWithoutMethods,
			Extra:                     registeredExtras.Get().newChainConfig(),
		}
		err = json.Unmarshal(data, configWrapper)
		if err != nil {
			return fmt.Errorf("decoding chain config and extra: %s", err)
		}
		config.extra = configWrapper.extra
		return nil
	}

	// registered extra and re-use JSON root, the extra config is contained
	// in the root config object.
	err = json.Unmarshal(data, chainConfigWithoutMethods)
	if err != nil {
		return fmt.Errorf("decoding chain config: %s", err)
	}

	config.extra = registeredExtras.Get().newChainConfig()
	err = json.Unmarshal(data, config.extra)
	if err != nil {
		config.extra = nil
		return fmt.Errorf("decoding extra config to %T: %s", config.extra, err)
	}
	return nil
}
