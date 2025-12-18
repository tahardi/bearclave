package networking

import (
	"net/http"
	"net/url"
)

func NewProxiedSocketClient(proxyAddr string) (*http.Client, error) {
	dialContext, err := NewSocketDialContext()
	if err != nil {
		return nil, err
	}
	return NewProxiedClient(dialContext, proxyAddr)
}

func NewProxiedVSocketClient(proxyAddr string) (*http.Client, error) {
	dialContext, err := NewVSocketDialContext()
	if err != nil {
		return nil, err
	}
	return NewProxiedClient(dialContext, proxyAddr)
}

func NewProxiedClient(
	dialContext DialContext,
	proxyAddr string,
) (*http.Client, error) {
	// Make the HTTP client send requests to the Proxy server instead of the
	// target server. The Proxy server should forward the request to the target.
	proxy := func(_ *http.Request) (*url.URL, error) { return url.Parse(proxyAddr) }
	transport := &http.Transport{
		DialContext: dialContext,
		Proxy:       proxy,
	}
	return &http.Client{Transport: transport}, nil
}
