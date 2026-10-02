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

package rpc

import (
	"context"
	"time"
)

// SetCallTimeout limits how long each method call may run on connections
// served by [Server.ServeCodec], such as WebSockets. When a call runs too long,
// its context is cancelled and the client gets a timeout error. The connection
// stays open. A non-positive timeout, the default, means no limit.
//
// It doesn't apply to HTTP. To limit HTTP calls, set a deadline on the
// request's context.
//
// This method should be called before processing any requests via ServeCodec.
func (s *Server) SetCallTimeout(timeout time.Duration) {
	s.services.libevm.callTimeout = timeout
}

// callTimeout returns the timeout set by [Server.SetCallTimeout] for the
// connection that ctx belongs to.
func callTimeout(ctx context.Context) (time.Duration, bool) {
	// The registry is the only server state reachable from ctx.
	c, ok := ctx.Value(clientContextKey{}).(*Client)
	if !ok {
		return 0, false
	}
	d := c.services.libevm.callTimeout
	return d, d > 0
}

// registryExtras holds the libevm additions to [serviceRegistry].
type registryExtras struct {
	callTimeout time.Duration
}
