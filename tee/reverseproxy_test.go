package tee_test

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tahardi/bearclave"
	"github.com/tahardi/bearclave/tee"
)

func TestReverseProxy(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// given
		backend := httptest.NewServer(http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}),
		)
		defer backend.Close()

		platform := bearclave.NoTEE
		route := "/api/v1"
		proxy, err := tee.NewReverseProxy(platform, backend.URL, route)
		require.NoError(t, err)

		usersPath := route + "/users"
		targetURL := fmt.Sprintf("%s%s", backend.URL, usersPath)
		req := makeRequest(t, "GET", targetURL, nil)
		recorder := httptest.NewRecorder()

		// when
		proxy.ServeHTTP(recorder, req)

		// then
		assert.Equal(t, http.StatusOK, recorder.Code)
	})

	t.Run("happy path - verify path stripping", func(t *testing.T) {
		// given
		platform := bearclave.NoTEE
		backendURL := "http://127.0.0.1:8080"
		route := "/api/v1"
		proxy, err := tee.NewReverseProxy(platform, backendURL, route)
		require.NoError(t, err)

		usersPath := route + "/users"
		targetURL := fmt.Sprintf("%s%s", backendURL, usersPath)
		req := makeRequest(t, "GET", targetURL, nil)

		// when
		proxy.Director(req)

		// then
		assert.Equal(t, "/users", req.URL.Path)
		assert.Equal(t, "127.0.0.1:8080", req.URL.Host)
	})

	t.Run("error - dialing target", func(t *testing.T) {
		// given
		backendURL := "http://127.0.0.1:8080"
		route := "/api/v1"
		dialer := func(string, string) (net.Conn, error) {
			return nil, assert.AnError
		}
		proxy, err := tee.NewReverseProxyWithDialer(dialer, backendURL, route)
		require.NoError(t, err)

		usersPath := route + "/users"
		targetURL := fmt.Sprintf("%s%s", backendURL, usersPath)
		req := makeRequest(t, "GET", targetURL, nil)
		recorder := httptest.NewRecorder()

		// when
		proxy.ServeHTTP(recorder, req)

		// then
		assert.Equal(t, http.StatusBadGateway, recorder.Code)
	})
}
