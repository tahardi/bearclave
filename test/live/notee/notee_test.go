package notee_test

//revive:disable:context-as-argument

import (
	"context"
	_ "embed"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tahardi/bearclave/tee"
)

//go:embed testdata/userdata.txt
var userdata []byte

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

type HTTPBinGetResponse struct {
	Args    map[string]string `json:"args"`
	Headers map[string]string `json:"headers"`
	Origin  string            `json:"origin"`
	URL     string            `json:"url"`
}

// TODO: Create a simpler example that only uses http. Then do https in another
func TestNoTEE_KitchenSink(t *testing.T) {
	// given
	ctx := context.Background()
	platform := tee.NoTEE
	attester, err := tee.NewAttester(platform)
	require.NoError(t, err)

	// given - make self-signed cert provider for our https enclave server
	domain := "bearclave.com"
	validity := 1 * time.Hour
	certProvider, err := tee.NewSelfSignedCertProvider(domain, validity)
	require.NoError(t, err)

	// given - Create an http enclave server that attests to the cert used by the
	// https enclave server
	serverMux := http.NewServeMux()
	serverMux.Handle(
		"POST "+AttestCertPath,
		MakeAttestCertHandler(t, attester, certProvider),
	)

	serverAddr := "http://127.0.0.1:8081"
	server, err := tee.NewServer(ctx, platform, serverAddr, serverMux)
	require.NoError(t, err)
	defer server.Close()

	// given - create a reverse proxy so we can call the http enclave server
	revProxyAddr := "http://127.0.0.1:8080"
	route := "app/v1"
	revProxy, err := tee.NewReverseProxy(
		ctx,
		platform,
		revProxyAddr,
		serverAddr,
		route,
	)
	require.NoError(t, err)
	defer revProxy.Close()

	// given - create a TLS proxy so the enclave can make HTTPS calls
	logger := slog.New(slog.DiscardHandler)
	proxyTLSAddr := "http://127.0.0.1:8082"
	proxyTLS, err := tee.NewProxyTLS(ctx, platform, proxyTLSAddr, logger)
	require.NoError(t, err)
	defer proxyTLS.Close()

	// given - configure a client that will route the enclave's HTTPS calls
	// through the TLS Proxy
	proxiedClient, err := tee.NewProxiedClient(platform, proxyTLSAddr)
	require.NoError(t, err)

	// given - create an https enclave server with a handler that makes HTTPS
	// calls on behalf of a client and attests to the output
	serverTLSMux := http.NewServeMux()
	serverTLSMux.Handle(
		"POST "+AttestHTTPSCallPath, MakeAttestHTTPSCallHandler(t, attester, proxiedClient),
	)

	serverTLSAddr := "https://127.0.0.1:8444"
	serverTLS, err := tee.NewServerTLS(
		ctx,
		platform,
		serverTLSAddr,
		serverTLSMux,
		certProvider,
	)
	require.NoError(t, err)
	defer serverTLS.Close()

	// given - create a reverse TLS proxy so we can make HTTPS calls to the
	// HTTPS enclave server
	revProxyTLSAddr := "https://127.0.0.1:8443"
	revProxyTLS, err := tee.NewReverseProxyTLS(
		ctx,
		platform,
		revProxyTLSAddr,
		serverTLSAddr,
	)
	require.NoError(t, err)
	defer revProxyTLS.Close()

	client := NewClient(t)

	// when - start proxies and servers
	runService(func() { _ = server.Serve() }, 100*time.Millisecond)
	runService(func() { _ = revProxy.Serve() }, 100*time.Millisecond)
	runService(func() { _ = serverTLS.Serve() }, 100*time.Millisecond)
	runService(func() { _ = revProxyTLS.Serve() }, 100*time.Millisecond)
	runService(func() { _ = proxyTLS.Serve() }, 100*time.Millisecond)

	// when - make an http call to get the attested cert chain
	nonce := []byte("first nonce")
	attested := client.AttestCertChain(ctx, revProxyAddr, nonce)

	// then - verify the attestation and add the cert chain to our client
	verifier, err := tee.NewVerifier(platform)
	require.NoError(t, err)

	verified, err := verifier.Verify(attested, tee.WithVerifyNonce(nonce))
	require.NoError(t, err)

	client.AddCertChain(verified.UserData)

	// when - make an https call with the https server's self-signed cert
	targetMethod := "GET"
	targetURL := "https://httpbin.org/get"
	attested = client.AttestHTTPSCall(ctx, revProxyTLSAddr, targetMethod, targetURL)

	// then - verify the attestation and user data
	verified, err = verifier.Verify(attested)
	require.NoError(t, err)

	httpBinResp := HTTPBinGetResponse{}
	err = json.Unmarshal(verified.UserData, &httpBinResp)
	require.NoError(t, err)
	assert.Equal(t, "https://httpbin.org/get", httpBinResp.URL)
}
