package tee_test

import (
	"context"
	"crypto/tls"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tahardi/bearclave/tee"
)

func runService(service func(), wait time.Duration) {
	var serviceReady sync.WaitGroup
	serviceReady.Add(1)
	go func() {
		serviceReady.Done()
		service()
	}()
	serviceReady.Wait()
	time.Sleep(wait)
}

func TestReverseProxy(t *testing.T) {
	ctx := context.Background()
	platform := tee.NoTEE
	revProxyAddr := "http://127.0.0.1:8080"
	logger := slog.New(slog.DiscardHandler)

	t.Run("happy path", func(t *testing.T) {
		// given
		backend := httptest.NewServer(http.HandlerFunc(
			func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			}),
		)
		defer backend.Close()

		targetAddr := backend.URL
		revProxy, err := tee.NewReverseProxy(ctx, platform, revProxyAddr, targetAddr, logger)
		require.NoError(t, err)
		defer revProxy.Close()

		req := makeRequest(t, "GET", revProxyAddr, nil)
		client := &http.Client{}

		// when
		runService(func() { _ = revProxy.Serve() }, 100*time.Millisecond)
		resp, err := client.Do(req)

		// then
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("error - dialing target", func(t *testing.T) {
		// given
		targetAddr := "http://127.0.0.1:9999"
		dialContext := func(context.Context, string, string) (net.Conn, error) {
			return nil, assert.AnError
		}
		revProxy, err := tee.NewReverseProxyWithDialContext(
			ctx, dialContext, revProxyAddr, targetAddr, logger,
		)
		require.NoError(t, err)
		defer revProxy.Close()

		req := makeRequest(t, "GET", revProxyAddr, nil)
		client := &http.Client{}

		// when
		runService(func() { _ = revProxy.Serve() }, 100*time.Millisecond)
		resp, err := client.Do(req)

		// then
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadGateway, resp.StatusCode)
	})
}

func TestReverseProxyTLS(t *testing.T) {
	ctx := context.Background()
	platform := tee.NoTEE
	revProxyAddr := "https://127.0.0.1:8443"
	logger := slog.New(slog.DiscardHandler)

	t.Run("happy path - TLS connection forwarded", func(t *testing.T) {
		// given
		backend := httptest.NewTLSServer(http.HandlerFunc(
			func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			}),
		)
		defer backend.Close()

		targetAddr := backend.URL
		revProxyTLS, err := tee.NewReverseProxyTLS(
			ctx,
			platform,
			revProxyAddr,
			targetAddr,
			logger,
		)
		require.NoError(t, err)
		defer revProxyTLS.Close()

		client := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}

		// when
		runService(func() { _ = revProxyTLS.Serve() }, 100*time.Millisecond)
		resp, err := client.Get(revProxyAddr)

		// then
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("error - dialing target fails", func(t *testing.T) {
		// given
		targetAddr := "127.0.0.1:9999" // Non-existent backend
		dialContext := func(context.Context, string, string) (net.Conn, error) {
			return nil, assert.AnError
		}

		revProxyTLS, err := tee.NewReverseProxyTLSWithDialContext(
			ctx,
			dialContext,
			revProxyAddr,
			targetAddr,
			logger,
		)
		require.NoError(t, err)
		defer revProxyTLS.Close()

		client := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}

		// when
		runService(func() { _ = revProxyTLS.Serve() }, 100*time.Millisecond)
		_, err = client.Get(revProxyAddr)

		// then
		require.ErrorIs(t, err, io.EOF)
	})
}
