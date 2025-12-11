package attestation

type Attester interface {
	Attest(options ...AttestOption) (result *AttestResult, err error)
}

type AttestResult struct {
	Report []byte `json:"report"`
}

type AttestOption func(*AttestOptions)
type AttestOptions struct {
	nonce     []byte
	publicKey []byte
	userData  []byte
}

func WithAttestNonce(nonce []byte) AttestOption {
	return func(opts *AttestOptions) {
		opts.nonce = nonce
	}
}

func WithPublicKey(publicKey []byte) AttestOption {
	return func(opts *AttestOptions) {
		opts.publicKey = publicKey
	}
}

func WithUserData(userData []byte) AttestOption {
	return func(opts *AttestOptions) {
		opts.userData = userData
	}
}
