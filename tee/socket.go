package tee

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/tahardi/bearclave"
)

const DefaultConnTimeout = 5 * time.Second

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

func (s *Socket) Send(
	ctx context.Context,
	addr string,
	data []byte,
) error {
	dialer, err := bearclave.NewDialContext(s.platform)
	if err != nil {
		return fmt.Errorf("creating dialer: %w", err)
	}
	return s.SendWithDialContext(ctx, DefaultConnTimeout, dialer, addr, data)
}

func (s *Socket) SendWithDialContext(
	ctx context.Context,
	connTimeout time.Duration,
	dialContext bearclave.DialContext,
	addr string,
	data []byte,
) error {
	errChan := make(chan error, 1)
	go func() {
		// This context is used to establish the connection. Once the connection
		// is established, this context no longer has any effect. The context
		// passed to Send is used for the actual data transfer.
		connCtx, cancel := context.WithTimeout(ctx, connTimeout)
		defer cancel()

		conn, err := dialContext(connCtx, s.network, addr)
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

		var buf bytes.Buffer
		_, err = io.Copy(&buf, conn)
		if err != nil {
			errChan <- fmt.Errorf("reading data: %w", err)
			return
		}

		data, err := base64.StdEncoding.DecodeString(buf.String())
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
