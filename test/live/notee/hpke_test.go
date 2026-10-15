package notee_test

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tahardi/bearclave/tee"
)

func TestNoTEE_HPKE(t *testing.T) {
	// given
	ctx := context.Background()
	platform := tee.NoTEE
	logger := slog.New(slog.DiscardHandler)

	attester, err := tee.NewAttester(platform)
	require.NoError(t, err)

	key, err := tee.NewHPKEKey()
	require.NoError(t, err)

	upper := func(_ context.Context, plaintext []byte) ([]byte, error) {
		return bytes.ToUpper(plaintext), nil
	}

	serverMux := http.NewServeMux()
	serverMux.HandleFunc("/hpke-key", tee.MakeHPKEKeyHandler(attester, key, logger))
	serverMux.HandleFunc("/hpke-echo", tee.MakeHPKEHandler(key, upper, logger))

	// given - a server that opens sealed requests and seals its replies
	serverAddr := "http://127.0.0.1:8085"
	server, err := tee.NewServer(ctx, platform, serverAddr, serverMux, logger)
	require.NoError(t, err)
	defer server.Close()

	// given - a reverse proxy that forwards a remote client's requests to the server
	revProxyAddr := "http://127.0.0.1:8086"
	revProxy, err := tee.NewReverseProxy(
		ctx,
		platform,
		revProxyAddr,
		serverAddr,
		logger,
	)
	require.NoError(t, err)
	defer revProxy.Close()

	verifier, err := tee.NewVerifier(platform)
	require.NoError(t, err)

	runService(func() { _ = server.Serve() }, 100*time.Millisecond)
	runService(func() { _ = revProxy.Serve() }, 100*time.Millisecond)

	// when
	publicKey, err := tee.FetchHPKEKey(
		ctx,
		http.DefaultClient,
		revProxyAddr+"/hpke-key",
		verifier,
		tee.KeyBindingPolicy{Measurement: tee.NoTEEMeasurement, MaxAge: time.Minute},
	)
	require.NoError(t, err)

	got, err := tee.DoHPKERequest(
		ctx,
		http.DefaultClient,
		revProxyAddr+"/hpke-echo",
		publicKey,
		[]byte("hello, sealed world"),
	)

	// then
	require.NoError(t, err)
	assert.Equal(t, []byte("HELLO, SEALED WORLD"), got)
}
