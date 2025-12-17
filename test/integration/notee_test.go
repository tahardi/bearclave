package integration_test

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tahardi/bearclave"
	"github.com/tahardi/bearclave/tee"
)

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

func TestIntegration_NoTEE(t *testing.T) {
	// given
	ctx := context.Background()
	platform := bearclave.NoTEE
	attester, err := bearclave.NewAttester(platform)
	require.NoError(t, err)

	discardLogger := slog.New(slog.DiscardHandler)
	serverMux := http.NewServeMux()
	serverMux.Handle(
		"POST "+tee.AttestPath,
		tee.MakeAttestHandler(attester, discardLogger),
	)

	serverAddr := "http://127.0.0.1:8081"
	server, err := tee.NewServer(ctx, platform, "tcp", serverAddr, serverMux)
	require.NoError(t, err)
	defer server.Close()

	route := "app/v1"
	proxy, err := tee.NewReverseProxy(platform, serverAddr, route)
	require.NoError(t, err)

	// #nosec G112 - this is just a test server. Ignore security warning
	proxyServer := &http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: proxy,
	}
	defer proxyServer.Close()

	client := tee.NewClient("http://127.0.0.1:8080")
	nonce := []byte("nonce")
	want := []byte("hello world")

	// when
	runService(func() { _ = server.ListenAndServe() }, 100*time.Millisecond)
	runService(func() { _ = proxyServer.ListenAndServe() }, 100*time.Millisecond)
	attestation, err := client.Attest(ctx, nonce, want)
	require.NoError(t, err)

	// then
	verifier, err := bearclave.NewVerifier(platform)
	require.NoError(t, err)

	got, err := verifier.Verify(attestation, bearclave.WithVerifyNonce(nonce))
	require.NoError(t, err)
	require.Equal(t, want, got.UserData)
}
