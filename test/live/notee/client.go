package notee_test

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tahardi/bearclave/tee"
)

type Client struct {
	t      *testing.T
	client *http.Client
}

func NewClient(t *testing.T) *Client {
	t.Helper()
	return &Client{t: t, client: &http.Client{}}
}

func (c *Client) AddCertChain(certChainJSON []byte) {
	c.t.Helper()
	chainDER := [][]byte{}
	err := json.Unmarshal(certChainJSON, &chainDER)
	require.NoError(c.t, err)

	if c.client.Transport == nil {
		c.client.Transport = &http.Transport{}
	}
	transport, ok := c.client.Transport.(*http.Transport)
	require.True(c.t, ok)

	//nolint:gosec
	if transport.TLSClientConfig == nil {
		transport.TLSClientConfig = &tls.Config{}
	}
	if transport.TLSClientConfig.RootCAs == nil {
		transport.TLSClientConfig.RootCAs = x509.NewCertPool()
	}

	for i, certBytes := range chainDER {
		x509Cert, err := x509.ParseCertificate(certBytes)
		require.NoError(c.t, err, "failed to parse cert #%d", i)
		transport.TLSClientConfig.RootCAs.AddCert(x509Cert)
	}
}

func (c *Client) AttestCertChain(
	ctx context.Context,
	host string,
) *tee.AttestResult {
	c.t.Helper()
	attReq := AttestCertRequest{}
	attRes := AttestCertResponse{}
	c.DoRequest(
		ctx,
		host,
		"POST",
		AttestCertPath,
		attReq,
		&attRes,
	)
	return attRes.Attestation
}

func (c *Client) AttestHTTPCall(
	ctx context.Context,
	host string,
	method string,
	url string,
) *tee.AttestResult {
	c.t.Helper()
	attReq := AttestHTTPCallRequest{Method: method, URL: url}
	attRes := AttestHTTPCallResponse{}
	c.DoRequest(
		ctx,
		host,
		"POST",
		AttestHTTPCallPath,
		attReq,
		&attRes,
	)
	return attRes.Attestation
}

func (c *Client) AttestHTTPSCall(
	ctx context.Context,
	host string,
	method string,
	url string,
) *tee.AttestResult {
	c.t.Helper()
	attReq := AttestHTTPSCallRequest{Method: method, URL: url}
	attRes := AttestHTTPSCallResponse{}
	c.DoRequest(
		ctx,
		host,
		"POST",
		AttestHTTPSCallPath,
		attReq,
		&attRes,
	)
	return attRes.Attestation
}

func (c *Client) DoRequest(
	ctx context.Context,
	host string,
	method string,
	api string,
	apiReq any,
	apiResp any,
) {
	c.t.Helper()
	bodyBytes, err := json.Marshal(apiReq)
	require.NoError(c.t, err)

	url := host + api
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(bodyBytes))
	require.NoError(c.t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	require.NoError(c.t, err)
	require.Equal(c.t, http.StatusOK, resp.StatusCode)
	defer resp.Body.Close()

	bodyBytes, err = io.ReadAll(resp.Body)
	require.NoError(c.t, err)

	err = json.Unmarshal(bodyBytes, apiResp)
	require.NoError(c.t, err)
}
