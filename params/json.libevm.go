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
)

var _ interface {
	json.Marshaler
	json.Unmarshaler
} = (*ChainConfig)(nil)

// chainConfigWithoutMethods avoids infinite recurion into
// [ChainConfig.UnmarshalJSON].
type chainConfigWithoutMethods ChainConfig

// UnmarshalJSON implements the [json.Unmarshaler] interface and JSON decodes
// `data` according to the following:
//   - extra is not registered:
//     `data` is decoded into `c` and the extra is ignored.
//   - extra is registered and the registered reuseJSONRoot field is false:
//     `data` is decoded into `c` and the "extra" JSON field in `data` is decoded into `c.extra`.
//   - extra is registered and the registered reuseJSONRoot field is true:
//     `data` is decoded into `c` and `data` is decoded into `c.extra`.
func (c *ChainConfig) UnmarshalJSON(data []byte) (err error) {
	if !registeredExtras.Registered() {
		// assume there is no extra
		return json.Unmarshal(data, (*chainConfigWithoutMethods)(c))
	}
	extraConstructors := registeredExtras.Get()
	c.extra = extraConstructors.newChainConfig()
	reuseJSONRoot := extraConstructors.reuseJSONRoot
	return UnmarshalChainConfigJSON(data, c, c.extra, reuseJSONRoot)
}

// UnmarshalChainConfigJSON JSON decodes `data` according to the following.
//   - `reuseJSONRoot` is false:
//     `data` is decoded into `config` and the "extra" JSON field in `data` is decoded into `extra`.
//   - `reuseJSONRoot` is true:
//     `data` is decoded into `config` and `data` is decoded into `extra`.
func UnmarshalChainConfigJSON[T any](data []byte, config *ChainConfig, extra *T, reuseJSONRoot bool) (err error) {
	if extra == nil {
		return fmt.Errorf("extra pointer argument is nil")
	}
	err = json.Unmarshal(data, (*chainConfigWithoutMethods)(config))
	switch {
	case err != nil:
		return fmt.Errorf("decoding root chain config: %s", err)
	case reuseJSONRoot:
		err = json.Unmarshal(data, extra)
		if err != nil {
			return fmt.Errorf("decoding extra config to %T: %s", extra, err)
		}
	default:
		jsonExtra := struct {
			Extra *T `json:"extra"`
		}{
			Extra: extra,
		}
		err = json.Unmarshal(data, &jsonExtra)
		if err != nil {
			return fmt.Errorf("decoding extra config to %T: %s", extra, err)
		}
	}
	return nil
}

// MarshalJSON implements the [json.Marshaler] interface and JSON encodes
// the chain config `c` according to the following:
//   - extra is not registered:
//     `c` is encoded into `data` and the extra is ignored.
//   - extra is registered and the registered reuseJSONRoot field is false:
//     `c` is encoded with `c.extra` encoded at the "extra" JSON field.
//   - extra is registered and the registered reuseJSONRoot field is true:
//     `c` is encoded with `c.extra` encoded at the root depth of the JSON object.
func (c *ChainConfig) MarshalJSON() ([]byte, error) {
	if !registeredExtras.Registered() {
		// assume there is no extra
		return json.Marshal((*chainConfigWithoutMethods)(c))
	}
	extraConstructors := registeredExtras.Get()
	reuseJSONRoot := extraConstructors.reuseJSONRoot
	return MarshalChainConfigJSON(*c, c.extra, reuseJSONRoot)
}

// MarshalChainConfigJSON JSON encodes `config` and `extra` according to the following.
//   - `reuseJSONRoot` is false:
//     `config` is encoded with `extra` encoded at the "extra" JSON field.
//   - `reuseJSONRoot` is true:
//     `config` is encoded with `extra` encoded at the root depth of the JSON object.
func MarshalChainConfigJSON[T any](config ChainConfig, extra T, reuseJSONRoot bool) (data []byte, err error) {
	if !reuseJSONRoot {
		jsonExtra := struct {
			ChainConfig
			Extra T `json:"extra,omitempty"`
		}{
			ChainConfig: config,
			Extra:       extra,
		}
		data, err = json.Marshal(jsonExtra)
		if err != nil {
			return nil, fmt.Errorf("encoding config with extra: %s", err)
		}
		return data, nil
	}

	// The inverse of reusing the JSON root is merging two JSON buffers,
	// which isn't supported by the native package. So we use
	// map[string]json.RawMessage intermediates.
	// Note we cannot encode a combined struct directly because of the extra
	// type generic nature which cannot be embedded in such a combined struct.
	configJSONRaw, err := toJSONRawMessages((chainConfigWithoutMethods)(config))
	if err != nil {
		return nil, fmt.Errorf("converting config to JSON raw messages: %s", err)
	}
	extraJSONRaw, err := toJSONRawMessages(extra)
	if err != nil {
		return nil, fmt.Errorf("converting extra config to JSON raw messages: %s", err)
	}

	for k, v := range extraJSONRaw {
		_, ok := configJSONRaw[k]
		if ok {
			return nil, fmt.Errorf("duplicate JSON key %q in ChainConfig and extra %T", k, extra)
		}
		configJSONRaw[k] = v
	}
	return json.Marshal(configJSONRaw)
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
