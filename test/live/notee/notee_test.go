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
var userData []byte

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

func TestNoTEE_Attestation(t *testing.T) {
	// given
	wantData := userData
	wantNonce := []byte("nonce")
	platform := tee.NoTEE

	attester, err := tee.NewAttester(platform)
	require.NoError(t, err)

	verifier, err := tee.NewVerifier(platform)
	require.NoError(t, err)

	// when
	attested, err := attester.Attest(
		tee.WithAttestNonce(wantNonce),
		tee.WithAttestUserData(wantData),
	)

	// then
	require.NoError(t, err)
	verified, err := verifier.Verify(attested, tee.WithVerifyNonce(wantNonce))
	require.NoError(t, err)
	assert.Equal(t, wantData, verified.UserData)
}

func TestNoTEE_Socket(t *testing.T) {
	// given
	ctx := context.Background()
	platform := tee.NoTEE

	service1Addr := "127.0.0.1:8080"
	service1, err := tee.NewSocket(ctx, platform, service1Addr)
	require.NoError(t, err)
	defer service1.Close()

	service2Addr := "127.0.0.1:8081"
	service2, err := tee.NewSocket(ctx, platform, service2Addr)
	require.NoError(t, err)
	defer service2.Close()

	want := []byte("hello from service 1")
	service2Func := func() {
		got, err := service2.Receive(ctx)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	}

	// when/then
	runService(service2Func, 100*time.Millisecond)
	err = service1.Send(ctx, service2Addr, want)
	require.NoError(t, err)
}

func TestNoTEE_HTTP(t *testing.T) {
	// given
	ctx := context.Background()
	platform := tee.NoTEE
	logger := slog.New(slog.DiscardHandler)

	// given - a proxy that forwards the server's HTTP requests
	proxyClient := &http.Client{}
	proxyAddr := "http://127.0.0.1:8082"
	proxy, err := tee.NewProxy(ctx, platform, proxyAddr, proxyClient, logger)
	require.NoError(t, err)
	defer proxy.Close()

	attester, err := tee.NewAttester(platform)
	require.NoError(t, err)

	// given - a client that routes the server's HTTP requests through the proxy
	proxiedClient, err := tee.NewProxiedClient(platform, proxyAddr)
	require.NoError(t, err)

	serverMux := http.NewServeMux()
	serverMux.HandleFunc(
		AttestHTTPCallPath,
		MakeAttestHTTPCallHandler(t, attester, proxiedClient),
	)

	// given - a server that makes HTTP requests on behalf of some remote client
	serverRoute := "app/v1"
	serverAddr := "http://127.0.0.1:8081"
	server, err := tee.NewServer(ctx, platform, serverAddr, serverMux)
	require.NoError(t, err)
	defer server.Close()

	// given - a reverse proxy that forwards a remote client's requests to the server
	targetAddr := serverAddr
	revProxyAddr := "http://127.0.0.1:8080"
	revProxy, err := tee.NewReverseProxy(
		ctx,
		platform,
		revProxyAddr,
		targetAddr,
		serverRoute,
	)
	require.NoError(t, err)
	defer revProxy.Close()

	// given - a remote client that wants the server to make an attested HTTP call
	wantMethod := "GET"
	wantURL := "http://httpbin.org/get"
	client := NewClient(t)

	// when
	runService(func() { _ = proxy.Serve() }, 100*time.Millisecond)
	runService(func() { _ = server.Serve() }, 100*time.Millisecond)
	runService(func() { _ = revProxy.Serve() }, 100*time.Millisecond)

	attested := client.AttestHTTPCall(ctx, revProxyAddr, wantMethod, wantURL)

	// then
	verifier, err := tee.NewVerifier(platform)
	require.NoError(t, err)

	verified, err := verifier.Verify(attested)
	require.NoError(t, err)

	httpBinResp := HTTPBinGetResponse{}
	err = json.Unmarshal(verified.UserData, &httpBinResp)
	require.NoError(t, err)
	assert.Equal(t, wantURL, httpBinResp.URL)
}

// This test demonstrates HTTPS in and out of a TEE. In this test scenario, a
// client wants our TEE server to make an attested HTTPS call on their behalf.
// The client also wishes to use HTTPS when communicating with our TEE server.
//
// To use HTTPS, the client must first retrieve our TEE server's certificate.
// This is done by making an HTTP call to our HTTP server, which returns an
// attestation containing the HTTPS server's certificate. Note that we have
// a separate reverse proxy and HTTP server for this workflow.
//
// After verifying the attestation and extracting the certificate, the client
// can now establish an HTTPS connection to our HTTPS TEE server via a
// reverse TLS proxy.
//
// Upon receiving a request, our HTTPS server makes an HTTPS call via a proxy
// TLS server. It attests to the result and returns it to the client.
//
// The test topology looks like this:
//
//	---- revProxy ------- server
//	|
//
// client -
//
//	|
//	---- revProxyTLS ---- serverTLS ---- proxyTLS ---- httpbin
func TestNoTEE_HTTPS(t *testing.T) {
	// given
	ctx := context.Background()
	platform := tee.NoTEE
	logger := slog.New(slog.DiscardHandler)

	attester, err := tee.NewAttester(platform)
	require.NoError(t, err)

	// given - a self-signed TLS certificate provider for our HTTPS server
	domain := "bearclave.org"
	validity := 1 * time.Hour
	certProvider, err := tee.NewSelfSignedCertProvider(domain, validity)
	require.NoError(t, err)

	// given - a server that attests to the HTTPS server's certificate
	serverMux := http.NewServeMux()
	serverMux.HandleFunc(
		AttestCertPath,
		MakeAttestCertHandler(t, attester, certProvider),
	)

	serverRoute := "app/v1"
	serverAddr := "http://127.0.0.1:8081"
	server, err := tee.NewServer(ctx, platform, serverAddr, serverMux)
	require.NoError(t, err)
	defer server.Close()

	// given - a reverse proxy that forwards a remote client's requests to the server
	targetAddr := serverAddr
	revProxyAddr := "http://127.0.0.1:8080"
	revProxy, err := tee.NewReverseProxy(
		ctx,
		platform,
		revProxyAddr,
		targetAddr,
		serverRoute,
	)
	require.NoError(t, err)
	defer revProxy.Close()

	// given - a proxy that forwards the HTTPS server's requests
	proxyTLSAddr := "http://127.0.0.1:8082"
	proxyTLS, err := tee.NewProxyTLS(ctx, platform, proxyTLSAddr, logger)
	require.NoError(t, err)
	defer proxyTLS.Close()

	// given - a client that routes the HTTPS server's requests through the proxy
	proxiedClient, err := tee.NewProxiedClient(platform, proxyTLSAddr)
	require.NoError(t, err)

	serverTLSMux := http.NewServeMux()
	serverTLSMux.HandleFunc(
		AttestHTTPSCallPath,
		MakeAttestHTTPSCallHandler(t, attester, proxiedClient),
	)

	// given - a server that makes HTTPS requests on behalf of some remote client
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

	// given - a reverse proxy that forwards a remote client's requests to the HTTPS server
	targetTLSAddr := serverTLSAddr
	revProxyTLSAddr := "https://127.0.0.1:8443"
	revProxyTLS, err := tee.NewReverseProxyTLS(
		ctx,
		platform,
		revProxyTLSAddr,
		targetTLSAddr,
	)
	require.NoError(t, err)
	defer revProxyTLS.Close()

	// given - a remote client that wants the server to make an attested HTTP call
	wantMethod := "GET"
	wantURL := "https://httpbin.org/get"
	client := NewClient(t)

	// when
	runService(func() { _ = server.Serve() }, 100*time.Millisecond)
	runService(func() { _ = revProxy.Serve() }, 100*time.Millisecond)
	runService(func() { _ = proxyTLS.Serve() }, 100*time.Millisecond)
	runService(func() { _ = serverTLS.Serve() }, 100*time.Millisecond)
	runService(func() { _ = revProxyTLS.Serve() }, 100*time.Millisecond)

	attestedCert := client.AttestCertChain(ctx, revProxyAddr)
	verifier, err := tee.NewVerifier(platform)
	require.NoError(t, err)
	verifiedCert, err := verifier.Verify(attestedCert)
	require.NoError(t, err)

	client.AddCertChain(verifiedCert.UserData)
	attestedCall := client.AttestHTTPSCall(ctx, revProxyTLSAddr, wantMethod, wantURL)

	// then
	verifiedCall, err := verifier.Verify(attestedCall)
	require.NoError(t, err)

	httpBinResp := HTTPBinGetResponse{}
	err = json.Unmarshal(verifiedCall.UserData, &httpBinResp)
	require.NoError(t, err)
	assert.Equal(t, wantURL, httpBinResp.URL)
}
