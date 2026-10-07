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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) { //nolint:thelper // False positive, fixed in thelper v0.7.1.
				srv := newTestServer()
				srv.SetCallTimeout(tt.timeout)
				defer srv.Stop()
				// In-process connections are served like WebSockets, by
				// [Server.ServeCodec].
				inProc := DialInProc(srv)
				defer inProc.Close()
				overHTTP, err := DialHTTPWithClient("http://localhost", &http.Client{Transport: handlerTransport{srv}})
				require.NoError(t, err, "DialHTTPWithClient()")
				defer overHTTP.Close()

				for name, client := range map[string]*Client{"in_proc": inProc, "http": overHTTP} {
					err := client.CallContext(t.Context(), nil, "test_sleep", tt.sleep)
					require.Equalf(t, tt.wantErr, err, "%s CallContext(test_sleep)", name)

					// A timed-out call doesn't close the connection.
					require.NoErrorf(t, client.CallContext(t.Context(), nil, "test_noArgsRets"), "%s CallContext(test_noArgsRets)", name)
				}

				// A timed-out call keeps running on the server, so let it return
				// before the bubble ends.
				time.Sleep(tt.sleep)
			})
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
