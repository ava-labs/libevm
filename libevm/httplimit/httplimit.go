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

// Package httplimit provides a single, shared bound on the size of HTTP request
// bodies accepted by handlers that decode untrusted input.
//
// It exists so that the limit is defined in exactly one place. Handlers that
// decode a request body without bounding it first can be driven to exhaust the
// process's memory by a single oversized request.
package httplimit

import "net/http"

// MaxBodySize is the maximum accepted size, in bytes, of an HTTP request body.
//
// It matches the default of the JSON-RPC server's own limit (see
// `rpc.Server.SetHTTPBodyLimit`), which is configured separately because it is
// upstream go-ethereum code and is tunable per server.
const MaxBodySize = 5 * 1024 * 1024

// LimitBody bounds `r.Body` to [MaxBodySize] and reports whether the request
// MAY proceed. It MUST be called before decoding the body.
//
// If the request declares a Content-Length exceeding the limit then a 413 is
// written to `w` and false is returned, in which case the caller MUST NOT write
// a further response. Otherwise `r.Body` is replaced with an equivalent reader
// that stops short of the limit, so an oversized body that lies about (or omits)
// its length surfaces as a read error from the decoder.
func LimitBody(w http.ResponseWriter, r *http.Request) bool {
	if r.ContentLength > MaxBodySize {
		http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
		return false
	}
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodySize)
	return true
}
