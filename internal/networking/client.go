package networking

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	host   string
	client *http.Client
}

func NewClient(host string) *Client {
	client := &http.Client{}
	return NewClientWithClient(host, client)
}

func NewClientWithClient(
	host string,
	client *http.Client,
) *Client {
	return &Client{
		host:   host,
		client: client,
	}
}

func (c *Client) AttestUserData(data []byte) ([]byte, error) {
	attestUserDataRequest := AttestUserDataRequest{Data: data}
	attestUserDataResponse := AttestUserDataResponse{}
	err := c.Do(
		"POST",
		AttestUserDataPath,
		attestUserDataRequest,
		&attestUserDataResponse,
	)
	if err != nil {
		return nil, fmt.Errorf("doing attest userdata request: %w", err)
	}
	return attestUserDataResponse.Attestation, nil
}

func (c *Client) Do(
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
