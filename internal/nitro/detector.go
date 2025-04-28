package nitro

import "os"

const Platform = "Nitro"

type Detector struct{}

func NewDetector() (*Detector, error) {
	return &Detector{}, nil
}

func (n *Detector) Detect() (string, bool) {
	if _, err := os.Stat("/dev/nsm"); err == nil {
		return Platform, true
	}
	return Platform, false
}
