package networking_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tahardi/bearclave/internal/networking"
)

func writeError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func writeResponse(t *testing.T, w http.ResponseWriter, out any) {
	data, err := json.Marshal(out)
	require.NoError(t, err)

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(data)
	require.NoError(t, err)
}

func TestClient_AttestUserData(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// given
		data := []byte("hello world")
		want := []byte("attestation")

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodPost, r.Method)
			require.Contains(t, r.URL.Path, networking.AttestUserDataPath)

			req := networking.AttestUserDataRequest{}
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			assert.Equal(t, data, req.Data)

			resp := networking.AttestUserDataResponse{Attestation: want}
			writeResponse(t, w, resp)
		})

		server := httptest.NewServer(handler)
		defer server.Close()

		client := networking.NewClientWithClient(server.URL, server.Client())

		// when
		got, err := client.AttestUserData(data)

		// then
		assert.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("error - doing attest userdata request", func(t *testing.T) {
		// given
		data := []byte("data")

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			writeError(w, assert.AnError)
		})

		server := httptest.NewServer(handler)
		defer server.Close()

		client := networking.NewClientWithClient(server.URL, server.Client())

		// when
		_, err := client.AttestUserData(data)

		// then
		assert.ErrorContains(t, err, "doing attest userdata request")
	})
}
