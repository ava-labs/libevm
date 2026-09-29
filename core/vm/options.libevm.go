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

package vm

import "github.com/ava-labs/libevm/libevm/options"

type callConfig struct {
	unsafeCallerAddressProxying bool
	legacyOutboundCallGas       bool
}

// A CallOption modifies the default behaviour of a contract call.
type CallOption = options.Option[callConfig]

// WithUNSAFECallerAddressProxying results in precompiles making contract calls
// specifying their own caller's address as the caller. This is NOT SAFE for
// regular use as callers of the precompile may not understand that they are
// escalating the precompile's privileges.
//
// Deprecated: this option MUST NOT be used other than to allow migration to
// libevm when backwards compatibility is required.
func WithUNSAFECallerAddressProxying() CallOption {
	return options.Func[callConfig](func(c *callConfig) {
		c.unsafeCallerAddressProxying = true
	})
}

// WithLegacyOutboundCallGas disables all constant- and dynamic-gas charges, as
// well as call stipends, the EIP-150 63/64 rule, and callee address warming for
// this call. The gas charged to the caller and the gas received by the callee
// are identical.
//
// Deprecated: only for backwards compatibility with historical chain behaviour
// (e.g. legacy native-asset precompile semantics). New precompiles MUST NOT use
// this option.
func WithLegacyOutboundCallGas() CallOption {
	return options.Func[callConfig](func(c *callConfig) {
		c.legacyOutboundCallGas = true
	})
}
