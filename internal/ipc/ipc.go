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
	Send(ctx context.Context, data []byte) (err error)
	Receive(ctx context.Context) (data []byte, err error)
}

func NewIPC(
	platform setup.Platform,
	sendCID int,
	sendPort int,
	receivePort int,
) (IPC, error) {
	switch platform {
	case setup.Nitro:
		return vsockets.NewIPC(sendCID, sendPort, receivePort)
	case setup.SEV:
		return sockets.NewIPC(sendPort, receivePort)
	case setup.TDX:
		return sockets.NewIPC(sendPort, receivePort)
	case setup.NoTEE:
		return sockets.NewIPC(sendPort, receivePort)
	default:
		return nil, fmt.Errorf("unsupported platform '%s'", platform)
	}
}
