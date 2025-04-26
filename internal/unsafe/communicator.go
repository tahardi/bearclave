package unsafe

import (
	"context"
	"encoding/base64"
	"fmt"
	"net"
)

type Communicator struct {
	receiveListener net.Listener
	sendAddr        string
}

func NewCommunicator(sendAddr string, receiveAddr string) (*Communicator, error) {
	receiveListener, err := net.Listen("tcp", receiveAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to set up TCP listener on %s: %w", receiveAddr, err)
	}

	return &Communicator{
		sendAddr:        sendAddr,
		receiveListener: receiveListener,
	}, nil
}

func (c *Communicator) Send(_ context.Context, data []byte) error {
	conn, err := net.Dial("tcp", c.sendAddr)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", c.sendAddr, err)
	}
	defer conn.Close()

	base64Data := base64.StdEncoding.EncodeToString(data)
	n, writeErr := conn.Write([]byte(base64Data))
	if n != len(base64Data) {
		return fmt.Errorf("failed to write all data: %w", writeErr)
	}
	return nil
}

func (c *Communicator) Receive(_ context.Context) ([]byte, error) {
	conn, err := c.receiveListener.Accept()
	if err != nil {
		return nil, fmt.Errorf("failed to accept connection: %w", err)
	}
	defer conn.Close()

	buf := make([]byte, 10000)
	n, readErr := conn.Read(buf)
	if readErr != nil {
		return nil, fmt.Errorf("failed to read data: %w", readErr)
	}

	base64Data := buf[:n]
	data, err := base64.StdEncoding.DecodeString(string(base64Data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode data: %w", err)
	}
	return data, nil
}

func (c *Communicator) Close() error {
	if c.receiveListener != nil {
		c.receiveListener.Close()
	}
	return nil
}
