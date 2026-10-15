package notee_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tahardi/bearclave/tee"
)

func TestNoTEE_AttestedTLS(t *testing.T) {
	// given
	ctx := context.Background()
	platform := tee.NoTEE
	logger := slog.New(slog.DiscardHandler)

	attester, err := tee.NewAttester(platform)
	require.NoError(t, err)

	verifier, err := tee.NewVerifier(platform)
	require.NoError(t, err)

	certProvider, err := tee.NewAttestedCertProvider(
		attester,
		tee.DefaultDomain,
		tee.DefaultIP,
		tee.DefaultAttestedCertValidity,
	)
	require.NoError(t, err)

	serverMux := http.NewServeMux()
	serverMux.HandleFunc("/hello", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("hello"))
	})

	serverTLSAddr := "https://127.0.0.1:8446"
	serverTLS, err := tee.NewServerTLS(
		ctx,
		platform,
		serverTLSAddr,
		serverMux,
		certProvider,
		logger,
	)
	require.NoError(t, err)
	defer serverTLS.Close()

	revProxyTLSAddr := "https://127.0.0.1:8445"
	revProxyTLS, err := tee.NewReverseProxyTLS(
		ctx,
		platform,
		revProxyTLSAddr,
		serverTLSAddr,
		logger,
	)
	require.NoError(t, err)
	defer revProxyTLS.Close()

	policy := tee.KeyBindingPolicy{Measurement: tee.NoTEEMeasurement, MaxAge: time.Hour}
	client := &http.Client{Transport: &http.Transport{
		TLSClientConfig:   tee.NewAttestedTLSConfig(verifier, policy),
		DisableKeepAlives: true,
	}}

	runService(func() { _ = serverTLS.Serve() }, 100*time.Millisecond)
	runService(func() { _ = revProxyTLS.Serve() }, 100*time.Millisecond)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, revProxyTLSAddr+"/hello", nil)
	require.NoError(t, err)

	// when
	resp, err := client.Do(req)

	// then
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, "hello", string(body))
}
