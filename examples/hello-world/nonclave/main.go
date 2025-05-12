package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/tahardi/bearclave/examples/hello-world/sdk"
)

type GatewayClient struct {
	host   string
	client *http.Client
}

func NewGatewayClient(host string) *GatewayClient {
	// TODO: Check configuration - should I set a timeout?
	client := &http.Client{}
	return NewGatewayClientWithClient(host, client)
}

func NewGatewayClientWithClient(
	host string,
	client *http.Client,
) *GatewayClient {
	return &GatewayClient{
		host:   host,
		client: client,
	}
}

type AttestUserDataRequest struct {
	Data []byte `json:"data"`
}
type AttestUserDataResponse struct {
	Attestation []byte `json:"attestation"`
}

func (c *GatewayClient) AttestUserData(data []byte) ([]byte, error) {
	attestUserDataRequest := AttestUserDataRequest{Data: data}
	attestUserDataResponse := AttestUserDataResponse{}
	err := c.Do("POST", "/attest-user-data", attestUserDataRequest, &attestUserDataResponse)
	if err != nil {
		return nil, fmt.Errorf("doing attest user data request: %w", err)
	}
	return attestUserDataResponse.Attestation, nil
}

func (c *GatewayClient) Do(
	method string,
	api string,
	apiReq any,
	apiResp any,
) error {
	bodyBytes, err := json.Marshal(apiReq)
	if err != nil {
		return fmt.Errorf("marshaling request body: %w", err)
	}

	url := c.host + api
	req, err := http.NewRequest(method, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err = io.ReadAll(resp.Body)
	switch {
	case err != nil:
		return fmt.Errorf("reading response body: %w", err)
	case resp.StatusCode != http.StatusOK:
		return fmt.Errorf("received non-200 response: %s", string(bodyBytes))
	}

	err = json.Unmarshal(bodyBytes, apiResp)
	if err != nil {
		return fmt.Errorf("unmarshaling response: %w", err)
	}
	return nil
}

var host string
var port int
var platform string

func main() {
	flag.StringVar(
		&host,
		"host",
		"127.0.0.1",
		"The hostname of the enclave gateway to connect to (default: 127.0.0.1)",
	)
	flag.IntVar(
		&port,
		"port",
		8080,
		"The port of the enclave gateway to connect to (default: 8080)",
	)
	flag.StringVar(
		&platform,
		"platform",
		"unsafe",
		"The Trusted Computing platform the enclave is running on. Options: "+
			"nitro, sev, tdx, unsafe (default: unsafe)",
	)
	flag.Parse()

	url := fmt.Sprintf("http://%s:%d", host, port)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	logger.Info("nonclave configuration",
		slog.String("platform", platform),
		slog.String("url", url),
	)

	verifier, err := sdk.MakeVerifier(sdk.Platform(platform))
	if err != nil {
		logger.Error("making verifier", slog.String("error", err.Error()))
		return
	}

	want := []byte("Hello, world!")
	client := NewGatewayClient(url)
	attestation, err := client.AttestUserData(want)
	if err != nil {
		logger.Error("attesting userdata", slog.String("error", err.Error()))
		return
	}

	got, err := verifier.Verify(attestation)
	if err != nil {
		logger.Error("verifying attestation", slog.String("error", err.Error()))
		return
	}

	if !bytes.Contains(got, want) {
		logger.Error("userdata verification failed")
		return
	}
	logger.Info("verified userdata", slog.String("userdata", string(got)))
}
