package networking

import (
	"net/http"

	"github.com/tahardi/bearclave/internal/networking"
	"github.com/tahardi/bearclave/pkg/setup"
)

type Server = networking.Server

func NewServer(
	platform setup.Platform,
	port int,
	mux *http.ServeMux,
) (*Server, error) {
	return networking.NewServer(platform, port, mux)
}
