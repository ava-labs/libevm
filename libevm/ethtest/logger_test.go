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
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slog"

	"github.com/ava-labs/libevm/log"
)

type tbRecorder struct {
	testing.TB
	logged, errored, fataled []string
}

func (r *tbRecorder) Logf(format string, a ...any) {
	r.logged = append(r.logged, fmt.Sprintf(format, a...))
}

func (r *tbRecorder) Errorf(format string, a ...any) {
	r.errored = append(r.errored, fmt.Sprintf(format, a...))
}

func (r *tbRecorder) Fatalf(format string, a ...any) {
	r.fataled = append(r.fataled, fmt.Sprintf(format, a...))
	panic("fatalf called") // prevent os.Exit(1) after log.Crit
}

func TestTBLogHandler(t *testing.T) {
	// Each entry in wantLog/wantErr is a list of substrings that must all
	// appear in the corresponding formatted log line.
	tests := []struct {
		name    string
		level   slog.Level
		wantLog [][]string
		wantErr [][]string
	}{
		{
			name:  "warn_level",
			level: slog.LevelWarn,
			wantLog: [][]string{
				{"Hi"},
				{"Austin", "you", "are"},
			},
			wantErr: [][]string{
				{"very"},
				{"cool"},
			},
		},
		{
			name:  "error_level",
			level: slog.LevelError,
			wantLog: [][]string{
				{"Hi"},
				{"Austin", "you", "are"},
				{"very"},
			},
			wantErr: [][]string{
				{"cool"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := &tbRecorder{}
			l := log.NewLogger(NewTBLogHandler(got, tt.level))

			l.Debug("Hi")
			l.Info("Austin", "you", "are")
			l.Warn("very")
			l.Error("cool")

			require.Len(t, got.logged, len(tt.wantLog), "Logf() calls")
			for i, wants := range tt.wantLog {
				for _, want := range wants {
					require.Contains(t, got.logged[i], want, "Logf()[%d]", i)
				}
			}

			require.Len(t, got.errored, len(tt.wantErr), "Errorf() calls")
			for i, wants := range tt.wantErr {
				for _, want := range wants {
					require.Contains(t, got.errored[i], want, "Errorf()[%d]", i)
				}
			}
		})
	}
}

func TestTBLogHandler_Crit(t *testing.T) {
	got := &tbRecorder{}
	l := log.NewLogger(NewTBLogHandler(got, slog.LevelWarn))

	require.Panics(t, func() { l.Crit("Explosion") }, "Crit()")
	require.Len(t, got.fataled, 1, "Fatalf() calls")
	require.Contains(t, got.fataled[0], "Explosion", "Fatalf()[0]")
}
