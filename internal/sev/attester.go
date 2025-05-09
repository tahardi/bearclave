package sev

import (
	"fmt"

	"github.com/google/go-sev-guest/client"
)

const AMD_SEV_USERDATA_SIZE = 64

type Attester struct{}

func NewAttester() (*Attester, error) {
	return &Attester{}, nil
}

func (n *Attester) Attest(userdata []byte) ([]byte, error) {
	if len(userdata) > AMD_SEV_USERDATA_SIZE {
		return nil, fmt.Errorf(
			"userdata must be less than %d bytes",
			AMD_SEV_USERDATA_SIZE,
		)
	}

	sevQP, err := client.GetQuoteProvider()
	if err != nil {
		return nil, fmt.Errorf("getting quote provider: %w", err)
	}

	var reportData [64]byte
	copy(reportData[:], userdata)
	attestation, err := sevQP.GetRawQuote(reportData)
	if err != nil {
		return nil, fmt.Errorf("getting quote: %w", err)
	}
	return attestation, nil
}
