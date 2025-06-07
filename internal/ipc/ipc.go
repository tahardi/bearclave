package ipc

import (
	"context"
	"fmt"

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
		return NewVSocketIPC(endpoint)
	case setup.SEV:
		return NewSocketIPC(endpoint)
	case setup.TDX:
		return NewSocketIPC(endpoint)
	case setup.NoTEE:
		return NewSocketIPC(endpoint)
	default:
		return nil, fmt.Errorf("unsupported platform '%s'", platform)
	}
}
