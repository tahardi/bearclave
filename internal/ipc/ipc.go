package ipc

import (
	"context"
)

type IPC interface {
	Close() error
	Send(ctx context.Context, endpoint string, data []byte) (err error)
	Receive(ctx context.Context) (data []byte, err error)
}
