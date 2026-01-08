package notee_test

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tahardi/bearclave/tee"
)

const (
	AttestCertPath      = "/attest-cert"
	AttestHTTPCallPath  = "/attest-http-call"
	AttestHTTPSCallPath = "/attest-https-call"
)

type AttestCertRequest struct {}
type AttestCertResponse struct {
	Attestation *tee.AttestResult `json:"attestation"`
}

func MakeAttestCertHandler(
	t *testing.T,
	attester *tee.Attester,
	certProvider tee.CertProvider,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t.Helper()
		certReq := AttestCertRequest{}
		err := json.NewDecoder(r.Body).Decode(&certReq)
		require.NoError(t, err)

		cert, err := certProvider.GetCert(r.Context())
		require.NoError(t, err)

		chainDER := [][]byte{}
		for _, certBytes := range cert.Certificate {
			chainDER = append(chainDER, certBytes)
		}
		chainJSON, err := json.Marshal(chainDER)
		require.NoError(t, err)

		att, err := attester.Attest(tee.WithAttestUserData(chainJSON))
		require.NoError(t, err)
		tee.WriteResponse(w, AttestCertResponse{Attestation: att})
	}
}

type AttestHTTPCallRequest struct {
	Method string `json:"method"`
	URL    string `json:"url"`
}

type AttestHTTPCallResponse struct {
	Attestation *tee.AttestResult `json:"attestation"`
}

func MakeAttestHTTPCallHandler(
	t *testing.T,
	attester *tee.Attester,
	client *http.Client,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t.Helper()
		httpCallReq := AttestHTTPCallRequest{}
		err := json.NewDecoder(r.Body).Decode(&httpCallReq)
		require.NoError(t, err)

		req, err := http.NewRequestWithContext(
			r.Context(),
			httpCallReq.Method,
			httpCallReq.URL,
			nil,
		)
		require.NoError(t, err)

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		respBytes, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		attestation, err := attester.Attest(tee.WithAttestUserData(respBytes))
		require.NoError(t, err)

		httpCallResp := AttestHTTPCallResponse{
			Attestation: attestation,
		}
		tee.WriteResponse(w, httpCallResp)
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
	return func(w http.ResponseWriter, r *http.Request) {
		t.Helper()
		httpsCallReq := AttestHTTPSCallRequest{}
		err := json.NewDecoder(r.Body).Decode(&httpsCallReq)
		require.NoError(t, err)

		req, err := http.NewRequestWithContext(
			r.Context(),
			httpsCallReq.Method,
			httpsCallReq.URL,
			nil,
		)
		require.NoError(t, err)

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		respBytes, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		attestation, err := attester.Attest(tee.WithAttestUserData(respBytes))
		require.NoError(t, err)

		httpsCallResp := AttestHTTPSCallResponse{
			Attestation: attestation,
		}
		tee.WriteResponse(w, httpsCallResp)
	}
}
