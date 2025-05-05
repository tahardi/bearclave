package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/tahardi/bearclave"
	"github.com/tahardi/bearclave/examples/hello-world/sdk"
	"io"
	"log/slog"
	"net/http"
	"os"
)

func MakeVerifier(config *sdk.Config) (bearclave.Verifier, error) {
	switch config.Platform {
	case sdk.Nitro:
		return bearclave.NewNitroVerifier()
	case sdk.SEV:
		return bearclave.NewSEVVerifier()
	case sdk.TDX:
		return bearclave.NewTDXVerifier()
	case sdk.Unsafe:
		return bearclave.NewUnsafeVerifier()
	default:
		return nil, fmt.Errorf("unsupported platform '%s'", config.Platform)
	}
}

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

var configFile string

func main() {
	flag.StringVar(
		&configFile,
		"config",
		sdk.DefaultConfigFile,
		"The Trusted Computing platform to use. Options: "+
			"nitro, sev, tdx, unsafe (default: unsafe)",
	)
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config, err := sdk.LoadConfig(configFile)
	if err != nil {
		logger.Error("loading config", slog.String("error", err.Error()))
		return
	}
	logger.Info("loaded config", slog.Any(configFile, config))

	verifier, err := MakeVerifier(config)
	if err != nil {
		logger.Error("making verifier", slog.String("error", err.Error()))
		return
	}

	want := []byte("Hello, world!")
	client := NewGatewayClient("http://localhost:8080")
	attestation, err := client.AttestUserData(want)
	if err != nil {
		logger.Error("attesting userdata", slog.String("error", err.Error()))
	}

	got, err := verifier.Verify(attestation)
	if err != nil {
		logger.Error("verifying attestation", slog.String("error", err.Error()))
		return
	}

	if !bytes.Equal(got, want) {
		logger.Error("userdata verification failed")
		return
	}
	logger.Info("verified userdata", slog.String("userdata", string(got)))
}
