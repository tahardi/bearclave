package bearclave

import (
	"context"

	"github.com/tahardi/bearclave/internal/nitro"
	"github.com/tahardi/bearclave/internal/sev"
	"github.com/tahardi/bearclave/internal/tdx"
	"github.com/tahardi/bearclave/internal/unsafe"
)

const (
	NitroNonclaveCID = nitro.NonclaveCID
	NitroEnclaveCID  = nitro.EnclaveCID
)

type NitroCommunicator = nitro.Communicator
type SEVCommunicator = sev.Communicator
type TDXCommunicator = tdx.Communicator
type UnsafeCommunicator = unsafe.Communicator

type Communicator interface {
	Close() error
	Send(ctx context.Context, data []byte) (err error)
	Receive(ctx context.Context) (data []byte, err error)
}

func NewNitroCommunicator(
	sendContextID int,
	sendPort int,
	receivePort int,
) (*NitroCommunicator, error) {
	return nitro.NewCommunicator(sendContextID, sendPort, receivePort)
}

func NewSEVCommunicator(
	sendAddr string,
	receiveAddr string,
) (*SEVCommunicator, error) {
	return sev.NewCommunicator(sendAddr, receiveAddr)
}

func NewTDXCommunicator(
	sendContextID int,
	sendPort int,
	receivePort int,
) (*TDXCommunicator, error) {
	return tdx.NewCommunicator(sendContextID, sendPort, receivePort)
}

func NewUnsafeCommunicator(
	sendAddr string,
	receiveAddr string,
) (*UnsafeCommunicator, error) {
	return unsafe.NewCommunicator(sendAddr, receiveAddr)
}
