package bearclave

import (
	"github.com/tahardi/bearclave/internal/cvms"
	"github.com/tahardi/bearclave/internal/nitro"
	"github.com/tahardi/bearclave/internal/unsafe"
)

type Detector interface {
	Detect() (platform string, ok bool)
}

const CVMSPlatform = cvms.Platform
const NitroPlatform = nitro.Platform
const UnsafePlatform = unsafe.Platform

type CVMSDetector = cvms.Detector
type NitroDetector = nitro.Detector
type UnsafeDetector = unsafe.Detector

func NewCVMSDetector() (*CVMSDetector, error) {
	return cvms.NewDetector()
}

func NewNitroDetector() (*NitroDetector, error) {
	return nitro.NewDetector()
}

func NewUnsafeDetector() (*UnsafeDetector, error) {
	return unsafe.NewDetector()
}
