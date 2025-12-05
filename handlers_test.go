package bearclave_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tahardi/bearclave"
	"github.com/tahardi/bearclave/internal/mocks"
)

func makeRequest(
	t *testing.T,
	method string,
	path string,
	body any,
) *http.Request {
	bodyBytes, err := json.Marshal(body)
	require.NoError(t, err)
	req, err := http.NewRequest(method, path, bytes.NewReader(bodyBytes))
	require.NoError(t, err)
	return req
}

func TestMakeAttestUserDataHandler(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// given
		data := []byte("hello world")
		attestation := []byte("attestation")
		attester := mocks.NewAttester(t)
		attester.On("Attest", data).Return(attestation, nil).Once()

		var logBuffer bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&logBuffer, nil))

		recorder := httptest.NewRecorder()
		body := bearclave.AttestUserDataRequest{Data: data}
		req := makeRequest(t, "POST", bearclave.AttestUserDataPath, body)

		handler := bearclave.MakeAttestUserDataHandler(attester, logger)

		// when
		handler.ServeHTTP(recorder, req)

		// then
		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Contains(t, logBuffer.String(), string(data))

		response := bearclave.AttestUserDataResponse{}
		err := json.NewDecoder(recorder.Body).Decode(&response)
		require.NoError(t, err)
		assert.Equal(t, attestation, response.Attestation)
	})

	t.Run("error - decoding request", func(t *testing.T) {
		// given
		attester := mocks.NewAttester(t)

		var logBuffer bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&logBuffer, nil))

		recorder := httptest.NewRecorder()
		body := []byte("invalid json")
		req := makeRequest(t, "POST", bearclave.AttestUserDataPath, body)

		handler := bearclave.MakeAttestUserDataHandler(attester, logger)

		// when
		handler.ServeHTTP(recorder, req)

		// then
		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
		assert.Equal(t, logBuffer.Len(), 0)
		assert.Contains(t, recorder.Body.String(), "decoding request")
	})

	t.Run("error - attesting userdata", func(t *testing.T) {
		// given
		data := []byte("hello world")
		attester := mocks.NewAttester(t)
		attester.On("Attest", data).Return(nil, assert.AnError).Once()

		var logBuffer bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&logBuffer, nil))

		recorder := httptest.NewRecorder()
		body := bearclave.AttestUserDataRequest{Data: data}
		req := makeRequest(t, "POST", bearclave.AttestUserDataPath, body)

		handler := bearclave.MakeAttestUserDataHandler(attester, logger)

		// when
		handler.ServeHTTP(recorder, req)

		// then
		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
		assert.Contains(t, logBuffer.String(), string(data))
		assert.Contains(t, recorder.Body.String(), "attesting userdata")
	})
}
