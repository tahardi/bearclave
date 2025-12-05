package networking

import "github.com/tahardi/bearclave/internal/networking"

type Client = networking.Client

func NewClient(host string) *Client {
	return networking.NewClient(host)
}
