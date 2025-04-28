package bearclave

import (
	"github.com/tahardi/bearclave/internal/cvms"
	"github.com/tahardi/bearclave/internal/nitro"
	"github.com/tahardi/bearclave/internal/unsafe"
)

type CVMSAttester = cvms.Attester
type NitroAttester = nitro.Attester
type UnsafeAttester = unsafe.Attester

type Attester interface {
	Attest(userdata []byte) (attestation []byte, err error)
}

func NewCVMSAttester() (*CVMSAttester, error) {
	return cvms.NewAttester()
}

func NewNitroAttester() (*NitroAttester, error) {
	return nitro.NewAttester()
}

func NewUnsafeAttester() (*UnsafeAttester, error) {
	return unsafe.NewAttester()
}
