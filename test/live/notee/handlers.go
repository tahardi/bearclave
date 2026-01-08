package notee_test

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tahardi/bearclave/tee"
)

const (
	AttestCertPath      = "/attest-cert"
	AttestUserDataPath  = "/attest-user-data"
	AttestHTTPSCallPath = "/attest-https-call"
)

type AttestCertRequest struct {
	Nonce []byte `json:"nonce,omitempty"`
}
type AttestCertResponse struct {
	Attestation *tee.AttestResult `json:"attestation"`
}

func MakeAttestCertHandler(
	t *testing.T,
	attester *tee.Attester,
	certProvider tee.CertProvider,
) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		req := AttestCertRequest{}
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)

		ctxt := r.Context()
		cert, err := certProvider.GetCert(ctxt)
		assert.NoError(t, err)

		var chainDER [][]byte
		for _, certBytes := range cert.Certificate {
			chainDER = append(chainDER, certBytes)
		}
		chainJSON, err := json.Marshal(chainDER)
		assert.NoError(t, err)

		att, err := attester.Attest(
			tee.WithAttestNonce(req.Nonce),
			tee.WithAttestUserData(chainJSON),
		)
		assert.NoError(t, err)
		tee.WriteResponse(w, AttestCertResponse{Attestation: att})
	}
}

type AttestUserDataRequest struct {
	Nonce    []byte `json:"nonce,omitempty"`
	UserData []byte `json:"userdata,omitempty"`
}
type AttestUserDataResponse struct {
	Attestation *tee.AttestResult `json:"attestation"`
}

func MakeAttestUserDataHandler(
	t *testing.T,
	attester *tee.Attester,
) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		req := AttestUserDataRequest{}
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)

		att, err := attester.Attest(
			tee.WithAttestNonce(req.Nonce),
			tee.WithAttestUserData(req.UserData),
		)
		assert.NoError(t, err)
		tee.WriteResponse(w, AttestUserDataResponse{Attestation: att})
	}
}

type AttestHTTPSCallRequest struct {
	Method string `json:"method"`
	URL    string `json:"url"`
}

type AttestHTTPSCallResponse struct {
	Attestation *tee.AttestResult `json:"attestation"`
}

func MakeAttestHTTPSCallHandler(
	t *testing.T,
	attester *tee.Attester,
	client *http.Client,
) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		apiCallReq := AttestHTTPSCallRequest{}
		err := json.NewDecoder(r.Body).Decode(&apiCallReq)
		assert.NoError(t, err)

		req, err := http.NewRequestWithContext(r.Context(), apiCallReq.Method, apiCallReq.URL, nil)
		assert.NoError(t, err)

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		respBytes, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)

		attestation, err := attester.Attest(tee.WithAttestUserData(respBytes))
		assert.NoError(t, err)

		apiCallResp := AttestHTTPSCallResponse{
			Attestation: attestation,
		}
		tee.WriteResponse(w, apiCallResp)
	}
}
