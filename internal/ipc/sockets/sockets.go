package sockets

import (
	"context"
	"encoding/base64"
	"fmt"
	"net"
)

type IPC struct {
	receiveListener net.Listener
}

func NewIPC(port int) (*IPC, error) {
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	receiveListener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf(
			"initializing TCP listener on '%s': %w",
			addr,
			err,
		)
	}

	return &IPC{
		receiveListener: receiveListener,
	}, nil
}

func (i *IPC) Close() error {
	if i.receiveListener != nil {
		i.receiveListener.Close()
	}
	return nil
}

func (i *IPC) Send(ctx context.Context, _ int, port int, data []byte) error {
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	errChan := make(chan error, 1)
	go func() {
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			errChan <- fmt.Errorf("dialing '%s': %w", addr, err)
			return
		}
		defer conn.Close()

		base64Data := base64.StdEncoding.EncodeToString(data)

		n, writeErr := conn.Write([]byte(base64Data))
		switch {
		case writeErr != nil:
			errChan <- fmt.Errorf("writing data: %w", writeErr)
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

func (i *IPC) Receive(ctx context.Context) ([]byte, error) {
	dataChan := make(chan []byte, 1)
	errChan := make(chan error, 1)
	go func() {
		conn, err := i.receiveListener.Accept()
		if err != nil {
			errChan <- fmt.Errorf("accepting connection: %w", err)
			return
		}
		defer conn.Close()

		buf := make([]byte, 10000)
		n, readErr := conn.Read(buf)
		if readErr != nil {
			errChan <- fmt.Errorf("reading data: %w", readErr)
			return
		}

		base64Data := buf[:n]
		data, err := base64.StdEncoding.DecodeString(string(base64Data))
		if err != nil {
			errChan <- fmt.Errorf("decoding data: %w", err)
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
