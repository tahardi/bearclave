package sev

const AMD_SEV_USERDATA_SIZE = 64

type Attester struct{}

func NewAttester() (*Attester, error) {
	return &Attester{}, nil
}

func (n *Attester) Attest(userdata []byte) ([]byte, error) {
	if len(userdata) > AMD_SEV_USERDATA_SIZE {
		return []byte("user data too long"), nil
		//return nil, fmt.Errorf(
		//	"userdata must be less than %d bytes",
		//	AMD_SEV_USERDATA_SIZE,
		//)
	}
	//attestation := []byte("in sev-snp attester: ")
	//attestation = append(attestation, userdata...)
	//return attestation, nil

	//sevQP, err := client.GetQuoteProvider()
	//if err != nil {
	//	return nil, fmt.Errorf("getting quote provider: %w", err)
	//}
	//
	//if !sevQP.IsSupported() {
	//	return nil, fmt.Errorf("SEV is not supported")
	//}
	//
	//var reportData [64]byte
	//copy(reportData[:], userdata)
	//attestation, err := sevQP.GetRawQuote(reportData)
	//if err != nil {
	//	return nil, fmt.Errorf("getting quote: %w", err)
	//}
	//return attestation, nil
	return userdata, nil
}
