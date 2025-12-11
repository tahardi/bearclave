package networking_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tahardi/bearclave/internal/networking"
)

func TestParseSocketAddr(t *testing.T) {
	t.Run("happy path - no scheme", func(t *testing.T) {
		// given
		addr := "127.0.0.1:8080"
		wantAddr := "127.0.0.1:8080"

		// when
		gotAddr, err := networking.ParseSocketAddr(addr)

		// then
		require.NoError(t, err)
		assert.Equal(t, wantAddr, gotAddr)
	})

	t.Run("happy path - with scheme", func(t *testing.T) {
		// given
		addr := "http://127.0.0.1:8080"
		wantAddr := "127.0.0.1:8080"

		// when
		gotAddr, err := networking.ParseSocketAddr(addr)

		// then
		require.NoError(t, err)
		assert.Equal(t, wantAddr, gotAddr)
	})
}

func TestParseVSocketAddr(t *testing.T) {
	t.Run("happy path - no scheme", func(t *testing.T) {
		// given
		wantCID := uint32(4)
		wantPort := uint32(8080)
		addr := "4:8080"

		// when
		cid, port, err := networking.ParseVSocketAddr(addr)

		// then
		require.NoError(t, err)
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
		require.NoError(t, err)
		assert.Equal(t, wantCID, cid)
		assert.Equal(t, wantPort, port)
	})

	t.Run("error - invalid format", func(t *testing.T) {
		// given
		addr := "string:8080"

		// when
		_, _, err := networking.ParseVSocketAddr(addr)

		// then
		require.ErrorIs(t, err, networking.ErrVSocketParseAddr)
		assert.ErrorContains(t, err, "expected format")
	})
}
