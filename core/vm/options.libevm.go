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

// A CallOption modifies the default behaviour of a contract call.
type CallOption interface {
	libevmCallOption() // noop to only allow internally defined options
}

// WithUNSAFEForceDelegate results in precompiles making contract calls acting
// as if they themselves were DELEGATECALLed. This is not safe for regular use
// as the precompile will act as its own caller even when not expected to.
//
// Deprecated: this option MUST NOT be used other than to allow migration to
// libevm when backwards compatibility is required.
func WithUNSAFEForceDelegate() CallOption {
	return callOptForceDelegate{}
}

type callOptForceDelegate struct{}

func (callOptForceDelegate) libevmCallOption() {}
