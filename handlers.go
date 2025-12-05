package bearclave

import (
	"log/slog"
	"net/http"

	"github.com/tahardi/bearclave/internal/networking"
)

const AttestUserDataPath = networking.AttestUserDataPath

type AttestUserDataRequest = networking.AttestUserDataRequest
type AttestUserDataResponse = networking.AttestUserDataResponse

func MakeAttestUserDataHandler(
	attester Attester,
	logger *slog.Logger,
) http.HandlerFunc {
	return networking.MakeAttestUserDataHandler(attester, logger)
}

func WriteError(w http.ResponseWriter, err error) {
	networking.WriteError(w, err)
}

func WriteResponse(w http.ResponseWriter, out any) {
	networking.WriteResponse(w, out)
}
