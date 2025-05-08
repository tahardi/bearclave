package bearclave

import (
	"context"

	"github.com/tahardi/bearclave/internal/sock"
	"github.com/tahardi/bearclave/internal/vsock"
)

type NitroTransporter = vsock.Transporter
type SEVTransporter = sock.Transporter
type TDXTransporter = sock.Transporter
type UnsafeTransporter = sock.Transporter

// Transporter is used to send and receive data between locally between enclave programs.
// For example, the enclave program running on the host can send data to the enclave program
// running on the guest, and vice versa.
type Transporter interface {
	Close() error
	Send(ctx context.Context, data []byte) (err error)
	Receive(ctx context.Context) (data []byte, err error)
}

// NewNitroTransporter Nitro Enclaves do not have access to any standard networking interfaces.
// Instead, their only form of communication is via Linux Virtual Sockets (VSOCK), which allows
// for communication between programs running on a VM and those running on the host.
func NewNitroTransporter(
	sendContextID int,
	sendPort int,
	receivePort int,
) (*NitroTransporter, error) {
	return vsock.NewTransporter(sendContextID, sendPort, receivePort)
}

// NewSEVTransporter SEV Enclaves have access to standard networking interfaces, hence the use of
// Unix Sockets. Note that communication only happens between enclave programs running within the
// confidential VM. Communication between the host programs and enclave programs is not possible.
func NewSEVTransporter(
	sendPort int,
	receivePort int,
) (*SEVTransporter, error) {
	return sock.NewTransporter(sendPort, receivePort)
}

// NewTDXTransporter TDX Enclaves have access to standard networking interfaces, hence the use of
// Unix Sockets. Note that communication only happens between enclave programs running within the
// confidential VM. Communication between the host programs and enclave programs is not possible.
func NewTDXTransporter(
	sendPort int,
	receivePort int,
) (*TDXTransporter, error) {
	return sock.NewTransporter(sendPort, receivePort)
}

func NewUnsafeTransporter(
	sendPort int,
	receivePort int,
) (*UnsafeTransporter, error) {
	return sock.NewTransporter(sendPort, receivePort)
}
