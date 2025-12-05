package bearclave

import (
	"fmt"

	"github.com/tahardi/bearclave/internal/ipc"
)

type IPC = ipc.IPC

func NewIPC(platform Platform, endpoint string) (IPC, error) {
	switch platform {
	case Nitro:
		return ipc.NewVSocketIPC(endpoint)
	case SEV:
		return ipc.NewSocketIPC(endpoint)
	case TDX:
		return ipc.NewSocketIPC(endpoint)
	case NoTEE:
		return ipc.NewSocketIPC(endpoint)
	default:
		return nil, fmt.Errorf("unsupported platform '%s'", platform)
	}
}
