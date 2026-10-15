package tee_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tahardi/bearclave/tee"
)

func TestFetchHPKEKey(t *testing.T) {
	tests := []struct {
		name        string
		platform    tee.Platform
		measurement string
		server      func(t *testing.T, attester *tee.Attester, key *tee.HPKEKey) *httptest.Server
		wantErr     error
	}{
		{
			name:        "happy path",
			platform:    tee.NoTEE,
			measurement: tee.NoTEEMeasurement,
			server:      realKeyServer,
		},
		{
			name:        "error - wrong measurement",
			platform:    tee.NoTEE,
			measurement: "wrong",
			server:      realKeyServer,
			wantErr:     tee.ErrVerifierMeasurement,
		},
		{
			name:        "error - replayed attestation",
			platform:    tee.NoTEE,
			measurement: tee.NoTEEMeasurement,
			server: func(t *testing.T, attester *tee.Attester, key *tee.HPKEKey) *httptest.Server {
				t.Helper()
				recorded, err := tee.AttestKeyBinding(attester, tee.KeyBinding{
					Purpose:   tee.HPKEKeyBindingPurpose,
					PublicKey: key.PublicKey(),
					Nonce:     []byte("old nonce"),
					IssuedAt:  time.Now().Unix(),
				})
				require.NoError(t, err)
				return fakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
					tee.WriteResponse(w, recorded)
				})
			},
			wantErr: tee.ErrKeyBindingNonce,
		},
		{
			name:        "error - wrong purpose",
			platform:    tee.NoTEE,
			measurement: tee.NoTEEMeasurement,
			server: func(t *testing.T, attester *tee.Attester, key *tee.HPKEKey) *httptest.Server {
				t.Helper()
				return fakeServer(t, func(w http.ResponseWriter, r *http.Request) {
					nonce, err := base64.RawURLEncoding.DecodeString(r.URL.Query().Get(tee.HPKENonceParam))
					if err != nil {
						http.Error(w, err.Error(), http.StatusBadRequest)
						return
					}
					att, err := tee.AttestKeyBinding(attester, tee.KeyBinding{
						Purpose:   tee.TLSKeyBindingPurpose,
						PublicKey: key.PublicKey(),
						Nonce:     nonce,
						IssuedAt:  time.Now().Unix(),
					})
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
					tee.WriteResponse(w, att)
				})
			},
			wantErr: tee.ErrKeyBindingPurpose,
		},
		{
			name:        "error - wrong platform",
			platform:    tee.Nitro,
			measurement: tee.NoTEEMeasurement,
			server:      realKeyServer,
			wantErr:     tee.ErrVerifier,
		},
		{
			name:        "error - server error",
			platform:    tee.NoTEE,
			measurement: tee.NoTEEMeasurement,
			server: func(t *testing.T, _ *tee.Attester, _ *tee.HPKEKey) *httptest.Server {
				t.Helper()
				return fakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
					http.Error(w, "boom", http.StatusInternalServerError)
				})
			},
			wantErr: tee.ErrHPKE,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			attester, err := tee.NewAttester(tee.NoTEE)
			require.NoError(t, err)
			verifier, err := tee.NewVerifier(tt.platform)
			require.NoError(t, err)
			key, err := tee.NewHPKEKey()
			require.NoError(t, err)
			server := tt.server(t, attester, key)
			policy := tee.KeyBindingPolicy{Measurement: tt.measurement, MaxAge: time.Minute}

			// when
			got, err := tee.FetchHPKEKey(
				context.Background(),
				server.Client(),
				server.URL+"/hpke-key",
				verifier,
				policy,
			)

			// then
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, key.PublicKey(), got)
		})
	}
}

func TestDoHPKERequest(t *testing.T) {
	plaintext := []byte("hello")
	tests := []struct {
		name       string
		publicKey  func(t *testing.T, key *tee.HPKEKey) []byte
		server     func(t *testing.T, attester *tee.Attester, key *tee.HPKEKey) *httptest.Server
		wantErr    error
		wantStatus string
	}{
		{
			name:      "happy path",
			publicKey: currentPublicKey,
			server:    realKeyServer,
		},
		{
			name: "error - stale public key",
			publicKey: func(t *testing.T, _ *tee.HPKEKey) []byte {
				t.Helper()
				older, err := tee.NewHPKEKey()
				require.NoError(t, err)
				return older.PublicKey()
			},
			server:     realKeyServer,
			wantErr:    tee.ErrHPKE,
			wantStatus: "400",
		},
		{
			name:      "error - tampered response",
			publicKey: currentPublicKey,
			server: func(t *testing.T, _ *tee.Attester, key *tee.HPKEKey) *httptest.Server {
				t.Helper()
				return fakeServer(t, func(w http.ResponseWriter, r *http.Request) {
					var msg tee.SealedMessage
					if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
						http.Error(w, err.Error(), http.StatusBadRequest)
						return
					}
					got, responseKey, err := key.Open(msg, tee.HPKEAAD(r.Method, r.URL.Path))
					if err != nil {
						http.Error(w, err.Error(), http.StatusBadRequest)
						return
					}
					reply := responseKey.Seal(got)
					reply.Ciphertext[len(reply.Ciphertext)-1] ^= 0xff
					tee.WriteResponse(w, reply)
				})
			},
			wantErr: tee.ErrHPKE,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			attester, err := tee.NewAttester(tee.NoTEE)
			require.NoError(t, err)
			key, err := tee.NewHPKEKey()
			require.NoError(t, err)
			server := tt.server(t, attester, key)
			publicKey := tt.publicKey(t, key)

			// when
			got, err := tee.DoHPKERequest(
				context.Background(),
				server.Client(),
				server.URL+"/echo",
				publicKey,
				plaintext,
			)

			// then
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Contains(t, err.Error(), tt.wantStatus)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, []byte("HELLO"), got)
		})
	}
}

func realKeyServer(t *testing.T, attester *tee.Attester, key *tee.HPKEKey) *httptest.Server {
	t.Helper()
	return newHPKEServer(t, attester, key, upperHandler)
}

func currentPublicKey(_ *testing.T, key *tee.HPKEKey) []byte {
	return key.PublicKey()
}

func fakeServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return server
}
