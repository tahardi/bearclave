package integration_test

//revive:disable:context-as-argument

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tahardi/bearclave/tee"
)

const attestPath = "/attest"

//go:embed testdata/userdata.txt
var userdata []byte

func doRequest(
	t *testing.T,
	ctx context.Context,
	client *http.Client,
	host string,
	method string,
	api string,
	apiReq any,
	apiResp any,
) {
	t.Helper()
	bodyBytes, err := json.Marshal(apiReq)
	require.NoError(t, err)

	url := host + api
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(bodyBytes))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	defer resp.Body.Close()

	bodyBytes, err = io.ReadAll(resp.Body)
	require.NoError(t, err)

	err = json.Unmarshal(bodyBytes, apiResp)
	require.NoError(t, err)
}

func makeAttestRequest(
	t *testing.T,
	ctx context.Context,
	client *http.Client,
	host string,
	nonce []byte,
	userData []byte,
) *tee.AttestResult {
	t.Helper()
	attReq := attestRequest{Nonce: nonce, UserData: userData}
	attRes := attestResponse{}
	doRequest(
		t,
		ctx,
		client,
		host,
		"POST",
		attestPath,
		attReq,
		&attRes,
	)
	return attRes.Attestation
}

type attestRequest struct {
	Nonce    []byte `json:"nonce,omitempty"`
	UserData []byte `json:"userdata,omitempty"`
}
type attestResponse struct {
	Attestation *tee.AttestResult `json:"attestation"`
}

func makeAttestHandler(
	t *testing.T,
	attester *tee.Attester,
	logger *slog.Logger,
) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		req := attestRequest{}
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)

		logger.Info(
			"attesting",
			slog.String("nonce", base64.StdEncoding.EncodeToString(req.Nonce)),
			slog.String("userdata", base64.StdEncoding.EncodeToString(req.UserData)),
		)
		att, err := attester.Attest(
			tee.WithAttestNonce(req.Nonce),
			tee.WithAttestUserData(req.UserData),
		)
		assert.NoError(t, err)
		tee.WriteResponse(w, attestResponse{Attestation: att})
	}
}

func runService(service func(), wait time.Duration) {
	var serviceReady sync.WaitGroup
	serviceReady.Add(1)
	go func() {
		serviceReady.Done()
		service()
	}()
	serviceReady.Wait()
	time.Sleep(wait)
}

func TestIntegration_TEE(t *testing.T) {
	t.Run("happy path - NoTEE platform", func(t *testing.T) {
		// given
		ctx := context.Background()
		platform := tee.NoTEE
		attester, err := tee.NewAttester(platform)
		require.NoError(t, err)

		discardLogger := slog.New(slog.DiscardHandler)
		serverMux := http.NewServeMux()
		serverMux.Handle(
			"POST "+attestPath,
			makeAttestHandler(t, attester, discardLogger),
		)

		network := "tcp"
		enclaveAddr := "http://127.0.0.1:8081"
		server, err := tee.NewServer(ctx, platform, network, enclaveAddr, serverMux)
		require.NoError(t, err)
		defer server.Close()

		proxyAddr := "http://127.0.0.1:8080"
		route := "app/v1"
		proxy, err := tee.NewReverseProxy(
			ctx, platform, network, proxyAddr, enclaveAddr, route,
		)
		require.NoError(t, err)
		defer proxy.Close()

		client := &http.Client{}
		nonce := []byte("nonce")
		want := userdata

		// when
		runService(func() { _ = server.ListenAndServe() }, 100*time.Millisecond)
		runService(func() { _ = proxy.ListenAndServe() }, 100*time.Millisecond)
		attestation := makeAttestRequest(t, ctx, client, proxyAddr, nonce, want)

		// then
		verifier, err := tee.NewVerifier(platform)
		require.NoError(t, err)

		got, err := verifier.Verify(attestation, tee.WithVerifyNonce(nonce))
		require.NoError(t, err)
		require.Equal(t, want, got.UserData)
	})
}
