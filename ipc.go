package bearclave

import (
	"github.com/tahardi/bearclave/internal/ipc"
)

type IPC = ipc.IPC

func NewIPC(platform Platform, endpoint string) (IPC, error) {
	return ipc.NewIPC(platform, endpoint)
}
