package bearclave

import (
	"context"

	"github.com/tahardi/bearclave/internal/sock"
	"github.com/tahardi/bearclave/internal/vsock"
)

const (
	// TODO: Update Nonclave to Gateway or whatever term you end up using
	// NonclaveCID In AWS Nitro the "nonclave" program runs on the host (i.e.,
	// the parent EC2 instance), which, according to documentation, is always 3.
	NitroNonclaveCID = 3

	// NitroEnclaveCID In AWS Nitro the "enclave" program runs on the guest (i.e., the
	// VM), which can be any value between 4 and 1023. We use 4 here because it's
	// the default value for the `cid` argument to `nitro-cli run-enclave`.
	NitroEnclaveCID = 4
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
	sendAddr string,
	receiveAddr string,
) (*SEVTransporter, error) {
	return sock.NewTransporter(sendAddr, receiveAddr)
}

// NewTDXTransporter TDX Enclaves have access to standard networking interfaces, hence the use of
// Unix Sockets. Note that communication only happens between enclave programs running within the
// confidential VM. Communication between the host programs and enclave programs is not possible.
func NewTDXTransporter(
	sendAddr string,
	receiveAddr string,
) (*TDXTransporter, error) {
	return sock.NewTransporter(sendAddr, receiveAddr)
}

func NewUnsafeTransporter(
	sendAddr string,
	receiveAddr string,
) (*UnsafeTransporter, error) {
	return sock.NewTransporter(sendAddr, receiveAddr)
}
