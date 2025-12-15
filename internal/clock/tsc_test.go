package clock_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tahardi/bearclave/internal/clock"
)

func TestNewTSCTimer(t *testing.T) {
	vendor := clock.GetVendor()
	t.Run("happy path - "+vendor, func(t *testing.T) {
		timer, err := clock.NewTSCTimer()
		require.NoError(t, err)
		require.NotNil(t, timer)
	})
}

func TestTSCTimer_Roundtrip(t *testing.T) {
	vendor := clock.GetVendor()
	timer, err := clock.NewTSCTimer()
	require.NoError(t, err)

	t.Run("happy path - "+vendor, func(t *testing.T) {
		// given
		wait := 100 * time.Millisecond
		tolerance := 2 * time.Millisecond

		// when
		timer.Start()
		time.Sleep(wait)
		timer.Stop()

		// then
		assert.InDelta(t, wait, timer.ElapsedNanoseconds(), float64(tolerance))
	})
}
