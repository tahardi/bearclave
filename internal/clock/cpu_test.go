package clock_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tahardi/bearclave/internal/clock"
)

func TestGetTSCFrequency(t *testing.T) {
	vendor := clock.GetVendor()
	if vendor != clock.Intel {
		t.Skip("skipping test for non-Intel CPU")
	}

	t.Run("happy path - "+vendor, func(t *testing.T) {
		frequency, err := clock.GetTSCFrequency()
		require.NoError(t, err)
		require.Positive(t, frequency)
	})
}
