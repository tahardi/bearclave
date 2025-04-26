package bearclave

import (
	"context"

	"github.com/tahardi/bearclave/internal/cvms"
	"github.com/tahardi/bearclave/internal/nitro"
	"github.com/tahardi/bearclave/internal/unsafe"
)

type Communicator interface {
	Close() error
	Send(ctx context.Context, data []byte) (err error)
	Receive(ctx context.Context) (data []byte, err error)
}

type CVMSCommunicator = cvms.Communicator
type NitroCommunicator = nitro.Communicator
type UnsafeEnclaveCommunicator = unsafe.Communicator
type UnsafeNonclaveCommunicator = unsafe.Communicator

func NewCVMSCommunicator() (*CVMSCommunicator, error) {
	return cvms.NewCommunicator()
}

func NewNitroCommunicator() (*NitroCommunicator, error) {
	return nitro.NewCommunicator()
}
func NewUnsafeEnclaveCommunicator(
	sendAddr string,
	receiveAddr string,
) (*UnsafeEnclaveCommunicator, error) {
	return unsafe.NewCommunicator(sendAddr, receiveAddr)
}

func NewUnsafeNonclaveCommunicator(
	sendAddr string,
	receiveAddr string,
) (*UnsafeNonclaveCommunicator, error) {
	return unsafe.NewCommunicator(sendAddr, receiveAddr)
}
