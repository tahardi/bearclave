package examples

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/tahardi/bearclave/internal/networking"
)

const HelloMultipleServersPath = "/hello-multiple-servers"

type MultipleServersResponse struct {
	Message string `json:"message"`
}

func MakeHelloMultipleServersHandler(
	logger *slog.Logger,
	serviceName string,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("handling hello-multiple-servers request")
		message := fmt.Sprintf("Hello from %s!", serviceName)
		networking.WriteResponse(w, MultipleServersResponse{Message: message})
	}
}
