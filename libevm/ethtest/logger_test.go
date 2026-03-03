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
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slog"

	"github.com/ava-labs/libevm/log"
)

type tbRecorder struct {
	testing.TB
	logged, errored, fataled []string
}

// message extracts the log message from the handler's format args.
func message(_ string, a ...any) string {
	s, _ := a[1].(string)
	return s
}

func (r *tbRecorder) Logf(format string, a ...any) {
	r.logged = append(r.logged, message(format, a...))
}

func (r *tbRecorder) Errorf(format string, a ...any) {
	r.errored = append(r.errored, message(format, a...))
}

func (r *tbRecorder) Fatalf(format string, a ...any) {
	r.fataled = append(r.fataled, message(format, a...))
	panic("fatalf called") // prevent os.Exit(1) after log.Crit
}

func TestTBLogHandler(t *testing.T) {
	doLogging := func(t *testing.T, l log.Logger) {
		t.Helper()
		l.Debug("Hi")
		l.Info("Austin", "you", "are")
		l.Warn("very")
		l.Error("cool!")
		require.Panics(t, func() { l.Crit("oh no you exploded") }, "Crit()")
	}

	tests := []struct {
		name      string
		level     slog.Level
		wantLog   []string
		wantErr   []string
		wantFatal []string
	}{
		{
			name:  "warn_level",
			level: slog.LevelWarn,
			wantLog: []string{
				"Hi", "Austin",
			},
			wantErr: []string{
				"very", "cool!",
			},
			wantFatal: []string{
				"oh no you exploded",
			},
		},
		{
			name:  "error_level",
			level: slog.LevelError,
			wantLog: []string{
				"Hi", "Austin", "very",
			},
			wantErr: []string{
				"cool!",
			},
			wantFatal: []string{
				"oh no you exploded",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := &tbRecorder{}
			l := log.NewLogger(NewTBLogHandler(got, tt.level))

			doLogging(t, l)

			if diff := cmp.Diff(tt.wantLog, got.logged); diff != "" {
				t.Errorf("Logf() calls diff (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.wantErr, got.errored); diff != "" {
				t.Errorf("Errorf() calls diff (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.wantFatal, got.fataled); diff != "" {
				t.Errorf("Fatalf() calls diff (-want +got):\n%s", diff)
			}
		})
	}
}
