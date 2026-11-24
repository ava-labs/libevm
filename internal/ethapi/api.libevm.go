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
	"maps"
	"math/big"
	"slices"

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
		callOpts       []CallOption
		accessListOpts []AccessListOption
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

type (
	// An AccessListOption configures [AccessList].
	AccessListOption = options.Option[accessListConfig]

	accessListConfig struct {
		verbatim map[common.Address]struct{}
	}
)

// WithDefaultAccessListOptions returns an option to configure a
// [BlockChainAPI] with default options to be passed to [AccessList] by
// [BlockChainAPI.CreateAccessList]. Repeated options result in a concatenation
// of the arguments to each.
func WithDefaultAccessListOptions(opts ...AccessListOption) BlockChainAPIOption {
	return options.Func[blockChainAPIConfig](func(c *blockChainAPIConfig) {
		c.accessListOpts = append(c.accessListOpts, opts...)
	})
}

// WithVerbatimAccessTuples returns an option to configure [AccessList] to
// return the caller's access tuples for the given addresses unchanged and in
// their original order, and to not generate tuples for them when they are
// called. It is intended for precompiles whose storage keys carry data rather
// than storage slots, which the access-list tracer would otherwise reorder and
// deduplicate.
func WithVerbatimAccessTuples(addrs ...common.Address) AccessListOption {
	return options.Func[accessListConfig](func(c *accessListConfig) {
		if c.verbatim == nil {
			c.verbatim = make(map[common.Address]struct{})
		}
		for _, a := range addrs {
			c.verbatim[a] = struct{}{}
		}
	})
}

// tracerExclusions returns the precompiles and the verbatim addresses, which the
// access-list tracer excludes from the access lists that it generates.
func (c *accessListConfig) tracerExclusions(precompiles []common.Address) []common.Address {
	return slices.Concat(precompiles, slices.Collect(maps.Keys(c.verbatim)))
}

// split returns the tuples of acl for the access-list tracer, and the verbatim
// tuples, in their original order.
func (c *accessListConfig) split(acl *types.AccessList) (traced, verbatim types.AccessList) {
	if acl == nil {
		return nil, nil
	}
	for _, t := range *acl {
		if _, ok := c.verbatim[t.Address]; ok {
			verbatim = append(verbatim, t)
		} else {
			traced = append(traced, t)
		}
	}
	return traced, verbatim
}

// withVerbatim returns acl followed by the verbatim tuples.
func withVerbatim(acl, verbatim types.AccessList) types.AccessList {
	if len(verbatim) == 0 {
		return acl
	}
	return slices.Concat(acl, verbatim)
}
