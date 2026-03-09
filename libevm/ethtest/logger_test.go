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
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slog"

	"github.com/ava-labs/libevm/log"
)

type tbRecorder struct {
	testing.TB
	got recorded
}

type recorded struct {
	Logf, Errorf, Fatalf []string
}

// message extracts the log message from the [tbHandler.Handle] call, which
// always passes the original message as a string in `a[1]`.
func message(_ string, a ...any) string {
	s, _ := a[1].(string)
	return s
}

func (r *tbRecorder) Logf(format string, a ...any) {
	r.got.Logf = append(r.got.Logf, message(format, a...))
}

func (r *tbRecorder) Errorf(format string, a ...any) {
	r.got.Errorf = append(r.got.Errorf, message(format, a...))
}

func (r *tbRecorder) Fatalf(format string, a ...any) {
	r.got.Fatalf = append(r.got.Fatalf, message(format, a...))
	panic("fatalf called") // prevent os.Exit(1) after log.Crit
}

func TestTBLogHandler(t *testing.T) {
	doLogging := func(t *testing.T, l log.Logger) {
		t.Helper()
		l.Debug("Austin")
		l.Info("you")
		l.Warn("are")
		l.Error("cool!")
		require.Panics(t, func() { l.Crit("Oh no you lost aura!") }, "Crit()")
	}

	tests := []struct {
		level slog.Level
		want  recorded
	}{
		{
			level: slog.LevelWarn,
			want: recorded{
				Logf: []string{
					"Austin", "you",
				},
				Errorf: []string{
					"are", "cool!",
				},
				Fatalf: []string{
					"Oh no you lost aura!",
				},
			},
		},
		{
			level: slog.LevelError,
			want: recorded{
				Logf: []string{
					"Austin", "you", "are",
				},
				Errorf: []string{
					"cool!",
				},
				Fatalf: []string{
					"Oh no you lost aura!",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.level.String(), func(t *testing.T) {
			rec := &tbRecorder{}
			doLogging(t, log.NewLogger(NewTBLogHandler(rec, tt.level)))

			opts := cmp.Options{
				cmp.AllowUnexported(tbRecorder{}),
				cmpopts.IgnoreInterfaces(struct{ testing.TB }{}),
			}
			if diff := cmp.Diff(tt.want, rec.got, opts); diff != "" {
				t.Errorf("Logf() calls diff (-want +got):\n%s", diff)
			}
		})
	}
}
