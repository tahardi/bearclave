package tee_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tahardi/bearclave/tee"
)

func TestReverseProxy(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// given
		backend := httptest.NewServer(http.HandlerFunc(
			func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			}),
		)
		defer backend.Close()

		ctx := context.Background()
		platform := tee.NoTEE
		network := "tcp"
		proxyAddr := "http://127.0.0.1:8080"
		targetAddr := backend.URL
		route := "/api/v1"
		proxy, err := tee.NewReverseProxy(
			ctx, platform, network, proxyAddr, targetAddr, route,
		)
		require.NoError(t, err)
		defer proxy.Close()

		usersPath := route + "/users"
		targetURL := fmt.Sprintf("%s%s", backend.URL, usersPath)
		req := makeRequest(t, "GET", targetURL, nil)
		recorder := httptest.NewRecorder()

		// when
		proxy.Handler().ServeHTTP(recorder, req)

		// then
		assert.Equal(t, http.StatusOK, recorder.Code)
	})

	t.Run("error - dialing target", func(t *testing.T) {
		// given
		ctx := context.Background()
		platform := tee.NoTEE
		network := "tcp"
		proxyAddr := "http://127.0.0.1:8080"
		targetAddr := "http://127.0.0.1:8081"
		route := "/api/v1"
		dialContext := func(context.Context, string, string) (net.Conn, error) {
			return nil, assert.AnError
		}
		proxy, err := tee.NewReverseProxyWithDialContext(
			ctx, platform, dialContext, network, proxyAddr, targetAddr, route,
		)
		require.NoError(t, err)
		defer proxy.Close()

		usersPath := route + "/users"
		targetURL := fmt.Sprintf("%s%s", targetAddr, usersPath)
		req := makeRequest(t, "GET", targetURL, nil)
		recorder := httptest.NewRecorder()

		// when
		proxy.Handler().ServeHTTP(recorder, req)

		// then
		assert.Equal(t, http.StatusBadGateway, recorder.Code)
	})
}
