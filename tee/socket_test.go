package tee_test

import (
	"context"
	"encoding/base64"
	"io"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tahardi/bearclave/mocks"
	"github.com/tahardi/bearclave/tee"
)

func TestSocketRoundTrip(t *testing.T) {
	t.Run("happy path - unix socket", func(t *testing.T) {
		// given
		ctx := context.Background()
		want := []byte("hello world")
		platform := tee.NoTEE
		network := "unix"

		addr1 := "127.0.0.1:8080"
		socket1, err := tee.NewSocket(ctx, platform, network, addr1)
		require.NoError(t, err)
		defer socket1.Close()

		addr2 := "127.0.0.1:8081"
		socket2, err := tee.NewSocket(ctx, platform, network, addr2)
		require.NoError(t, err)
		defer socket2.Close()

		// when
		err = socket1.Send(ctx, addr2, want)
		require.NoError(t, err)

		got, err := socket2.Receive(ctx)
		require.NoError(t, err)

		// then
		assert.Equal(t, want, got)
	})

	t.Run("happy path - tcp socket", func(t *testing.T) {
		// given
		ctx := context.Background()
		want := []byte("hello world")
		platform := tee.NoTEE
		network := "tcp"
		addr1 := "127.0.0.1:8080"
		socket1, err := tee.NewSocket(ctx, platform, network, addr1)
		require.NoError(t, err)
		defer socket1.Close()

		addr2 := "127.0.0.1:8081"
		socket2, err := tee.NewSocket(ctx, platform, network, addr2)
		require.NoError(t, err)
		defer socket2.Close()

		// when
		err = socket1.Send(ctx, addr2, want)
		require.NoError(t, err)

		got, err := socket2.Receive(ctx)
		require.NoError(t, err)

		// then
		assert.Equal(t, want, got)
	})
}

func TestSocket_SendWithDialContext(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// given
		ctx := context.Background()
		want := []byte("hello world")
		platform := tee.NoTEE
		network := "tcp"
		addr := "127.0.0.1:8080"

		connTimeout := tee.DefaultConnTimeout
		dialContext, err := tee.NewDialContext(platform)
		require.NoError(t, err)

		socket, err := tee.NewSocket(ctx, platform, network, addr)
		require.NoError(t, err)
		defer socket.Close()

		// when
		err = socket.SendWithDialContext(ctx, connTimeout, dialContext, addr, want)

		// then
		assert.NoError(t, err)
	})

	t.Run("error - dialing", func(t *testing.T) {
		// given
		ctx := context.Background()
		want := []byte("hello world")
		platform := tee.NoTEE
		network := "tcp"
		addr := "127.0.0.1:8080"

		connTimeout := tee.DefaultConnTimeout
		dialContext := func(context.Context, string, string) (net.Conn, error) {
			return nil, assert.AnError
		}
		socket, err := tee.NewSocket(ctx, platform, network, addr)
		require.NoError(t, err)
		defer socket.Close()

		// when
		err = socket.SendWithDialContext(ctx, connTimeout, dialContext, addr, want)

		// then
		assert.ErrorContains(t, err, "dialing")
	})

	t.Run("error - writing data", func(t *testing.T) {
		// given
		ctx := context.Background()
		want := []byte("hello world")
		wantB64 := base64.StdEncoding.EncodeToString(want)
		platform := tee.NoTEE
		network := "tcp"
		addr := "127.0.0.1:8080"

		connTimeout := tee.DefaultConnTimeout
		conn := mocks.NewConn(t)
		conn.On("Write", []byte(wantB64)).Return(len(wantB64), assert.AnError)
		conn.On("Close").Return(nil).Maybe()
		dialContext := func(context.Context, string, string) (net.Conn, error) {
			return conn, nil
		}

		socket, err := tee.NewSocket(ctx, platform, network, addr)
		require.NoError(t, err)
		defer socket.Close()

		// when
		err = socket.SendWithDialContext(ctx, connTimeout, dialContext, addr, want)

		// then
		assert.ErrorContains(t, err, "writing data")
	})

	t.Run("error - writing all data", func(t *testing.T) {
		// given
		ctx := context.Background()
		want := []byte("hello world")
		wantB64 := base64.StdEncoding.EncodeToString(want)
		platform := tee.NoTEE
		network := "tcp"
		addr := "127.0.0.1:8080"

		connTimeout := tee.DefaultConnTimeout
		conn := mocks.NewConn(t)
		conn.On("Write", []byte(wantB64)).Return(len(wantB64)-1, nil)
		conn.On("Close").Return(nil).Maybe()
		dialContext := func(context.Context, string, string) (net.Conn, error) {
			return conn, nil
		}

		socket, err := tee.NewSocket(ctx, platform, network, addr)
		require.NoError(t, err)
		defer socket.Close()

		// when
		err = socket.SendWithDialContext(ctx, connTimeout, dialContext, addr, want)

		// then
		assert.ErrorContains(t, err, "failed to write all data")
	})

	t.Run("error - send context cancelled", func(t *testing.T) {
		// given
		ctx, cancel := context.WithCancel(context.Background())
		want := []byte("hello world")
		wantB64 := base64.StdEncoding.EncodeToString(want)
		platform := tee.NoTEE
		network := "tcp"
		addr := "127.0.0.1:8080"

		connTimeout := tee.DefaultConnTimeout
		conn := mocks.NewConn(t)
		conn.On("Write", []byte(wantB64)).Return(len(wantB64), nil).Maybe()
		conn.On("Close").Return(nil).Maybe()
		dialContext := func(context.Context, string, string) (net.Conn, error) {
			return conn, nil
		}

		socket, err := tee.NewSocket(ctx, platform, network, addr)
		require.NoError(t, err)
		defer socket.Close()

		// when
		cancel()
		err = socket.SendWithDialContext(ctx, connTimeout, dialContext, addr, want)

		// then
		assert.ErrorContains(t, err, "context cancelled")
	})

	t.Run("error - dial context timeout", func(t *testing.T) {
		// given
		ctx := context.Background()
		want := []byte("hello world")
		platform := tee.NoTEE
		network := "tcp"
		addr := "127.0.0.1:8080"

		// I want to test that the actual Bearclave.DialContext implementation
		// respects context timeouts. Thus, I'm wrapping it so that I can
		// ensure that enough time has passed for the context to timeout.
		connTimeout := 10 * time.Millisecond
		dialContext := func(c context.Context, network string, addr string) (net.Conn, error) {
			// Make sure "establishing the connection" takes longer than connTimeout
			time.Sleep(connTimeout * 2)
			dc, err := tee.NewDialContext(platform)
			require.NoError(t, err)
			return dc(c, network, addr)
		}

		socket, err := tee.NewSocket(ctx, platform, network, addr)
		require.NoError(t, err)
		defer socket.Close()

		// when
		err = socket.SendWithDialContext(ctx, connTimeout, dialContext, addr, want)

		// then
		require.Error(t, err)
		assert.ErrorContains(t, err, "timeout")
	})
}

func TestSocket_Receive(t *testing.T) {
	t.Run("error - accepting connection", func(t *testing.T) {
		// given
		ctx := context.Background()
		network := "tcp"
		platform := tee.NoTEE

		listener := mocks.NewListener(t)
		listener.On("Accept").Return(nil, assert.AnError)
		listener.On("Close").Return(nil).Maybe()

		socket, err := tee.NewSocketWithListener(platform, network, listener)
		require.NoError(t, err)
		defer socket.Close()

		// when
		_, err = socket.Receive(ctx)

		// then
		assert.ErrorContains(t, err, "accepting connection")
	})

	t.Run("error - reading data", func(t *testing.T) {
		// given
		ctx := context.Background()
		want := []byte("hello world")
		wantB64 := base64.StdEncoding.EncodeToString(want)
		network := "tcp"
		platform := tee.NoTEE

		conn := mocks.NewConn(t)
		conn.On("Read", mock.Anything).Return(len(wantB64), assert.AnError)
		conn.On("Close").Return(nil).Maybe()

		listener := mocks.NewListener(t)
		listener.On("Accept").Return(conn, nil)
		listener.On("Close").Return(nil).Maybe()

		socket, err := tee.NewSocketWithListener(platform, network, listener)
		require.NoError(t, err)
		defer socket.Close()

		// when
		_, err = socket.Receive(ctx)

		// then
		assert.ErrorContains(t, err, "reading data")
	})

	t.Run("error - decoding data", func(t *testing.T) {
		// given
		ctx := context.Background()
		want := []byte("hello world")
		network := "tcp"
		platform := tee.NoTEE

		conn := mocks.NewConn(t)
		conn.On("Read", mock.Anything).Return(len(want), nil).Once()
		conn.On("Read", mock.Anything).Return(0, io.EOF).Once()
		conn.On("Close").Return(nil).Maybe()

		listener := mocks.NewListener(t)
		listener.On("Accept").Return(conn, nil)
		listener.On("Close").Return(nil).Maybe()

		socket, err := tee.NewSocketWithListener(platform, network, listener)
		require.NoError(t, err)
		defer socket.Close()

		// when
		_, err = socket.Receive(ctx)

		// then
		assert.ErrorContains(t, err, "decoding data")
	})

	t.Run("error - context cancelled", func(t *testing.T) {
		// given
		ctx, cancel := context.WithCancel(context.Background())
		want := []byte("hello world")
		wantB64 := base64.StdEncoding.EncodeToString(want)
		network := "tcp"
		platform := tee.NoTEE

		conn := mocks.NewConn(t)
		conn.On("Read", mock.Anything).Return(len(wantB64), nil).Maybe()
		conn.On("Close").Return(nil).Maybe()

		listener := mocks.NewListener(t)
		listener.On("Accept").Return(conn, nil).Maybe()
		listener.On("Close").Return(nil).Maybe()

		socket, err := tee.NewSocketWithListener(platform, network, listener)
		require.NoError(t, err)
		defer socket.Close()

		// when
		cancel()
		_, err = socket.Receive(ctx)

		// then
		assert.ErrorContains(t, err, "context cancelled")
	})
}
