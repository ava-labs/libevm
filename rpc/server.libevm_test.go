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
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/require"
)

func TestServerSetCallTimeout(t *testing.T) {
	tests := []struct {
		name           string
		timeout, sleep time.Duration
		wantErr        error
	}{
		{
			name:  "no_timeout",
			sleep: time.Hour,
		},
		{
			name:    "timeout_longer_than_sleep",
			timeout: time.Second + time.Nanosecond,
			sleep:   time.Second,
		},
		{
			name:    "timeout_shorter_than_sleep",
			timeout: time.Second,
			sleep:   time.Second + time.Nanosecond,
			wantErr: &jsonError{Code: errcodeTimeout, Message: errMsgTimeout},
		},
	}

	transports := []struct {
		name   string
		client func(*testing.T, *Server) *Client
	}{
		{
			// In-process connections are served like WebSockets, by
			// [Server.ServeCodec].
			name: "in-process",
			client: func(t *testing.T, s *Server) *Client {
				t.Helper()
				return DialInProc(s)
			},
		},
		{
			name: "http",
			client: func(t *testing.T, s *Server) *Client {
				t.Helper()
				opt := WithHTTPClient(&http.Client{Transport: handlerTransport{s}})
				c, err := DialOptions(t.Context(), "http://localhost", opt)
				require.NoError(t, err, "DialOptions()")
				return c
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, tr := range transports {
				t.Run(tr.name, func(t *testing.T) {
					synctest.Test(t, func(t *testing.T) { //nolint:thelper // False positive, fixed in thelper v0.7.1.
						srv := newTestServer()
						srv.SetCallTimeout(tt.timeout)
						defer srv.Stop()

						c := tr.client(t, srv)
						t.Cleanup(c.Close)

						err := c.CallContext(t.Context(), nil, "test_sleep", tt.sleep)
						require.Equal(t, tt.wantErr, err, "CallContext(test_sleep)")

						// A timed-out call doesn't close the connection.
						require.NoError(t, c.CallContext(t.Context(), nil, "test_noArgsRets"), "CallContext(test_noArgsRets)")

						// A timed-out call keeps running on the server, so let it return
						// before the bubble ends.
						time.Sleep(tt.sleep)
					})
				})
			}
		})
	}
}

// handlerTransport serves each HTTP request in memory with its Handler, as
// synctest can't advance time while a goroutine waits on a network.
type handlerTransport struct {
	http.Handler
}

func (h handlerTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, r)
	return rec.Result(), nil
}
