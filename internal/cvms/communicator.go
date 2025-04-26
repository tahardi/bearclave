package cvms

import (
	"context"
	"fmt"
)

type Communicator struct{}

func NewCommunicator() (*Communicator, error) {
	return &Communicator{}, nil
}

func (c *Communicator) Close() error {
	return fmt.Errorf("not implemented")
}

func (c *Communicator) Send(ctx context.Context, data []byte) error {
	return fmt.Errorf("not implemented")
}

func (c *Communicator) Receive(ctx context.Context) ([]byte, error) {
	return nil, fmt.Errorf("not implemented")
}
