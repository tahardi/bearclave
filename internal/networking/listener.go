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

func WithListenControl(
	control func(network, address string, c syscall.RawConn) error,
) ListenerOption {
	return func(opts *ListenerOptions) {
		opts.Control = control
	}
}

func WithListenKeepAlive(keepAlive time.Duration) ListenerOption {
	return func(opts *ListenerOptions) {
		opts.KeepAlive = keepAlive
	}
}

func WithListenKeepAliveConfig(config net.KeepAliveConfig) ListenerOption {
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

	// Wrap user's Control function with dual-stack control
	userControl := opts.Control
	opts.Control = func(network, address string, c syscall.RawConn) error {
		// Always apply dual-stack control first
		if err := dualStackControl(network, address, c); err != nil {
			return err
		}
		// Then apply user's custom control if provided
		if userControl != nil {
			return userControl(network, address, c)
		}
		return nil
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

// dualStackControl sets IPV6_V6ONLY to 0 to enable dual-stack (IPv4 + IPv6)
func dualStackControl(network, address string, c syscall.RawConn) error {
	return c.Control(func(fd uintptr) {
		// Only apply to IPv6 sockets
		if network == "tcp" || network == "tcp6" {
			_ = syscall.SetsockoptInt(int(fd), syscall.IPPROTO_IPV6, syscall.IPV6_V6ONLY, 0)
		}
	})
}
