// Copyright 2026 the libevm authors.
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

package ethapi

import (
	"math/big"

	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/libevm/core"
	"github.com/ava-labs/libevm/core/types"
	"github.com/ava-labs/libevm/libevm/options"
	"github.com/ava-labs/libevm/params"
)

// NewRPCTransaction exports the [newRPCTransaction] function.
func NewRPCTransaction(tx *types.Transaction, blockHash common.Hash, blockNumber uint64, blockTime uint64, index uint64, baseFee *big.Int, config *params.ChainConfig) *RPCTransaction {
	return newRPCTransaction(tx, blockHash, blockNumber, blockTime, index, baseFee, config)
}

// MarshalReceipt exports the [marshalReceipt] function.
func MarshalReceipt(r *types.Receipt, blockHash common.Hash, blockNumber uint64, signer types.Signer, tx *types.Transaction, txIndex int) map[string]any {
	return marshalReceipt(r, blockHash, blockNumber, signer, tx, txIndex)
}

// AccessListResult exports the [accessListResult] type.
type AccessListResult = accessListResult

// RevertError exports the [revertError] type.
type RevertError = revertError

// NewRevertError exports the [newRevertError] constructor.
func NewRevertError(revert []byte) *RevertError {
	return newRevertError(revert)
}

type (
	// A BlockChainAPIOption configures a [BlockChainAPI].
	BlockChainAPIOption = options.Option[blockChainAPIConfig]

	blockChainAPIConfig struct {
		callOpts []CallOption
	}

	// A CallOption configures [DoCall].
	CallOption = options.Option[callConfig]

	callConfig struct {
		interceptResult CallResultInterceptor
	}

	// A CallResultInterceptor receives the return argument of [DoCall] before
	// it is returned. If the interceptor returns an error then it is propagated
	// along with a nil [core.ExecutionResult].
	CallResultInterceptor func(*core.ExecutionResult) error
)

// WithDefaultCallOptions returns an option to configure a [BlockChainAPI] with
// default options to be passed to [DoCall] by [BlockChainAPI.Call]. Repeated
// options result in a concatenation of the arguments to each.
func WithDefaultCallOptions(opts ...CallOption) BlockChainAPIOption {
	return options.Func[blockChainAPIConfig](func(c *blockChainAPIConfig) {
		c.callOpts = append(c.callOpts, opts...)
	})
}

// WithCallResultInterceptor returns an option to configure [DoCall] with the
// provided interceptor.
func WithCallResultInterceptor(fn CallResultInterceptor) CallOption {
	return options.Func[callConfig](func(c *callConfig) {
		c.interceptResult = fn
	})
}

func interceptCallResult(r *core.ExecutionResult, err error, opts ...CallOption) (*core.ExecutionResult, error) {
	if err != nil {
		return r, err
	}
	fn := options.As[callConfig](opts...).interceptResult
	if fn == nil {
		return r, nil
	}
	if err := fn(r); err != nil {
		return nil, err
	}
	return r, nil
}
