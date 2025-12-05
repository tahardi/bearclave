package bearclave

import (
	"fmt"
	"net/http"

	"github.com/tahardi/bearclave/internal/networking"
)

type Server = networking.Server

func NewServer(
	platform Platform,
	port int,
	mux *http.ServeMux,
) (*Server, error) {
	switch platform {
	case Nitro:
		return networking.NewVSocketServer(port, mux)
	case SEV:
		return networking.NewSocketServer(port, mux)
	case TDX:
		return networking.NewSocketServer(port, mux)
	case NoTEE:
		return networking.NewSocketServer(port, mux)
	default:
		return nil, fmt.Errorf("unsupported platform '%s'", platform)
	}
}
