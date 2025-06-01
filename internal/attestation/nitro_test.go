package attestation_test

import (
	_ "embed"
	"encoding/base64"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tahardi/bearclave/internal/attestation"
)

//go:embed testdata/nitro-attestation-b64.txt
var nitroAttestationB64 string

func TestNitro_Interfaces(t *testing.T) {
	t.Run("Attester", func(t *testing.T) {
		var _ attestation.Attester = &attestation.NitroAttester{}
	})
	t.Run("Verifier", func(t *testing.T) {
		var _ attestation.Verifier = &attestation.NitroVerifier{}
	})
}

func TestNitroVerifier_Verify(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// given
		want := []byte("Hello, world!")

		// this is the actual timestamp from the testdata report
		unixtime := uint64(1748808574295)
		seconds := int64(unixtime / 1000)
		nanoseconds := int64((unixtime % 1000) * 1_000_000)
		timestamp := time.Unix(seconds, nanoseconds)

		report, err := base64.StdEncoding.DecodeString(nitroAttestationB64)
		require.NoError(t, err)

		verifier, err := attestation.NewNitroVerifier()
		require.NoError(t, err)

		// when
		got, err := verifier.Verify(report, attestation.WithTimestamp(timestamp))

		// then
		assert.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("error - invalid attestation report", func(t *testing.T) {
		// given
		report := []byte("invalid attestation report")

		verifier, err := attestation.NewNitroVerifier()
		require.NoError(t, err)

		// when
		_, err = verifier.Verify(report)

		// then
		assert.ErrorContains(t, err, "verifying attestation")
	})

	t.Run("error - expired report", func(t *testing.T) {
		// given
		// this is 5 years after the testdata report
		unixtime := uint64(1906596574295)
		seconds := int64(unixtime / 1000)
		nanoseconds := int64((unixtime % 1000) * 1_000_000)
		timestamp := time.Unix(seconds, nanoseconds)

		report, err := base64.StdEncoding.DecodeString(nitroAttestationB64)
		require.NoError(t, err)

		verifier, err := attestation.NewNitroVerifier()
		require.NoError(t, err)

		// when
		_, err = verifier.Verify(report, attestation.WithTimestamp(timestamp))

		// then
		assert.ErrorContains(t, err, "certificate has expired")
	})
}
