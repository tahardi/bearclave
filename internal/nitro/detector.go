package nitro

const Platform = "Nitro"

type Detector struct{}

func NewDetector() (*Detector, error) {
	return &Detector{}, nil
}

func (n *Detector) Detect() (string, bool) {
	return Platform, false
}
