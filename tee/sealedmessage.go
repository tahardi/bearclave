package tee

type SealedMessage struct {
	Enc        []byte `json:"enc,omitempty"`
	Ciphertext []byte `json:"ciphertext"`
}
