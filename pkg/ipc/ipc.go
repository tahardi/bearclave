package ipc

import (
	"github.com/tahardi/bearclave/internal/ipc"
	"github.com/tahardi/bearclave/pkg/setup"
)

type IPC = ipc.IPC

func NewIPC(platform setup.Platform, endpoint string) (IPC, error) {
	return ipc.NewIPC(platform, endpoint)
}
