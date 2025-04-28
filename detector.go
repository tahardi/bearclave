package bearclave

import (
	"github.com/tahardi/bearclave/internal/cvms"
	"github.com/tahardi/bearclave/internal/nitro"
	"github.com/tahardi/bearclave/internal/unsafe"
)

const (
	CVMSPlatform   = cvms.Platform
	NitroPlatform  = nitro.Platform
	UnsafePlatform = unsafe.Platform
)

type CVMSDetector = cvms.Detector
type NitroDetector = nitro.Detector
type UnsafeDetector = unsafe.Detector

type Detector interface {
	Detect() (platform string, ok bool)
}

func NewCVMSDetector() (*CVMSDetector, error) {
	return cvms.NewDetector()
}

func NewNitroDetector() (*NitroDetector, error) {
	return nitro.NewDetector()
}

func NewUnsafeDetector() (*UnsafeDetector, error) {
	return unsafe.NewDetector()
}
