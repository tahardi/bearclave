package bearclave

import (
	"context"

	"github.com/tahardi/bearclave/internal/cvms"
	"github.com/tahardi/bearclave/internal/nitro"
	"github.com/tahardi/bearclave/internal/unsafe"
)

const (
	NitroNonclaveCID = nitro.NonclaveCID
	NitroEnclaveCID  = nitro.EnclaveCID
)

type CVMSCommunicator = cvms.Communicator
type NitroCommunicator = nitro.Communicator
type UnsafeCommunicator = unsafe.Communicator

type Communicator interface {
	Close() error
	Send(ctx context.Context, data []byte) (err error)
	Receive(ctx context.Context) (data []byte, err error)
}

func NewCVMSCommunicator() (*CVMSCommunicator, error) {
	return cvms.NewCommunicator()
}

func NewNitroCommunicator(
	sendContextID int,
	sendPort int,
	receivePort int,
) (*NitroCommunicator, error) {
	return nitro.NewCommunicator(sendContextID, sendPort, receivePort)
}

func NewUnsafeCommunicator(
	sendAddr string,
	receiveAddr string,
) (*UnsafeCommunicator, error) {
	return unsafe.NewCommunicator(sendAddr, receiveAddr)
}
