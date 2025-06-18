package ipc

import (
	"context"
	"encoding/base64"
	"fmt"
	"net"
)

type SocketIPC struct {
	dial     func(network, address string) (net.Conn, error)
	listener net.Listener
}

func NewSocketIPC(endpoint string) (*SocketIPC, error) {
	listener, err := net.Listen("tcp", endpoint)
	if err != nil {
		return nil, fmt.Errorf(
			"initializing TCP listener on '%s': %w",
			endpoint,
			err,
		)
	}
	return NewSocketIPCWithDialAndListener(net.Dial, listener)
}

func NewSocketIPCWithDialAndListener(
	dial func(network, address string) (net.Conn, error),
	listener net.Listener,
) (*SocketIPC, error) {
	return &SocketIPC{
		dial:     dial,
		listener: listener,
	}, nil
}

func (s *SocketIPC) Close() error {
	return s.listener.Close()
}

func (s *SocketIPC) Send(ctx context.Context, endpoint string, data []byte) error {
	errChan := make(chan error, 1)
	go func() {
		conn, err := s.dial("tcp", endpoint)
		if err != nil {
			errChan <- fmt.Errorf("dialing '%s': %w", endpoint, err)
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

func (s *SocketIPC) Receive(ctx context.Context) ([]byte, error) {
	dataChan := make(chan []byte, 1)
	errChan := make(chan error, 1)
	go func() {
		conn, err := s.listener.Accept()
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
