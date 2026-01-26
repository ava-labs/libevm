package ethtest

import (
	"context"
	"testing"

	"github.com/ava-labs/libevm/log"
	"golang.org/x/exp/slog"
)

// NewTBLogHandler constructs a [slog.Handler] that propagates logs to [testing.TB].
// Logs at [log.LevelWarn] or above go to [testing.TB.Errorf], except
// [log.LevelCrit] which goes to [testing.TB.Fatalf]. All other logs go to
// [testing.TB.Logf].
//
//nolint:thelper // The outputs include the logging site while the TB site is most useful if here
func NewTBLogHandler(tb testing.TB) slog.Handler {
	return &tbHandler{tb: tb}
}

type tbHandler struct {
	tb testing.TB
}

func (h *tbHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= log.LevelWarn
}

func (h *tbHandler) Handle(_ context.Context, r slog.Record) error {
	to := h.tb.Logf
	switch {
	case r.Level >= log.LevelCrit:
		to = h.tb.Fatalf
	case r.Level >= log.LevelWarn:
		to = h.tb.Errorf
	}
	to("[%s] %s", log.LevelAlignedString(r.Level), r.Message)
	return nil
}

// WithAttrs and WithGroup return the receiver unchanged. Attribute/group
// context is not needed for test failure detection.
func (h *tbHandler) WithAttrs([]slog.Attr) slog.Handler { return h }
func (h *tbHandler) WithGroup(string) slog.Handler      { return h }
