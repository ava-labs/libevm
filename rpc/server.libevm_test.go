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
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/require"
)

func TestServerSetCallTimeout(t *testing.T) {
	tests := []struct {
		name    string
		timeout time.Duration
		wantErr error
	}{
		{
			name: "no_timeout",
		},
		{
			name:    "timeout",
			timeout: time.Second,
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
				client := DialInProc(srv)
				defer client.Close()

				const sleep = time.Minute
				err := client.CallContext(t.Context(), nil, "test_sleep", sleep)
				require.Equal(t, tt.wantErr, err, "CallContext(test_sleep)")

				// A timed-out call doesn't close the connection.
				require.NoError(t, client.CallContext(t.Context(), nil, "test_noArgsRets"), "CallContext(test_noArgsRets)")

				// A timed-out call keeps running on the server, so let it return
				// before the bubble ends.
				time.Sleep(sleep)
			})
		})
	}
}
