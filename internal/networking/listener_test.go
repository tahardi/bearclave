package networking_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tahardi/bearclave/internal/networking"
)

func TestParseVSocketAddr(t *testing.T) {
	t.Run("happy path - no scheme", func(t *testing.T) {
		// given
		wantCID := uint32(4)
		wantPort := uint32(8080)
		addr := "4:8080"

		// when
		cid, port, err := networking.ParseVSocketAddr(addr)

		// then
		assert.NoError(t, err)
		assert.Equal(t, wantCID, cid)
		assert.Equal(t, wantPort, port)
	})

	t.Run("happy path - with scheme", func(t *testing.T) {
		// given
		wantCID := uint32(4)
		wantPort := uint32(8080)
		addr := "http://4:8080"

		// when
		cid, port, err := networking.ParseVSocketAddr(addr)

		// then
		assert.NoError(t, err)
		assert.Equal(t, wantCID, cid)
		assert.Equal(t, wantPort, port)
	})

	t.Run("error - invalid format", func(t *testing.T) {
		// given
		addr := "string:8080"

		// when
		_, _, err := networking.ParseVSocketAddr(addr)

		// then
		assert.ErrorContains(t, err, "invalid format")
	})
}