// Copyright 2025 the libevm authors.
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

// Package jsoncmp provides a [cmp.Option] for comparing JSON buffers.
package jsoncmp

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// AsMapToAny returns a [cmp.Transformer] that unmarshals all `[]byte` slices
// into a `map[string]any`, treating numerical values as [json.Number] to avoid
// loss of precision. Empty slices are transformed into nil maps. Any error
// returned by the [json] package will be reported via [testing.TB.Errorf].
func AsMapToAny(tb testing.TB) cmp.Option {
	tb.Helper()

	return cmp.Transformer("unmarshal_json", func(buf []byte) map[string]any {
		tb.Helper()

		if len(buf) == 0 {
			return nil
		}

		dec := json.NewDecoder(bytes.NewReader(buf))
		dec.UseNumber()
		out := make(map[string]any)
		if err := dec.Decode(&out); err != nil {
			tb.Errorf("json.Unmarshal(..., %T) error %v", out, err)
			return nil
		}
		return out
	})
}
