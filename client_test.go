package bearclave_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/tahardi/bearclave"
	"github.com/tahardi/bearclave/internal/mocks"
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
			require.Contains(t, r.URL.Path, bearclave.AttestUserDataPath)

			req := bearclave.AttestUserDataRequest{}
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			assert.Equal(t, data, req.Data)

			resp := bearclave.AttestUserDataResponse{Attestation: want}
			writeResponse(t, w, resp)
		})

		server := httptest.NewServer(handler)
		defer server.Close()

		client := bearclave.NewClientWithClient(server.URL, server.Client())

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

		client := bearclave.NewClientWithClient(server.URL, server.Client())

		// when
		_, err := client.AttestUserData(data)

		// then
		assert.ErrorContains(t, err, "doing attest userdata request")
	})
}

type doRequest struct {
	Data []byte `json:"data"`
}

type doResponse struct {
	Data []byte `json:"data"`
}

func TestClient_Do(t *testing.T) {
	t.Run("happy path - GET", func(t *testing.T) {
		// given
		want := []byte("data")
		method := http.MethodGet
		api := "/"
		apiReq := &doRequest{}
		apiResp := &doResponse{}

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, method, r.Method)

			resp := doResponse{Data: want}
			writeResponse(t, w, resp)
		})

		server := httptest.NewServer(handler)
		defer server.Close()

		client := bearclave.NewClientWithClient(server.URL, server.Client())

		// when
		err := client.Do(method, api, apiReq, apiResp)

		// then
		assert.NoError(t, err)
		assert.Equal(t, want, apiResp.Data)
	})

	t.Run("happy path - POST", func(t *testing.T) {
		// given
		want := []byte("data")
		method := http.MethodPost
		api := "/"
		apiReq := &doRequest{Data: want}
		apiResp := &doResponse{}

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, method, r.Method)

			req := doRequest{}
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			assert.Equal(t, want, req.Data)

			resp := doResponse{}
			writeResponse(t, w, resp)
		})

		server := httptest.NewServer(handler)
		defer server.Close()

		client := bearclave.NewClientWithClient(server.URL, server.Client())

		// when
		err := client.Do(method, api, apiReq, apiResp)

		// then
		assert.NoError(t, err)
	})

	t.Run("error - creating request", func(t *testing.T) {
		// given
		method := "invalid method"
		api := "/"
		apiReq := &doRequest{}
		apiResp := &doResponse{}

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, method, r.Method)

			resp := doResponse{}
			writeResponse(t, w, resp)
		})

		server := httptest.NewServer(handler)
		defer server.Close()

		client := bearclave.NewClientWithClient(server.URL, server.Client())

		// when
		err := client.Do(method, api, apiReq, apiResp)

		// then
		assert.ErrorContains(t, err, "creating request")
	})

	t.Run("error - sending request", func(t *testing.T) {
		// given
		method := http.MethodPost
		api := "/"
		apiReq := &doRequest{}
		apiResp := &doResponse{}

		roundTripper := mocks.NewRoundTripper(t)
		roundTripper.On("RoundTrip", mock.Anything).Return(nil, assert.AnError)

		httpClient := &http.Client{Transport: roundTripper}
		client := bearclave.NewClientWithClient("127.0.0.1", httpClient)

		// when
		err := client.Do(method, api, apiReq, apiResp)

		// then
		assert.ErrorContains(t, err, "sending request")
	})

	t.Run("error - received non-200 response", func(t *testing.T) {
		// given
		method := http.MethodPost
		api := "/"
		apiReq := &doRequest{}
		apiResp := &doResponse{}

		httpResp := &http.Response{
			StatusCode: http.StatusInternalServerError,
		}

		roundTripper := mocks.NewRoundTripper(t)
		roundTripper.On("RoundTrip", mock.Anything).Return(httpResp, nil)

		httpClient := &http.Client{Transport: roundTripper}
		client := bearclave.NewClientWithClient("127.0.0.1", httpClient)

		// when
		err := client.Do(method, api, apiReq, apiResp)

		// then
		assert.ErrorContains(t, err, "received non-200 response")
	})

	t.Run("error - reading response body", func(t *testing.T) {
		// given
		method := http.MethodPost
		api := "/"
		apiReq := &doRequest{}
		apiResp := &doResponse{}

		readCloser := mocks.NewReadCloser(t)
		readCloser.On("Read", mock.Anything).Return(0, assert.AnError)
		readCloser.On("Close").Return(nil)

		httpResp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       readCloser,
		}

		roundTripper := mocks.NewRoundTripper(t)
		roundTripper.On("RoundTrip", mock.Anything).Return(httpResp, nil)

		httpClient := &http.Client{Transport: roundTripper}
		client := bearclave.NewClientWithClient("127.0.0.1", httpClient)

		// when
		err := client.Do(method, api, apiReq, apiResp)

		// then
		assert.ErrorContains(t, err, "reading response body")
	})
}
