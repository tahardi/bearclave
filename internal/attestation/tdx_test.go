package attestation_test

import (
	"testing"

	"github.com/tahardi/bearclave/internal/attestation"
)

func TestTDX_Interfaces(t *testing.T) {
	t.Run("Attester", func(t *testing.T) {
		var _ attestation.Attester = &attestation.TDXAttester{}
	})
	t.Run("Verifier", func(t *testing.T) {
		var _ attestation.Verifier = &attestation.TDXVerifier{}
	})
}
