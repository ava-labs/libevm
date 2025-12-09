// Copyright 2024-2025 the libevm authors.
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

package abi

import (
	"fmt"
	"math/big"
	"reflect"

	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/libevm/common/math"
	"github.com/ava-labs/libevm/crypto"
)

// packTopic packs rule into the corresponding hash value for a log's topic
// according to the Solidity documentation:
// https://docs.soliditylang.org/en/v0.8.17/abi-spec.html#indexed-event-encoding.
func packTopic(rule interface{}) (common.Hash, error) {
	var topic common.Hash

	// Try to generate the topic based on simple types
	switch rule := rule.(type) {
	case common.Hash:
		copy(topic[:], rule[:])
	case common.Address:
		copy(topic[common.HashLength-common.AddressLength:], rule[:])
	case *big.Int:
		copy(topic[:], math.U256Bytes(rule))
	case bool:
		if rule {
			topic[common.HashLength-1] = 1
		}
	case int8:
		copy(topic[:], genIntType(int64(rule), 1))
	case int16:
		copy(topic[:], genIntType(int64(rule), 2))
	case int32:
		copy(topic[:], genIntType(int64(rule), 4))
	case int64:
		copy(topic[:], genIntType(rule, 8))
	case uint8:
		blob := new(big.Int).SetUint64(uint64(rule)).Bytes()
		copy(topic[common.HashLength-len(blob):], blob)
	case uint16:
		blob := new(big.Int).SetUint64(uint64(rule)).Bytes()
		copy(topic[common.HashLength-len(blob):], blob)
	case uint32:
		blob := new(big.Int).SetUint64(uint64(rule)).Bytes()
		copy(topic[common.HashLength-len(blob):], blob)
	case uint64:
		blob := new(big.Int).SetUint64(rule).Bytes()
		copy(topic[common.HashLength-len(blob):], blob)
	case string:
		hash := crypto.Keccak256Hash([]byte(rule))
		copy(topic[:], hash[:])
	case []byte:
		hash := crypto.Keccak256Hash(rule)
		copy(topic[:], hash[:])

	default:
		// todo(rjl493456442) according to solidity documentation, indexed event
		// parameters that are not value types i.e. arrays and structs are not
		// stored directly but instead a keccak256-hash of an encoding is stored.
		//
		// We only convert strings and bytes to hash, still need to deal with
		// array(both fixed-size and dynamic-size) and struct.

		// Attempt to generate the topic from funky types
		val := reflect.ValueOf(rule)
		switch {
		// static byte array
		case val.Kind() == reflect.Array && reflect.TypeOf(rule).Elem().Kind() == reflect.Uint8:
			reflect.Copy(reflect.ValueOf(topic[:val.Len()]), val)
		default:
			return common.Hash{}, fmt.Errorf("unsupported indexed type: %T", rule)
		}
	}
	return topic, nil
}

// PackTopics packs the array of filters into an array of corresponding topics
// according to the Solidity documentation.
// Note: PackTopics does not support array (fixed or dynamic-size) or struct types.
func PackTopics(filter []interface{}) ([]common.Hash, error) {
	topics := make([]common.Hash, len(filter))
	for i, rule := range filter {
		topic, err := packTopic(rule)
		if err != nil {
			return nil, err
		}
		topics[i] = topic
	}

	return topics, nil
}
