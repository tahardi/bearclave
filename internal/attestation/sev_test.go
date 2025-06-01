package attestation_test

import (
	"testing"

	"github.com/tahardi/bearclave/internal/attestation"
)

func TestSEV_Interfaces(t *testing.T) {
	t.Run("Attester", func(t *testing.T) {
		var _ attestation.Attester = &attestation.SEVAttester{}
	})
	t.Run("Verifier", func(t *testing.T) {
		var _ attestation.Verifier = &attestation.SEVVerifier{}
	})
}
