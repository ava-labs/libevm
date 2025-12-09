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

	"github.com/ava-labs/libevm/common"
)

// PackEvent packs the given event name and arguments to conform the ABI.
// Returns the topics for the event including the event signature (if non-anonymous event) and
// hashes derived from indexed arguments and the packed data of non-indexed args according to
// the event ABI specification.
// The order of arguments must match the order of the event definition.
// https://docs.soliditylang.org/en/v0.8.17/abi-spec.html#indexed-event-encoding.
// Note: PackEvent does not support array (fixed or dynamic-size) or struct types.
func (abi ABI) PackEvent(name string, args ...interface{}) ([]common.Hash, []byte, error) {
	event, exist := abi.Events[name]
	if !exist {
		return nil, nil, fmt.Errorf("event '%s' not found", name)
	}
	if len(args) != len(event.Inputs) {
		return nil, nil, fmt.Errorf("event '%s' unexpected number of inputs %d", name, len(args))
	}

	var (
		nonIndexedInputs = make([]interface{}, 0)
		indexedInputs    = make([]interface{}, 0)
		nonIndexedArgs   Arguments
		indexedArgs      Arguments
	)

	for i, arg := range event.Inputs {
		if arg.Indexed {
			indexedArgs = append(indexedArgs, arg)
			indexedInputs = append(indexedInputs, args[i])
		} else {
			nonIndexedArgs = append(nonIndexedArgs, arg)
			nonIndexedInputs = append(nonIndexedInputs, args[i])
		}
	}

	packedArguments, err := nonIndexedArgs.Pack(nonIndexedInputs...)
	if err != nil {
		return nil, nil, err
	}
	topics := make([]common.Hash, 0, len(indexedArgs)+1)
	if !event.Anonymous {
		topics = append(topics, event.ID)
	}
	indexedTopics, err := PackTopics(indexedInputs)
	if err != nil {
		return nil, nil, err
	}

	return append(topics, indexedTopics...), packedArguments, nil
}

// PackOutput packs the given [args] as the output of given method [name] to conform the ABI.
// This does not include method ID.
func (abi ABI) PackOutput(name string, args ...interface{}) ([]byte, error) {
	// Fetch the ABI of the requested method
	method, exist := abi.Methods[name]
	if !exist {
		return nil, fmt.Errorf("method '%s' not found", name)
	}
	arguments, err := method.Outputs.Pack(args...)
	if err != nil {
		return nil, err
	}
	return arguments, nil
}
