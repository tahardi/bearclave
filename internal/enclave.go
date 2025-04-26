package internal

import (
	"fmt"

	"github.com/tahardi/bearclave/internal/cvms"
	"github.com/tahardi/bearclave/internal/natee"
	"github.com/tahardi/bearclave/internal/nitro"
)

type PlatformDetector interface {
	Detect() (platform string, ok bool)
}

type Attester interface {
	Attest(userdata []byte) (attestation []byte, err error)
}

// TODO: Put these NewX interface funcs inside the app not here in the mod
func NewAttester(detectors []PlatformDetector) (Attester, error) {
	for _, detector := range detectors {
		platform, ok := detector.Detect()
		switch {
		case !ok:
			continue
		case platform == cvms.Platform:
			return cvms.NewAttester()
		case platform == natee.Platform:
			return natee.NewAttester()
		case platform == nitro.Platform:
			return nitro.NewAttester()
		}
	}
	return nil, fmt.Errorf("no supported platforms detected")
}

type Verifier interface {
	Verify(attestation []byte) (userdata []byte, err error)
}

type Communicator interface {
	Send(data []byte) (err error)
	Receive() (data []byte, err error)
}
