package unsafe_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tahardi/bearclave/internal/unsafe"
)

func TestDetector_Detect(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// given
		detector, err := unsafe.NewDetector()
		require.NoError(t, err)

		// when
		platform, ok := detector.Detect()

		// then
		assert.True(t, ok)
		assert.Equal(t, unsafe.Platform, platform)
	})
}
