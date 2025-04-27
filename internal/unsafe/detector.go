package unsafe

const Platform = "Unsafe"

type Detector struct{}

func NewDetector() (*Detector, error) {
	return &Detector{}, nil
}

func (n *Detector) Detect() (string, bool) {
	return Platform, true
}
