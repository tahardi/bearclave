package attestation

type Attester interface {
	Attest(options ...AttestOption) (result *AttestResult, err error)
}

type AttestResult struct {
	Report []byte `json:"report"`
}

type AttestOption func(*AttestOptions)
type AttestOptions struct {
	Nonce     []byte
	UserData  []byte
}

func MakeDefaultAttestOptions() AttestOptions {
	return AttestOptions{
		Nonce:     nil,
		UserData:  nil,
	}
}

func WithAttestNonce(nonce []byte) AttestOption {
	return func(opts *AttestOptions) {
		opts.Nonce = nonce
	}
}

func WithAttestUserData(userData []byte) AttestOption {
	return func(opts *AttestOptions) {
		opts.UserData = userData
	}
}
