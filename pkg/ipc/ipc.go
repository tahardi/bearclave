package ipc

import (
	"github.com/tahardi/bearclave/internal/ipc"
	"github.com/tahardi/bearclave/pkg/setup"
)

type IPC = ipc.IPC

func NewIPC(
	platform setup.Platform,
	sendCID int,
	sendPort int,
	receivePort int,
) (IPC, error) {
	return ipc.NewIPC(platform, sendCID, sendPort, receivePort)
}
