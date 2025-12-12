package networking

import (
	"context"
	"fmt"
	"net"
	"syscall"
	"time"

	"github.com/mdlayher/vsock"
)

const NumVSockAddrFields = 2

type ListenerOption func(*ListenerOptions)
type ListenerOptions struct {
	Control         func(network, address string, c syscall.RawConn) error
	KeepAlive       time.Duration
	KeepAliveConfig net.KeepAliveConfig
}

func WithListenerControl(
	control func(network, address string, c syscall.RawConn) error,
) ListenerOption {
	return func(opts *ListenerOptions) {
		opts.Control = control
	}
}

func WithListenerKeepAlive(keepAlive time.Duration) ListenerOption {
	return func(opts *ListenerOptions) {
		opts.KeepAlive = keepAlive
	}
}

func WithListenerKeepAliveConfig(config net.KeepAliveConfig) ListenerOption {
	return func(opts *ListenerOptions) {
		opts.KeepAliveConfig = config
	}
}

func NewSocketListener(
	ctx context.Context,
	network string,
	addr string,
	options ...ListenerOption,
) (net.Listener, error) {
	opts := ListenerOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	listenConfig := &net.ListenConfig{
		Control:         opts.Control,
		KeepAlive:       opts.KeepAlive,
		KeepAliveConfig: opts.KeepAliveConfig,
	}

	parsedAddr, err := ParseSocketAddr(addr)
	if err != nil {
		return nil, listenerError("", err)
	}

	// The context is only used while resolving the address. It does not
	// affect the returned Listener.
	listener, err := listenConfig.Listen(ctx, network, parsedAddr)
	if err != nil {
		msg := fmt.Sprintf(
			"creating %s socket listener on %s",
			network,
			parsedAddr,
		)
		return nil, listenerError(msg, err)
	}
	return listener, nil
}

func NewVSocketListener(
	ctx context.Context,
	_ string,
	addr string,
	_ ...ListenerOption,
) (net.Listener, error) {
	dataChan := make(chan net.Listener, 1)
	errChan := make(chan error, 1)
	go func() {
		cid, port, err := ParseVSocketAddr(addr)
		if err != nil {
			errChan <- listenerError("", err)
		}

		listener, err := vsock.ListenContextID(cid, port, nil)
		if err != nil {
			msg := fmt.Sprintf("creating vsocket listener on '%s'", addr)
			errChan <- listenerError(msg, err)
		}
		dataChan <- listener
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
