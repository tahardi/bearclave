package vsockets

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/mdlayher/vsock"
)

type IPC struct {
	sendContextID   uint32
	sendPort        uint32
	receiveListener *vsock.Listener
}

func NewIPC(
	sendContextID int,
	sendPort int,
	receivePort int,
) (*IPC, error) {
	receiveListener, err := vsock.Listen(uint32(receivePort), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to set up vsock listener: %w", err)
	}

	return &IPC{
		sendContextID:   uint32(sendContextID),
		sendPort:        uint32(sendPort),
		receiveListener: receiveListener,
	}, nil
}

func (c *IPC) Close() error {
	return fmt.Errorf("not implemented")
}

func (c *IPC) Send(ctx context.Context, data []byte) error {
	errChan := make(chan error, 1)
	go func() {
		conn, err := vsock.Dial(c.sendContextID, c.sendPort, nil)
		if err != nil {
			errChan <- fmt.Errorf("failed to connect to %v: %w", c.sendPort, err)
			return
		}
		defer conn.Close()

		base64Data := base64.StdEncoding.EncodeToString(data)

		n, writeErr := conn.Write([]byte(base64Data))
		switch {
		case writeErr != nil:
			errChan <- fmt.Errorf("failed to write data: %w", writeErr)
		case n != len(base64Data):
			errChan <- fmt.Errorf("failed to write all data: %w", writeErr)
		default:
			errChan <- nil
		}
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("send context cancelled: %w", ctx.Err())
	case err := <-errChan:
		return err
	}
}

func (c *IPC) Receive(ctx context.Context) ([]byte, error) {
	dataChan := make(chan []byte, 1)
	errChan := make(chan error, 1)
	go func() {
		conn, err := c.receiveListener.Accept()
		if err != nil {
			errChan <- fmt.Errorf("failed to accept connection: %w", err)
			return
		}
		defer conn.Close()

		buf := make([]byte, 10000)
		n, readErr := conn.Read(buf)
		if readErr != nil {
			errChan <- fmt.Errorf("failed to read data: %w", readErr)
			return
		}

		base64Data := buf[:n]
		data, err := base64.StdEncoding.DecodeString(string(base64Data))
		if err != nil {
			errChan <- fmt.Errorf("failed to decode data: %w", err)
			return
		}
		dataChan <- data
	}()

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("receive context cancelled: %w", ctx.Err())
	case err := <-errChan:
		return nil, err
	case data := <-dataChan:
		return data, nil
	}
}
