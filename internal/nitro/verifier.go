package nitro

import (
	"fmt"
	"time"

	"github.com/hf/nitrite"
)

type Verifier struct{}

func NewVerifier() (*Verifier, error) {
	return &Verifier{}, nil
}

func (n *Verifier) Verify(attestation []byte) ([]byte, error) {
	resp, err := nitrite.Verify(
		attestation,
		nitrite.VerifyOptions{
			CurrentTime: time.Now(),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("verifying attestation: %w", err)
	}

	return resp.Document.UserData, nil
}
