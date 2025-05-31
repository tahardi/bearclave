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
	Send(ctx context.Context, cid int, port int, data []byte) (err error)
	Receive(ctx context.Context) (data []byte, err error)
}

func NewIPC(
	platform setup.Platform,
	cid int,
	port int,
) (IPC, error) {
	switch platform {
	case setup.Nitro:
		return vsockets.NewIPC(cid, port)
	case setup.SEV:
		return sockets.NewIPC(port)
	case setup.TDX:
		return sockets.NewIPC(port)
	case setup.NoTEE:
		return sockets.NewIPC(port)
	default:
		return nil, fmt.Errorf("unsupported platform '%s'", platform)
	}
}
