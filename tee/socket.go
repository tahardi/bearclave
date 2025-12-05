package tee

import (
	"context"
	"encoding/base64"
	"fmt"
	"net"

	"github.com/tahardi/bearclave"
)

type Socket struct {
	platform bearclave.Platform
	network  string
	listener net.Listener
}

func NewSocket(
	platform bearclave.Platform,
	network string,
	addr string,
) (*Socket, error) {
	listener, err := bearclave.NewListener(platform, network, addr)
	if err != nil {
		return nil, err
	}
	return &Socket{
		platform: platform,
		network:  network,
		listener: listener,
	}, nil
}

func NewSocketWithListener(
	platform bearclave.Platform,
	network string,
	listener net.Listener,
) (*Socket, error) {
	return &Socket{
		platform: platform,
		network:  network,
		listener: listener,
	}, nil
}

func (s *Socket) Close() error {
	return s.listener.Close()
}

func (s *Socket) Send(ctx context.Context, addr string, data []byte) error {
	dialer, err := bearclave.NewDialer(s.platform)
	if err != nil {
		return fmt.Errorf("creating dialer: %w", err)
	}
	return s.SendWithDialer(ctx, dialer, addr, data)
}

func (s *Socket) SendWithDialer(
	ctx context.Context,
	dialer bearclave.Dialer,
	addr string,
	data []byte,
) error {
	errChan := make(chan error, 1)
	go func() {
		conn, err := dialer(s.network, addr)
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
		return fmt.Errorf("deadline exceeded or context cancelled: %w", ctx.Err())
	case err := <-errChan:
		return err
	}
}

func (s *Socket) Receive(ctx context.Context) ([]byte, error) {
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
		return nil, fmt.Errorf("deadline exceeded or context cancelled: %w", ctx.Err())
	case err := <-errChan:
		return nil, err
	case data := <-dataChan:
		return data, nil
	}
}
