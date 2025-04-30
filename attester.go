package bearclave

import (
	"github.com/tahardi/bearclave/internal/nitro"
	"github.com/tahardi/bearclave/internal/sev"
	"github.com/tahardi/bearclave/internal/tdx"
	"github.com/tahardi/bearclave/internal/unsafe"
)

type NitroAttester = nitro.Attester
type SEVAttester = sev.Attester
type TDXAttester = tdx.Attester
type UnsafeAttester = unsafe.Attester

type Attester interface {
	Attest(userdata []byte) (attestation []byte, err error)
}

func NewNitroAttester() (*NitroAttester, error) {
	return nitro.NewAttester()
}
func NewSEVAttester() (*SEVAttester, error) {
	return sev.NewAttester()
}
func NewTDXAttester() (*TDXAttester, error) {
	return tdx.NewAttester()
}
func NewUnsafeAttester() (*UnsafeAttester, error) {
	return unsafe.NewAttester()
}
