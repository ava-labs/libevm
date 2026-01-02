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

package filters

import "github.com/ava-labs/libevm/libevm/options"

// A NewFilterAPIOption configures the behaviour of [NewFilterAPI].
type NewFilterAPIOption = options.Option[newFilterAPIConfig]

type newFilterAPIConfig struct {
	quit chan struct{}
}

// WithQuitter sets a channel that can be closed to signal the FilterAPI to
// quit.
func WithQuitter(quit chan struct{}) NewFilterAPIOption {
	return options.Func[newFilterAPIConfig](func(c *newFilterAPIConfig) {
		c.quit = quit
	})
}
