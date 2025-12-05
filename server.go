package bearclave

import (
	"net/http"

	"github.com/tahardi/bearclave/internal/networking"
)

type Server = networking.Server

func NewServer(
	platform Platform,
	port int,
	mux *http.ServeMux,
) (*Server, error) {
	return networking.NewServer(platform, port, mux)
}
