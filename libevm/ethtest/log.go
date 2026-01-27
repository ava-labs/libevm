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

package ethtest

import (
	"context"
	"runtime"
	"testing"

	"golang.org/x/exp/slog"

	"github.com/ava-labs/libevm/log"
)

// NewTBLogHandler constructs a [slog.Handler] that propagates logs to [testing.TB].
// Logs at [log.LevelWarn] or above go to [testing.TB.Errorf], except
// [log.LevelCrit] which goes to [testing.TB.Fatalf]. All other logs go to
// [testing.TB.Logf]. The level parameter controls which logs are enabled.
//
//nolint:thelper // The outputs include the logging site while the TB site is most useful if here
func NewTBLogHandler(tb testing.TB, level slog.Level) slog.Handler {
	return &tbHandler{
		tb:    tb,
		level: level,
	}
}

type tbHandler struct {
	tb    testing.TB
	level slog.Level
	attrs []slog.Attr
}

func (h *tbHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= min(h.level, slog.LevelWarn)
}

func (h *tbHandler) Handle(_ context.Context, rec slog.Record) error {
	to := h.tb.Logf
	switch {
	case r.Level >= log.LevelCrit:
		to = h.tb.Fatalf
	case r.Level >= log.LevelWarn:
		to = h.tb.Errorf
	}

	_, file, line, _ := runtime.Caller(3)

	fields := make(map[string]any, len(h.attrs)+r.NumAttrs())
	for _, attr := range h.attrs {
		fields[attr.Key] = attr.Value.Any()
	}
	r.Attrs(func(attr slog.Attr) bool {
		fields[attr.Key] = attr.Value.Any()
		return true
	})

	to("[%s] %s %v - %s:%d", log.LevelAlignedString(r.Level), r.Message, fields, file, line)
	return nil
}

func (h *tbHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &tbHandler{
		tb:    h.tb,
		level: h.level,
		attrs: append(h.attrs[:len(h.attrs):len(h.attrs)], attrs...),
	}
}

func (h *tbHandler) WithGroup(string) slog.Handler { return h }
