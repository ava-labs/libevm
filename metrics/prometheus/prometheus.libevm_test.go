// (c) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package prometheus

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ava-labs/libevm/metrics"
)

func TestGatherer(t *testing.T) {
	registry := metrics.NewRegistry()

	counter := metrics.NewCounter()
	counter.Inc(12345)

	err := registry.Register("test/counter", counter)
	require.NoError(t, err)

	gauge := metrics.NewGauge()
	gauge.Update(23456)

	err = registry.Register("test/gauge", gauge)
	require.NoError(t, err)

	gaugeFloat64 := metrics.NewGaugeFloat64()
	gaugeFloat64.Update(34567.89)

	err = registry.Register("test/gauge_float64", gaugeFloat64)
	require.NoError(t, err)

	sample := metrics.NewUniformSample(1028)
	histogram := metrics.NewHistogram(sample)

	err = registry.Register("test/histogram", histogram)
	require.NoError(t, err)

	meter := metrics.NewMeter()
	defer meter.Stop()
	meter.Mark(9999999)

	err = registry.Register("test/meter", meter)
	require.NoError(t, err)

	timer := metrics.NewTimer()
	defer timer.Stop()
	timer.Update(20 * time.Millisecond)
	timer.Update(21 * time.Millisecond)
	timer.Update(22 * time.Millisecond)
	timer.Update(120 * time.Millisecond)
	timer.Update(23 * time.Millisecond)
	timer.Update(24 * time.Millisecond)

	err = registry.Register("test/timer", timer)
	require.NoError(t, err)

	resettingTimer := metrics.NewResettingTimer()
	resettingTimer.Update(10 * time.Millisecond)
	resettingTimer.Update(11 * time.Millisecond)
	resettingTimer.Update(12 * time.Millisecond)
	resettingTimer.Update(120 * time.Millisecond)
	resettingTimer.Update(13 * time.Millisecond)
	resettingTimer.Update(14 * time.Millisecond)

	err = registry.Register("test/resetting_timer", resettingTimer)
	require.NoError(t, err)

	err = registry.Register("test/resetting_timer_snapshot", resettingTimer.Snapshot())
	require.NoError(t, err)

	emptyResettingTimer := metrics.NewResettingTimer()

	err = registry.Register("test/empty_resetting_timer", emptyResettingTimer)
	require.NoError(t, err)

	err = registry.Register("test/empty_resetting_timer_snapshot", emptyResettingTimer.Snapshot())
	require.NoError(t, err)

	g := NewGatherer(registry)

	_, err = g.Gather()
	require.NoError(t, err)
}
