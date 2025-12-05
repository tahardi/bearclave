package tee

import (
	"fmt"
	"net/http"

	"github.com/tahardi/bearclave"
	"github.com/tahardi/bearclave/internal/networking"
)

type Server = networking.Server

func NewServer(
	platform bearclave.Platform,
	port int,
	mux *http.ServeMux,
) (*Server, error) {
	switch platform {
	case bearclave.Nitro:
		return networking.NewVSocketServer(port, mux)
	case bearclave.SEV:
		return networking.NewSocketServer(port, mux)
	case bearclave.TDX:
		return networking.NewSocketServer(port, mux)
	case bearclave.NoTEE:
		return networking.NewSocketServer(port, mux)
	default:
		return nil, fmt.Errorf("unsupported platform '%s'", platform)
	}
}
