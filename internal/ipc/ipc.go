package ipc

import (
	"context"
	"fmt"

	"github.com/tahardi/bearclave/internal/ipc/sockets"
	"github.com/tahardi/bearclave/internal/ipc/vsockets"
	"github.com/tahardi/bearclave/internal/setup"
)

type IPC interface {
	Close() error
	Send(ctx context.Context, endpoint string, data []byte) (err error)
	Receive(ctx context.Context) (data []byte, err error)
}

func NewIPC(platform setup.Platform, endpoint string) (IPC, error) {
	switch platform {
	case setup.Nitro:
		return vsockets.NewIPC(endpoint)
	case setup.SEV:
		return sockets.NewIPC(endpoint)
	case setup.TDX:
		return sockets.NewIPC(endpoint)
	case setup.NoTEE:
		return sockets.NewIPC(endpoint)
	default:
		return nil, fmt.Errorf("unsupported platform '%s'", platform)
	}
}
