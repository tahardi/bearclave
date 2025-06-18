package ipc

import (
	"context"
	"encoding/base64"
	"fmt"
	"net"

	"github.com/mdlayher/vsock"
)

type VSocketIPC struct {
	dial     func(contextID, port uint32, cfg *vsock.Config) (net.Conn, error)
	listener net.Listener
}

func NewVSocketIPC(endpoint string) (*VSocketIPC, error) {
	cid, port, err := ParseEndpoint(endpoint)
	if err != nil {
		return nil, fmt.Errorf("parsing endpoint: %w", err)
	}

	listener, err := vsock.ListenContextID(cid, port, nil)
	if err != nil {
		return nil, fmt.Errorf(
			"initializing vsock listener on '%s': %w",
			endpoint,
			err,
		)
	}

	dial := func(contextID, port uint32, cfg *vsock.Config) (net.Conn, error) {
		return vsock.Dial(contextID, port, cfg)
	}
	return NewVSocketIPCWithDialAndListener(dial, listener)
}

func NewVSocketIPCWithDialAndListener(
	dial func(contextID, port uint32, cfg *vsock.Config) (net.Conn, error),
	listener net.Listener,
) (*VSocketIPC, error) {
	return &VSocketIPC{
		dial:     dial,
		listener: listener,
	}, nil
}

func (v *VSocketIPC) Close() error {
	return v.listener.Close()
}

func (v *VSocketIPC) Send(ctx context.Context, endpoint string, data []byte) error {
	cid, port, err := ParseEndpoint(endpoint)
	if err != nil {
		return fmt.Errorf("parsing endpoint: %w", err)
	}

	errChan := make(chan error, 1)
	go func() {
		conn, err := v.dial(cid, port, nil)
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

func (v *VSocketIPC) Receive(ctx context.Context) ([]byte, error) {
	dataChan := make(chan []byte, 1)
	errChan := make(chan error, 1)
	go func() {
		conn, err := v.listener.Accept()
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

func ParseEndpoint(endpoint string) (uint32, uint32, error) {
	var cid, port int
	n, err := fmt.Sscanf(endpoint, "%d:%d", &cid, &port)
	switch {
	case err != nil:
		return 0, 0, fmt.Errorf("invalid format '%s': %w", endpoint, err)
	case n != 2:
		return 0, 0, fmt.Errorf("expected '2' got '%d", n)
	}
	return uint32(cid), uint32(port), nil
}
