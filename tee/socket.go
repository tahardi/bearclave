package tee

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"time"
)

const DefaultConnTimeout = 5 * time.Second

type Socket struct {
	platform Platform
	network  string
	listener net.Listener
}

func NewSocket(
	ctx context.Context,
	platform Platform,
	network string,
	addr string,
) (*Socket, error) {
	listener, err := NewListener(ctx, platform, network, addr)
	if err != nil {
		return nil, socketError("creating listener", err)
	}
	return &Socket{
		platform: platform,
		network:  network,
		listener: listener,
	}, nil
}

func NewSocketWithListener(
	platform Platform,
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
	dialer, err := NewDialContext(s.platform)
	if err != nil {
		return socketError("creating dialer", err)
	}
	return s.SendWithDialContext(ctx, DefaultConnTimeout, dialer, addr, data)
}

func (s *Socket) SendWithDialContext(
	ctx context.Context,
	connTimeout time.Duration,
	dialContext DialContext,
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
			msg := fmt.Sprintf("dialing '%s'", addr)
			errChan <- socketError(msg, err)
			return
		}
		defer conn.Close()

		base64Data := base64.StdEncoding.EncodeToString(data)

		n, writeErr := conn.Write([]byte(base64Data))
		switch {
		case writeErr != nil:
			errChan <- socketError("writing data", writeErr)
		case n != len(base64Data):
			errChan <- socketError("failed to write all data", writeErr)
		default:
			errChan <- nil
		}
	}()

	select {
	case <-ctx.Done():
		return socketError("deadline exceeded or context cancelled", ctx.Err())
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
			errChan <- socketError("accepting connection", err)
			return
		}
		defer conn.Close()

		var buf bytes.Buffer
		_, err = io.Copy(&buf, conn)
		if err != nil {
			errChan <- socketError("reading data", err)
			return
		}

		data, err := base64.StdEncoding.DecodeString(buf.String())
		if err != nil {
			errChan <- socketError("decoding data", err)
			return
		}
		dataChan <- data
	}()

	select {
	case <-ctx.Done():
		return nil, socketError("deadline exceeded or context cancelled", ctx.Err())
	case err := <-errChan:
		return nil, socketError("receiving data", err)
	case data := <-dataChan:
		return data, nil
	}
}
