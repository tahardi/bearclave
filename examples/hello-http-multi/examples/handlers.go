package examples

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/tahardi/bearclave/pkg/attestation"
	"github.com/tahardi/bearclave/pkg/networking"
)

const HelloMultipleServersPath = "/hello-multiple-servers"

type MultipleServersResponse struct {
	Report []byte `json:"report"`
}

func MakeHelloMultipleServersHandler(
	logger *slog.Logger,
	serviceName string,
	attester attestation.Attester,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userdata := fmt.Sprintf("Hello from %s!", serviceName)
		logger.Info(
			"attesting userdata",
			slog.String("userdata", userdata),
		)

		report, err := attester.Attest([]byte(userdata))
		if err != nil {
			networking.WriteError(w, fmt.Errorf("attesting userdata: %w", err))
			return
		}
		networking.WriteResponse(w, MultipleServersResponse{Report: report})
	}
}
