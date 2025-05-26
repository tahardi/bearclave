package networking

import (
	"log/slog"
	"net/http"

	"github.com/tahardi/bearclave/internal/networking"
	"github.com/tahardi/bearclave/pkg/attestation"
)

const AttestUserDataPath = networking.AttestUserDataPath

type AttestUserDataRequest = networking.AttestUserDataRequest
type AttestUserDataResponse = networking.AttestUserDataResponse

func MakeAttestUserDataHandler(
	attester attestation.Attester,
	logger *slog.Logger,
) http.HandlerFunc {
	return networking.MakeAttestUserDataHandler(attester, logger)
}
