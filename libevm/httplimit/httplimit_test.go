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

package httplimit

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLimitBody(t *testing.T) {
	t.Run("declared Content-Length over the limit", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("x"))
		req.ContentLength = MaxBodySize + 1
		rec := httptest.NewRecorder()

		assert.False(t, LimitBody(rec, req), "LimitBody() with an oversized Content-Length")
		assert.Equal(t, http.StatusRequestEntityTooLarge, rec.Code, "status code")
	})

	t.Run("within the limit", func(t *testing.T) {
		const body = "hello"
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
		rec := httptest.NewRecorder()

		require.True(t, LimitBody(rec, req), "LimitBody() with a small body")
		got, err := io.ReadAll(req.Body)
		require.NoError(t, err, "io.ReadAll(r.Body)")
		assert.Equal(t, body, string(got), "body is readable and unchanged")
	})

	t.Run("undeclared body over the limit", func(t *testing.T) {
		// A body that lies about (or omits) its length MUST surface as a read
		// error rather than being read into memory in full.
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(strings.Repeat("x", MaxBodySize+100)))
		req.ContentLength = -1 // i.e. unknown
		rec := httptest.NewRecorder()

		require.True(t, LimitBody(rec, req), "LimitBody() with an unknown Content-Length")
		n, err := io.Copy(io.Discard, req.Body)
		require.Error(t, err, "reading a body that exceeds the limit")
		assert.LessOrEqual(t, n, int64(MaxBodySize), "bytes read before the error")
	})
}
