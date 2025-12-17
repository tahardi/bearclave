package tee

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/tahardi/bearclave"
)

const AttestPath = "/attest"

type AttestRequest struct {
	Data []byte `json:"data"`
}
type AttestResponse struct {
	Attestation *bearclave.AttestResult `json:"attestation"`
}

func MakeAttestHandler(
	attester bearclave.Attester,
	logger *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := AttestRequest{}
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			WriteError(w, fmt.Errorf("decoding request: %w", err))
			return
		}

		logger.Info(
			"attesting userdata",
			slog.String("userdata", string(req.Data)),
		)
		att, err := attester.Attest(bearclave.WithAttestUserData(req.Data))
		if err != nil {
			WriteError(w, fmt.Errorf("attesting userdata: %w", err))
			return
		}
		WriteResponse(w, AttestResponse{Attestation: att})
	}
}

func WriteError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func WriteResponse(w http.ResponseWriter, out any) {
	data, err := json.Marshal(out)
	if err != nil {
		WriteError(w, fmt.Errorf("marshaling response: %w", err))
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(data)
	if err != nil {
		WriteError(w, fmt.Errorf("writing response: %w", err))
		return
	}
}
