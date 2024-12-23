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

			expectedEncoded := bytes.NewBuffer(nil)
			err := json.Compact(expectedEncoded, []byte(tt.jsonInput))
			require.NoError(t, err)

			encoded, err := json.Marshal(tt.want)
			require.NoError(t, err, "encoding error")
			require.Equal(t, expectedEncoded.String(), string(encoded))

			decoded := new(ChainConfig)
			err = json.Unmarshal(encoded, decoded)
			require.NoError(t, err, "decoding error")
			require.Equal(t, tt.want, decoded)
		})
	}
}

func TestUnmarshalChainConfigJSON(t *testing.T) {
	t.Parallel()

	type testExtra struct {
		Field string `json:"field"`
	}

	testCases := map[string]struct {
		jsonData       string // string for convenience
		extra          *testExtra
		reuseJSONRoot  bool
		expectedConfig ChainConfig
		expectedExtra  any
		errMessage     string
	}{
		"invalid_json": {
			extra:         &testExtra{},
			expectedExtra: &testExtra{},
			errMessage:    "decoding root chain config: unexpected end of JSON input",
		},
		"nil_extra_at_root_depth": {
			jsonData:       `{"chainId": 1}`,
			extra:          nil,
			reuseJSONRoot:  true,
			expectedConfig: ChainConfig{ChainID: big.NewInt(1)},
			expectedExtra:  (*testExtra)(nil),
			errMessage:     "extra pointer argument is nil",
		},
		"nil_extra_at_extra_key": {
			jsonData:       `{"chainId": 1}`,
			extra:          nil,
			expectedConfig: ChainConfig{ChainID: big.NewInt(1)},
			expectedExtra:  (*testExtra)(nil),
			errMessage:     "extra pointer argument is nil",
		},
		"no_extra_at_extra_key": {
			jsonData:       `{"chainId": 1}`,
			extra:          &testExtra{},
			expectedConfig: ChainConfig{ChainID: big.NewInt(1)},
			expectedExtra:  &testExtra{},
		},
		"no_extra_at_root_depth": {
			jsonData:       `{"chainId": 1}`,
			extra:          &testExtra{},
			reuseJSONRoot:  true,
			expectedConfig: ChainConfig{ChainID: big.NewInt(1)},
			expectedExtra:  &testExtra{},
		},
		"wrong_extra_type_at_extra_key": {
			jsonData:       `{"chainId": 1, "extra": 1}`,
			extra:          &testExtra{},
			expectedConfig: ChainConfig{ChainID: big.NewInt(1)},
			expectedExtra:  &testExtra{},
			errMessage: "decoding extra config to *params.testExtra: " +
				"json: cannot unmarshal number into Go struct field " +
				".extra of type params.testExtra",
		},
		"wrong_extra_type_at_root_depth": {
			jsonData:       `{"chainId": 1, "field": 1}`,
			extra:          &testExtra{},
			reuseJSONRoot:  true,
			expectedConfig: ChainConfig{ChainID: big.NewInt(1)},
			expectedExtra:  &testExtra{},
			errMessage: "decoding extra config to *params.testExtra: " +
				"json: cannot unmarshal number into Go struct field " +
				"testExtra.field of type string",
		},
		"extra_success_at_extra_key": {
			jsonData:       `{"chainId": 1, "extra": {"field":"value"}}`,
			extra:          &testExtra{},
			expectedConfig: ChainConfig{ChainID: big.NewInt(1)},
			expectedExtra:  &testExtra{Field: "value"},
		},
		"extra_success_at_root_depth": {
			jsonData:       `{"chainId": 1, "field":"value"}`,
			extra:          &testExtra{},
			reuseJSONRoot:  true,
			expectedConfig: ChainConfig{ChainID: big.NewInt(1)},
			expectedExtra:  &testExtra{Field: "value"},
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			data := []byte(testCase.jsonData)
			config := ChainConfig{}
			err := UnmarshalChainConfigJSON(data, &config, testCase.extra, testCase.reuseJSONRoot)
			if testCase.errMessage == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, testCase.errMessage)
			}
			assert.Equal(t, testCase.expectedConfig, config)
			assert.Equal(t, testCase.expectedExtra, testCase.extra)
		})
	}
}

func TestMarshalChainConfigJSON(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		config        ChainConfig
		extra         any
		reuseJSONRoot bool
		jsonData      string // string for convenience
		errMessage    string
	}{
		"invalid_extra_at_extra_key": {
			extra: struct {
				Field chan struct{} `json:"field"`
			}{},
			errMessage: "encoding config with extra: " +
				"json: unsupported type: chan struct {}",
		},
		"nil_extra_at_extra_key": {
			jsonData: `{"chainId":null}`,
		},
		"extra_at_extra_key": {
			extra: struct {
				Field string `json:"field"`
			}{Field: "value"},
			jsonData: `{"chainId":null,"extra":{"field":"value"}}`,
		},
		"invalid_extra_at_root_depth": {
			extra: struct {
				Field chan struct{} `json:"field"`
			}{},
			reuseJSONRoot: true,
			errMessage: "converting extra config to JSON raw messages: " +
				"json: unsupported type: chan struct {}",
		},
		"duplicate_key": {
			extra: struct {
				Field string `json:"chainId"`
			}{},
			reuseJSONRoot: true,
			errMessage: `duplicate JSON key "chainId" in ChainConfig` +
				` and extra struct { Field string "json:\"chainId\"" }`,
		},
		"nil_extra_at_root_depth": {
			extra:         nil,
			reuseJSONRoot: true,
			jsonData:      `{"chainId":null}`,
		},
		"extra_at_root_depth": {
			extra: struct {
				Field string `json:"field"`
			}{},
			reuseJSONRoot: true,
			jsonData:      `{"chainId":null,"field":""}`,
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			config := ChainConfig{}
			data, err := MarshalChainConfigJSON(config, testCase.extra, testCase.reuseJSONRoot)
			if testCase.errMessage == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, testCase.errMessage)
			}
			assert.Equal(t, testCase.jsonData, string(data))
		})
	}
}
