package bearclave

import (
	"github.com/tahardi/bearclave/internal/cvms"
	"github.com/tahardi/bearclave/internal/nitro"
	"github.com/tahardi/bearclave/internal/sev"
	"github.com/tahardi/bearclave/internal/tdx"
	"github.com/tahardi/bearclave/internal/unsafe"
)

type CVMSVerifier = cvms.Verifier
type NitroVerifier = nitro.Verifier
type SEVVerifier = sev.Verifier
type TDXVerifier = tdx.Verifier
type UnsafeVerifier = unsafe.Verifier

type Verifier interface {
	Verify(attestation []byte) (userdata []byte, err error)
}

func NewNitroVerifier() (*NitroVerifier, error) {
	return nitro.NewVerifier()
}
func NewSEVVerifier() (*SEVVerifier, error) {
	return sev.NewVerifier()
}
func NewTDXVerifier() (*TDXVerifier, error) {
	return tdx.NewVerifier()
}
func NewUnsafeVerifier() (*UnsafeVerifier, error) {
	return unsafe.NewVerifier()
}
