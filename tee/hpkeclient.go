package tee

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	neturl "net/url"
)

func FetchHPKEKey(
	ctx context.Context,
	client *http.Client,
	keyURL string,
	verifier *Verifier,
	policy KeyBindingPolicy,
) ([]byte, error) {
	nonce := make([]byte, HPKENonceSize)
	_, err := rand.Read(nonce)
	if err != nil {
		return nil, hpkeError("generating nonce", err)
	}

	u, err := neturl.Parse(keyURL)
	if err != nil {
		return nil, hpkeError("parsing key url", err)
	}
	query := u.Query()
	query.Set(HPKENonceParam, base64.RawURLEncoding.EncodeToString(nonce))
	u.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, hpkeError("making key request", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, hpkeError("fetching key", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, hpkeError("fetching key: status "+resp.Status, nil)
	}

	var att AttestResult
	err = json.NewDecoder(io.LimitReader(resp.Body, DefaultHPKEMaxBodyBytes)).Decode(&att)
	if err != nil {
		return nil, hpkeError("decoding key attestation", err)
	}

	policy.Nonce = nonce
	binding, err := VerifyKeyBinding(verifier, &att, HPKEKeyBindingPurpose, policy)
	if err != nil {
		return nil, err
	}
	return binding.PublicKey, nil
}

func DoHPKERequest(
	ctx context.Context,
	client *http.Client,
	url string,
	publicKey []byte,
	plaintext []byte,
) ([]byte, error) {
	u, err := neturl.Parse(url)
	if err != nil {
		return nil, hpkeError("parsing url", err)
	}
	path := u.Path
	if path == "" {
		path = "/"
	}

	msg, responseKey, err := SealHPKERequest(publicKey, plaintext, HPKEAAD(http.MethodPost, path))
	if err != nil {
		return nil, err
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return nil, hpkeError("marshaling request", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, hpkeError("making request", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, hpkeError("sending request", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, hpkeError("sending request: status "+resp.Status, nil)
	}

	var reply SealedMessage
	err = json.NewDecoder(io.LimitReader(resp.Body, DefaultHPKEMaxBodyBytes)).Decode(&reply)
	if err != nil {
		return nil, hpkeError("decoding response", err)
	}
	return responseKey.Open(reply)
}
