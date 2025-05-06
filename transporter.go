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

	// EnclaveCID In AWS Nitro the "enclave" program runs on the guest (i.e., the
	// VM), which can be any value between 4 and 1023. We use 4 here because it's
	// the default value for the `cid` argument to `nitro-cli run-enclave`.
	NitroEnclaveCID = 4
)

type NitroTransporter = vsock.Transporter
type SEVTransporter = sock.Transporter
type TDXTransporter = sock.Transporter
type UnsafeTransporter = sock.Transporter

// TODO: Switch to Transport or Transporter
type Transporter interface {
	Close() error
	Send(ctx context.Context, data []byte) (err error)
	Receive(ctx context.Context) (data []byte, err error)
}

func NewNitroTransporter(
	sendContextID int,
	sendPort int,
	receivePort int,
) (*NitroTransporter, error) {
	return vsock.NewTransporter(sendContextID, sendPort, receivePort)
}

func NewSEVTransporter(
	sendAddr string,
	receiveAddr string,
) (*SEVTransporter, error) {
	return sock.NewTransporter(sendAddr, receiveAddr)
}

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
