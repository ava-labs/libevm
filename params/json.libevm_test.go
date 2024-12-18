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
	"bytes"
	"encoding/json"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ava-labs/libevm/libevm/pseudo"
)

type nestedChainConfigExtra struct {
	NestedFoo string `json:"foo"`

	NOOPHooks
}

type rootJSONChainConfigExtra struct {
	TopLevelFoo string `json:"foo"`

	NOOPHooks
}

func TestChainConfigJSONRoundTrip(t *testing.T) {
	tests := []struct {
		name      string
		register  func()
		jsonInput string
		want      *ChainConfig
	}{
		{
			name:     "no registered extras",
			register: func() {},
			jsonInput: `{
				"chainId": 1234
			}`,
			want: &ChainConfig{
				ChainID: big.NewInt(1234),
			},
		},
		{
			name: "reuse top-level JSON with non-pointer",
			register: func() {
				RegisterExtras(Extras[rootJSONChainConfigExtra, NOOPHooks]{
					ReuseJSONRoot: true,
				})
			},
			jsonInput: `{
				"chainId": 5678,
				"foo": "hello"
			}`,
			want: &ChainConfig{
				ChainID: big.NewInt(5678),
				extra:   pseudo.From(rootJSONChainConfigExtra{TopLevelFoo: "hello"}).Type,
			},
		},
		{
			name: "reuse top-level JSON with pointer",
			register: func() {
				RegisterExtras(Extras[*rootJSONChainConfigExtra, NOOPHooks]{
					ReuseJSONRoot: true,
				})
			},
			jsonInput: `{
				"chainId": 5678,
				"foo": "hello"
			}`,
			want: &ChainConfig{
				ChainID: big.NewInt(5678),
				extra:   pseudo.From(&rootJSONChainConfigExtra{TopLevelFoo: "hello"}).Type,
			},
		},
		{
			name: "nested JSON with non-pointer",
			register: func() {
				RegisterExtras(Extras[nestedChainConfigExtra, NOOPHooks]{
					ReuseJSONRoot: false, // explicit zero value only for tests
				})
			},
			jsonInput: `{
				"chainId": 42,
				"extra": {"foo": "world"}
			}`,
			want: &ChainConfig{
				ChainID: big.NewInt(42),
				extra:   pseudo.From(nestedChainConfigExtra{NestedFoo: "world"}).Type,
			},
		},
		{
			name: "nested JSON with pointer",
			register: func() {
				RegisterExtras(Extras[*nestedChainConfigExtra, NOOPHooks]{
					ReuseJSONRoot: false, // explicit zero value only for tests
				})
			},
			jsonInput: `{
				"chainId": 42,
				"extra": {"foo": "world"}
			}`,
			want: &ChainConfig{
				ChainID: big.NewInt(42),
				extra:   pseudo.From(&nestedChainConfigExtra{NestedFoo: "world"}).Type,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			TestOnlyClearRegisteredExtras()
			t.Cleanup(TestOnlyClearRegisteredExtras)
			tt.register()

			t.Run("json.Unmarshal()", func(t *testing.T) {
				got := new(ChainConfig)
				require.NoError(t, json.Unmarshal([]byte(tt.jsonInput), got))
				require.Equal(t, tt.want, got)
			})

			t.Run("json.Marshal()", func(t *testing.T) {
				var want bytes.Buffer
				require.NoError(t, json.Compact(&want, []byte(tt.jsonInput)), "json.Compact()")

				got, err := json.Marshal(tt.want)
				require.NoError(t, err, "json.Marshal()")
				require.Equal(t, want.String(), string(got))
			})
		})
	}
}

func Test_UnmarshalChainConfigJSON(t *testing.T) {
	t.Parallel()

	type testExtra struct {
		Field string `json:"field"`
	}

	testCases := map[string]struct {
		jsonData       string // string for convenience
		expectedConfig ChainConfig
		expectedExtra  any
		errMessage     string
	}{
		"invalid_json": {
			expectedExtra: testExtra{},
			errMessage:    "json decoding root chain config: unexpected end of JSON input",
		},
		"no_extra": {
			jsonData:       `{"chainId": 1}`,
			expectedConfig: ChainConfig{ChainID: big.NewInt(1)},
			expectedExtra:  testExtra{},
		},
		"wrong_extra_type": {
			jsonData:       `{"chainId": 1, "extra": 1}`,
			expectedConfig: ChainConfig{ChainID: big.NewInt(1)},
			expectedExtra:  testExtra{},
			errMessage: "json decoding extra chain config: " +
				"json: cannot unmarshal number into Go struct field " +
				".extra of type params.testExtra",
		},
		"extra_success": {
			jsonData:       `{"chainId": 1, "extra": {"field":"value"}}`,
			expectedConfig: ChainConfig{ChainID: big.NewInt(1)},
			expectedExtra:  testExtra{Field: "value"},
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			data := []byte(testCase.jsonData)
			config := ChainConfig{}
			var extra testExtra
			err := UnmarshalChainConfigJSON(data, &config, &extra)
			if testCase.errMessage == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, testCase.errMessage)
			}
			assert.Equal(t, testCase.expectedConfig, config)
			assert.Equal(t, testCase.expectedExtra, extra)
		})
	}
}
