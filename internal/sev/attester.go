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
		msg := fmt.Sprintf("userdata must be less than %d bytes", AMD_SEV_USERDATA_SIZE)
		return []byte(msg), nil
		//return nil, fmt.Errorf(
		//	"userdata must be less than %d bytes",
		//	AMD_SEV_USERDATA_SIZE,
		//)
	}

	if !client.UseDefaultSevGuest() {
		msg := fmt.Sprintf("Not using default SEV guest")
		return []byte(msg), nil
	}

	dev := &client.LinuxDevice{}
	if err := dev.Open("/dev/sev-guest"); err != nil {
		msg := fmt.Sprintf("opening device: %s", err)
		return []byte(msg), nil
	}
	dev.Close()

	// FAILING TO GET QUOTE PROVIDER WHY???
	// search their code for this string: no supported SEV-SNP QuoteProvider found
	// ALSO need to figure out how to debug on confidential VM
	sevQP, err := client.GetQuoteProvider()
	if err != nil {
		msg := fmt.Sprintf("error getting quote provider: %s", err)
		return []byte(msg), nil
		//return nil, fmt.Errorf("getting quote provider: %w", err)
	}

	if !sevQP.IsSupported() {
		msg := fmt.Sprintf("SEV is not supported")
		return []byte(msg), nil
		//return nil, fmt.Errorf("SEV is not supported")
	}

	var reportData [64]byte
	copy(reportData[:], userdata)
	attestation, err := sevQP.GetRawQuote(reportData)
	if err != nil {
		msg := fmt.Sprintf("error getting quote: %s", err)
		return []byte(msg), nil
		//return nil, fmt.Errorf("getting quote: %w", err)
	}
	return attestation, nil
	//return userdata, nil
}
