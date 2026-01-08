package tee_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tahardi/bearclave/tee"
)

const defaultPath = "/"

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

func TestMakeProxyHandler(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// given
		want := map[string]string{"status": "ok"}
		backend := httptest.NewServer(http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(want)
			}),
		)
		defer backend.Close()

		ctxTimeout := tee.DefaultProxyTimeout
		client := backend.Client()
		var logBuffer bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&logBuffer, nil))

		recorder := httptest.NewRecorder()
		body := map[string]string{"hello": "world"}
		req := makeRequest(t, "POST", defaultPath, body)
		req.Host = backend.Listener.Addr().String()

		handler := tee.MakeProxyHandler(client, logger, ctxTimeout)

		// when
		handler.ServeHTTP(recorder, req)

		// then
		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Contains(t, logBuffer.String(), "forwarding request")
		assert.Contains(t, logBuffer.String(), backend.URL)

		var got map[string]string
		err := json.NewDecoder(recorder.Body).Decode(&got)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("happy path - ignored headers", func(t *testing.T) {
		// given
		ignoredHeaders := []string{
			"Connection",
			"Keep-Alive",
			"Proxy-Authenticate",
			"Proxy-Authorization",
			"Te",
			"Trailers",
			"Transfer-Encoding",
			"Upgrade",
		}

		backend := httptest.NewServer(http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				for _, h := range ignoredHeaders {
					assert.Empty(t, r.Header.Get(h), "Header %s should have been ignored", h)
				}
				assert.Equal(t, "allowed-value", r.Header.Get("X-Custom-Header"))
				w.WriteHeader(http.StatusOK)
			}),
		)
		defer backend.Close()

		client := backend.Client()
		var logBuffer bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&logBuffer, nil))

		recorder := httptest.NewRecorder()
		req := makeRequest(t, "GET", defaultPath, nil)
		req.Host = backend.Listener.Addr().String()

		for _, h := range ignoredHeaders {
			req.Header.Set(h, "should-be-ignored")
		}
		req.Header.Set("X-Custom-Header", "allowed-value")

		handler := tee.MakeProxyHandler(client, logger, tee.DefaultProxyTimeout)

		// when
		handler.ServeHTTP(recorder, req)

		// then
		assert.Equal(t, http.StatusOK, recorder.Code)
	})

	t.Run("error - context deadline exceeded", func(t *testing.T) {
		// given
		timeout := 10 * time.Millisecond
		backend := httptest.NewServer(http.HandlerFunc(
			func(w http.ResponseWriter, _ *http.Request) {
				time.Sleep(timeout * 2)
				w.WriteHeader(http.StatusOK)
			}),
		)
		defer backend.Close()

		client := backend.Client()
		var logBuffer bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&logBuffer, nil))

		recorder := httptest.NewRecorder()
		req := makeRequest(t, "POST", defaultPath, map[string]string{"foo": "bar"})
		req.Host = backend.Listener.Addr().String()

		handler := tee.MakeProxyHandler(client, logger, timeout)

		// when
		handler.ServeHTTP(recorder, req)

		// then
		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
		assert.Contains(t, recorder.Body.String(), "forwarding request")
		assert.Contains(t, logBuffer.String(), "context deadline exceeded")
	})
}
