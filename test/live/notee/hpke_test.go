package notee_test

import (
	"bytes"
	"context"
	"encoding/json"
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
	serverMux.HandleFunc("/hpke-key", tee.MakeFetchHPKEKeyHandler(attester, key, logger))
	serverMux.HandleFunc("/hpke-echo", tee.MakeSendSealedRequestHandler(key, upper, logger))
	serverMux.HandleFunc("/hpke-other", tee.MakeSendSealedRequestHandler(key, upper, logger))

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
	policy := tee.KeyBindingPolicy{Measurement: tee.NoTEEMeasurement, MaxAge: time.Minute}

	runService(func() { _ = server.Serve() }, 100*time.Millisecond)
	runService(func() { _ = revProxy.Serve() }, 100*time.Millisecond)

	t.Run("happy path - sealed round trip through reverse proxy", func(t *testing.T) {
		// given
		publicKey, err := tee.FetchHPKEKey(ctx, http.DefaultClient, revProxyAddr+"/hpke-key", verifier, policy)
		require.NoError(t, err)

		// when
		got, err := tee.SendSealedRequest(
			ctx,
			http.DefaultClient,
			revProxyAddr+"/hpke-echo",
			publicKey,
			[]byte("hello, sealed world"),
		)

		// then
		require.NoError(t, err)
		assert.Equal(t, []byte("HELLO, SEALED WORLD"), got)
	})

	t.Run("error - wrong measurement", func(t *testing.T) {
		// given
		wrongPolicy := tee.KeyBindingPolicy{Measurement: "wrong", MaxAge: time.Minute}

		// when
		_, err := tee.FetchHPKEKey(ctx, http.DefaultClient, revProxyAddr+"/hpke-key", verifier, wrongPolicy)

		// then
		require.ErrorIs(t, err, tee.ErrVerifierMeasurement)
	})

	t.Run("error - stale public key", func(t *testing.T) {
		// given
		older, err := tee.NewHPKEKey()
		require.NoError(t, err)

		// when
		_, err = tee.SendSealedRequest(
			ctx,
			http.DefaultClient,
			revProxyAddr+"/hpke-echo",
			older.PublicKey(),
			[]byte("hello, sealed world"),
		)

		// then
		require.ErrorIs(t, err, tee.ErrHPKE)
		assert.Contains(t, err.Error(), "400")
	})

	t.Run("error - sealed for another path", func(t *testing.T) {
		// given
		publicKey, err := tee.FetchHPKEKey(ctx, http.DefaultClient, revProxyAddr+"/hpke-key", verifier, policy)
		require.NoError(t, err)

		msg, _, err := tee.SealHPKERequest(
			publicKey,
			[]byte("hello, sealed world"),
			tee.HPKEAAD(http.MethodPost, "/hpke-echo"),
		)
		require.NoError(t, err)
		body, err := json.Marshal(msg)
		require.NoError(t, err)

		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodPost,
			revProxyAddr+"/hpke-other",
			bytes.NewReader(body),
		)
		require.NoError(t, err)

		// when
		resp, err := http.DefaultClient.Do(req)

		// then
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}
