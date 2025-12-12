package networking

import (
	"context"
	"fmt"
	"net"
	"syscall"
	"time"

	"github.com/mdlayher/vsock"
)

type DialContext func(ctx context.Context, network string, addr string) (net.Conn, error)

type DialerOption func(*DialerOptions)
type DialerOptions struct {
	Control   func(network, address string, c syscall.RawConn) error
	KeepAlive time.Duration
	LocalAddr net.Addr
	Timeout   time.Duration
}

func WithControlContext(
	control func(network, address string, c syscall.RawConn) error,
) DialerOption {
	return func(opts *DialerOptions) {
		opts.Control = control
	}
}

func WithKeepAlive(keepAlive time.Duration) DialerOption {
	return func(opts *DialerOptions) {
		opts.KeepAlive = keepAlive
	}
}

func WithLocalAddr(localAddr net.Addr) DialerOption {
	return func(opts *DialerOptions) {
		opts.LocalAddr = localAddr
	}
}

func WithTimeout(timeout time.Duration) DialerOption {
	return func(opts *DialerOptions) {
		opts.Timeout = timeout
	}
}

func NewSocketDialContext(options ...DialerOption) (DialContext, error) {
	opts := DialerOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	dialer := &net.Dialer{
		Control:   opts.Control,
		KeepAlive: opts.KeepAlive,
		LocalAddr: opts.LocalAddr,
		Timeout:   opts.Timeout,
	}

	return func(ctx context.Context, network string, addr string) (net.Conn, error) {
		parsedAddr, err := ParseSocketAddr(addr)
		if err != nil {
			return nil, dialContextError("", err)
		}
		return dialer.DialContext(ctx, network, parsedAddr)
	}, nil
}

func NewVSocketDialContext(_ ...DialerOption) (DialContext, error) {
	return func(ctx context.Context, _, addr string) (net.Conn, error) {
		cid, port, err := ParseVSocketAddr(addr)
		if err != nil {
			return nil, dialContextError("", err)
		}

		dataChan := make(chan net.Conn, 1)
		errChan := make(chan error, 1)
		go func() {
			conn, err := vsock.Dial(cid, port, nil)
			if err != nil {
				errChan <- fmt.Errorf("dialing '%s': %w", addr, err)
			}
			dataChan <- conn
		}()
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("deadline exceeded or context cancelled: %w", ctx.Err())
		case err := <-errChan:
			return nil, err
		case data := <-dataChan:
			return data, nil
		}
	}, nil
}
