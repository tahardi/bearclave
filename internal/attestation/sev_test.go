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

//go:embed testdata/sev-attestation-b64.txt
var sevAttestationB64 string

func TestSEV_Interfaces(t *testing.T) {
	t.Run("Attester", func(t *testing.T) {
		var _ attestation.Attester = &attestation.SEVAttester{}
	})
	t.Run("Verifier", func(t *testing.T) {
		var _ attestation.Verifier = &attestation.SEVVerifier{}
	})
}

func TestSEVVerifier_Verify(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// given
		want := []byte("Hello, world!")

		// this is the actual timestamp from the testdata report
		unixtime := uint64(1748808574295)
		seconds := int64(unixtime / 1000)
		nanoseconds := int64((unixtime % 1000) * 1_000_000)
		timestamp := time.Unix(seconds, nanoseconds)

		report, err := base64.StdEncoding.DecodeString(sevAttestationB64)
		require.NoError(t, err)

		verifier, err := attestation.NewSEVVerifier()
		require.NoError(t, err)

		// when
		got, err := verifier.Verify(report, attestation.WithTimestamp(timestamp))

		// then
		assert.NoError(t, err)
		assert.Contains(t, string(got), string(want))
	})

	t.Run("error - invalid attestation report", func(t *testing.T) {
		// given
		report := []byte("invalid attestation report")

		verifier, err := attestation.NewSEVVerifier()
		require.NoError(t, err)

		// when
		_, err = verifier.Verify(report)

		// then
		assert.ErrorContains(t, err, "converting sev attestation to proto")
	})

	t.Run("error - expired report", func(t *testing.T) {
		// given
		timestamp := time.Unix(0, 0)

		report, err := base64.StdEncoding.DecodeString(sevAttestationB64)
		require.NoError(t, err)

		verifier, err := attestation.NewSEVVerifier()
		require.NoError(t, err)

		// when
		_, err = verifier.Verify(report, attestation.WithTimestamp(timestamp))

		// then
		assert.ErrorContains(t, err, "certificate has expired or is not yet valid")
	})
}
