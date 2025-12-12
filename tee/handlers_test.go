package tee_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tahardi/bearclave"
	"github.com/tahardi/bearclave/tee"

	"github.com/tahardi/bearclave/mocks"
)

func makeRequest(
	t *testing.T,
	method string,
	path string,
	body any,
) *http.Request {
	t.Helper()
	bodyBytes, err := json.Marshal(body)
	require.NoError(t, err)

	// nolint:noctx
	req, err := http.NewRequest(method, path, bytes.NewReader(bodyBytes))
	require.NoError(t, err)
	return req
}

func TestMakeAttestUserDataHandler(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// given
		data := []byte("hello world")
		attestation := &bearclave.AttestResult{Report: []byte("attestation")}
		attester := mocks.NewAttester(t)
		attester.
			On("Attest", mock.AnythingOfType("[]attestation.AttestOption")).
			Return(attestation, nil).Once()

		var logBuffer bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&logBuffer, nil))

		recorder := httptest.NewRecorder()
		body := tee.AttestUserDataRequest{Data: data}
		req := makeRequest(t, "POST", tee.AttestUserDataPath, body)

		handler := tee.MakeAttestUserDataHandler(attester, logger)

		// when
		handler.ServeHTTP(recorder, req)

		// then
		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Contains(t, logBuffer.String(), string(data))

		response := tee.AttestUserDataResponse{}
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
		req := makeRequest(t, "POST", tee.AttestUserDataPath, body)

		handler := tee.MakeAttestUserDataHandler(attester, logger)

		// when
		handler.ServeHTTP(recorder, req)

		// then
		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
		assert.Equal(t, 0, logBuffer.Len())
		assert.Contains(t, recorder.Body.String(), "decoding request")
	})

	t.Run("error - attesting userdata", func(t *testing.T) {
		// given
		data := []byte("hello world")
		attester := mocks.NewAttester(t)
		attester.
			On("Attest", mock.AnythingOfType("[]attestation.AttestOption")).
			Return(nil, assert.AnError).Once()

		var logBuffer bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&logBuffer, nil))

		recorder := httptest.NewRecorder()
		body := tee.AttestUserDataRequest{Data: data}
		req := makeRequest(t, "POST", tee.AttestUserDataPath, body)

		handler := tee.MakeAttestUserDataHandler(attester, logger)

		// when
		handler.ServeHTTP(recorder, req)

		// then
		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
		assert.Contains(t, logBuffer.String(), string(data))
		assert.Contains(t, recorder.Body.String(), "attesting userdata")
	})
}
