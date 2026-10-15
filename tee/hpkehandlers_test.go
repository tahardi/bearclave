package tee_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tahardi/bearclave/tee"
)

var errHandlerFailed = errors.New("handler failed")

func upperHandler(_ context.Context, plaintext []byte) ([]byte, error) {
	return bytes.ToUpper(plaintext), nil
}

func failingHandler(_ context.Context, _ []byte) ([]byte, error) {
	return nil, errHandlerFailed
}

func newHPKEServer(
	t *testing.T,
	attester *tee.Attester,
	key *tee.HPKEKey,
	handler tee.HPKEHandlerFunc,
) *httptest.Server {
	t.Helper()
	logger := slog.New(slog.DiscardHandler)
	mux := http.NewServeMux()
	mux.HandleFunc("/hpke-key", tee.MakeHPKEKeyHandler(attester, key, logger))
	mux.HandleFunc("/echo", tee.MakeHPKEHandler(key, handler, logger))
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return server
}

func TestHPKEAAD(t *testing.T) {
	// given
	method := http.MethodPost
	path := "/echo"

	// when
	got := tee.HPKEAAD(method, path)

	// then
	assert.Equal(t, []byte("POST /echo"), got)
}

func TestMakeHPKEKeyHandler(t *testing.T) {
	validNonce := bytes.Repeat([]byte{7}, tee.HPKENonceSize)
	tests := []struct {
		name       string
		method     string
		query      string
		wantStatus int
	}{
		{
			name:       "happy path",
			method:     http.MethodGet,
			query:      "?nonce=" + base64.RawURLEncoding.EncodeToString(validNonce),
			wantStatus: http.StatusOK,
		},
		{
			name:       "error - missing nonce",
			method:     http.MethodGet,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "error - invalid nonce encoding",
			method:     http.MethodGet,
			query:      "?nonce=not*base64",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "error - nonce too long",
			method: http.MethodGet,
			query: "?nonce=" + base64.RawURLEncoding.EncodeToString(
				bytes.Repeat([]byte{7}, tee.MaxHPKENonceSize+1),
			),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "error - wrong method",
			method:     http.MethodPost,
			query:      "?nonce=" + base64.RawURLEncoding.EncodeToString(validNonce),
			wantStatus: http.StatusMethodNotAllowed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			attester, err := tee.NewAttester(tee.NoTEE)
			require.NoError(t, err)
			verifier, err := tee.NewVerifier(tee.NoTEE)
			require.NoError(t, err)
			key, err := tee.NewHPKEKey()
			require.NoError(t, err)
			server := newHPKEServer(t, attester, key, upperHandler)

			req, err := http.NewRequestWithContext(
				context.Background(),
				tt.method,
				server.URL+"/hpke-key"+tt.query,
				nil,
			)
			require.NoError(t, err)

			// when
			resp, err := server.Client().Do(req)

			// then
			require.NoError(t, err)
			defer resp.Body.Close()
			require.Equal(t, tt.wantStatus, resp.StatusCode)
			if tt.wantStatus != http.StatusOK {
				return
			}

			var att tee.AttestResult
			require.NoError(t, json.NewDecoder(resp.Body).Decode(&att))
			binding, err := tee.VerifyKeyBinding(
				verifier,
				&att,
				tee.HPKEKeyBindingPurpose,
				tee.KeyBindingPolicy{Measurement: tee.NoTEEMeasurement, Nonce: validNonce, MaxAge: time.Minute},
			)
			require.NoError(t, err)
			assert.Equal(t, key.PublicKey(), binding.PublicKey)
		})
	}
}

func TestMakeHPKEHandler(t *testing.T) {
	plaintext := []byte("hello")
	tests := []struct {
		name       string
		method     string
		handler    tee.HPKEHandlerFunc
		body       func(t *testing.T, key *tee.HPKEKey) ([]byte, *tee.HPKEResponseKey)
		wantStatus int
		wantBody   string
	}{
		{
			name:       "happy path",
			method:     http.MethodPost,
			handler:    upperHandler,
			body:       sealedBody(plaintext, "/echo", nil),
			wantStatus: http.StatusOK,
		},
		{
			name:    "error - tampered ciphertext",
			method:  http.MethodPost,
			handler: upperHandler,
			body: sealedBody(plaintext, "/echo", func(msg *tee.SealedMessage) {
				msg.Ciphertext[len(msg.Ciphertext)-1] ^= 0xff
			}),
			wantStatus: http.StatusBadRequest,
			wantBody:   "opening request",
		},
		{
			name:       "error - sealed for another path",
			method:     http.MethodPost,
			handler:    upperHandler,
			body:       sealedBody(plaintext, "/other", nil),
			wantStatus: http.StatusBadRequest,
			wantBody:   "opening request",
		},
		{
			name:    "error - body too large",
			method:  http.MethodPost,
			handler: upperHandler,
			body: func(_ *testing.T, _ *tee.HPKEKey) ([]byte, *tee.HPKEResponseKey) {
				return bytes.Repeat([]byte("a"), 2*tee.Megabyte), nil
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "error - handler fails",
			method:     http.MethodPost,
			handler:    failingHandler,
			body:       sealedBody(plaintext, "/echo", nil),
			wantStatus: http.StatusInternalServerError,
			wantBody:   "handling request",
		},
		{
			name:       "error - wrong method",
			method:     http.MethodGet,
			handler:    upperHandler,
			body:       sealedBody(plaintext, "/echo", nil),
			wantStatus: http.StatusMethodNotAllowed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			attester, err := tee.NewAttester(tee.NoTEE)
			require.NoError(t, err)
			key, err := tee.NewHPKEKey()
			require.NoError(t, err)
			server := newHPKEServer(t, attester, key, tt.handler)
			body, responseKey := tt.body(t, key)

			req, err := http.NewRequestWithContext(
				context.Background(),
				tt.method,
				server.URL+"/echo",
				bytes.NewReader(body),
			)
			require.NoError(t, err)

			// when
			resp, err := server.Client().Do(req)

			// then
			require.NoError(t, err)
			defer resp.Body.Close()
			require.Equal(t, tt.wantStatus, resp.StatusCode)
			if tt.wantStatus != http.StatusOK {
				if tt.wantBody != "" {
					got, readErr := io.ReadAll(resp.Body)
					require.NoError(t, readErr)
					assert.Equal(t, tt.wantBody, strings.TrimSpace(string(got)))
				}
				return
			}

			var reply tee.SealedMessage
			require.NoError(t, json.NewDecoder(resp.Body).Decode(&reply))
			got, err := responseKey.Open(reply)
			require.NoError(t, err)
			assert.Equal(t, bytes.ToUpper(plaintext), got)
		})
	}
}

func sealedBody(
	plaintext []byte,
	path string,
	mutate func(msg *tee.SealedMessage),
) func(t *testing.T, key *tee.HPKEKey) ([]byte, *tee.HPKEResponseKey) {
	return func(t *testing.T, key *tee.HPKEKey) ([]byte, *tee.HPKEResponseKey) {
		t.Helper()
		msg, responseKey, err := tee.SealHPKERequest(
			key.PublicKey(),
			plaintext,
			tee.HPKEAAD(http.MethodPost, path),
		)
		require.NoError(t, err)
		if mutate != nil {
			mutate(&msg)
		}
		body, err := json.Marshal(msg)
		require.NoError(t, err)
		return body, responseKey
	}
}
