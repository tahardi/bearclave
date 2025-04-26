package bearclave

import (
	"github.com/tahardi/bearclave/internal/cvms"
	"github.com/tahardi/bearclave/internal/nitro"
	"github.com/tahardi/bearclave/internal/unsafe"
)

type Verifier interface {
	Verify(attestation []byte) (userdata []byte, err error)
}

type CVMSVerifier = cvms.Verifier
type NitroVerifier = nitro.Verifier
type UnsafeVerifier = unsafe.Verifier

func NewCVMSVerifier() (*CVMSVerifier, error) {
	return cvms.NewVerifier()
}

func NewNitroVerifier() (*NitroVerifier, error) {
	return nitro.NewVerifier()
}

func NewUnsafeVerifier() (*UnsafeVerifier, error) {
	return unsafe.NewVerifier()
}
