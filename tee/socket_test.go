package tee_test

import (
	"context"
	"encoding/base64"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tahardi/bearclave"
	"github.com/tahardi/bearclave/mocks"
	"github.com/tahardi/bearclave/tee"
)

func TestSocketRoundTrip(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// given
		ctx := context.Background()
		want := []byte("hello world")
		platform := bearclave.NoTEE
		network := "tcp"
		addr1 := "127.0.0.1:8080"
		addr2 := "127.0.0.1:8081"
		socket1, err := tee.NewSocket(platform, network, addr1)
		require.NoError(t, err)
		socket2, err := tee.NewSocket(platform, network, addr2)
		require.NoError(t, err)
		defer socket1.Close()
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

func TestSocket_Send(t *testing.T) {
	t.Run("error - dialing", func(t *testing.T) {
		// given
		ctx := context.Background()
		want := []byte("hello world")
		platform := bearclave.NoTEE
		network := "tcp"
		addr := "127.0.0.1:8080"

		dialer := func(string, string) (net.Conn, error) {
			return nil, assert.AnError
		}
		socket, err := tee.NewSocket(platform, network, addr)
		require.NoError(t, err)
		defer socket.Close()

		// when
		err = socket.SendWithDialer(ctx, dialer, addr, want)

		// then
		assert.ErrorContains(t, err, "dialing")
	})

	t.Run("error - writing data", func(t *testing.T) {
		// given
		ctx := context.Background()
		want := []byte("hello world")
		wantB64 := base64.StdEncoding.EncodeToString(want)
		platform := bearclave.NoTEE
		network := "tcp"
		addr := "127.0.0.1:8080"

		conn := mocks.NewConn(t)
		conn.On("Write", []byte(wantB64)).Return(len(wantB64), assert.AnError)
		conn.On("Close").Return(nil).Maybe()
		dialer := func(string, string) (net.Conn, error) {
			return conn, nil
		}

		socket, err := tee.NewSocket(platform, network, addr)
		require.NoError(t, err)
		defer socket.Close()

		// when
		err = socket.SendWithDialer(ctx, dialer, addr, want)

		// then
		assert.ErrorContains(t, err, "writing data")
	})

	t.Run("error - writing all data", func(t *testing.T) {
		// given
		ctx := context.Background()
		want := []byte("hello world")
		wantB64 := base64.StdEncoding.EncodeToString(want)
		platform := bearclave.NoTEE
		network := "tcp"
		addr := "127.0.0.1:8080"

		conn := mocks.NewConn(t)
		conn.On("Write", []byte(wantB64)).Return(len(wantB64)-1, nil)
		conn.On("Close").Return(nil).Maybe()
		dialer := func(string, string) (net.Conn, error) {
			return conn, nil
		}

		socket, err := tee.NewSocket(platform, network, addr)
		require.NoError(t, err)
		defer socket.Close()

		// when
		err = socket.SendWithDialer(ctx, dialer, addr, want)

		// then
		assert.ErrorContains(t, err, "failed to write all data")
	})

	t.Run("error - context cancelled", func(t *testing.T) {
		// given
		ctx, cancel := context.WithCancel(context.Background())
		want := []byte("hello world")
		wantB64 := base64.StdEncoding.EncodeToString(want)
		platform := bearclave.NoTEE
		network := "tcp"
		addr := "127.0.0.1:8080"

		conn := mocks.NewConn(t)
		conn.On("Write", []byte(wantB64)).Return(len(wantB64), nil).Maybe()
		conn.On("Close").Return(nil).Maybe()
		dialer := func(string, string) (net.Conn, error) {
			return conn, nil
		}

		socket, err := tee.NewSocket(platform, network, addr)
		require.NoError(t, err)
		defer socket.Close()

		// when
		cancel()
		err = socket.SendWithDialer(ctx, dialer, addr, want)

		// then
		assert.ErrorContains(t, err, "context cancelled")
	})
}

func TestSocket_Receive(t *testing.T) {
	t.Run("error - accepting connection", func(t *testing.T) {
		// given
		ctx := context.Background()
		network := "tcp"
		platform := bearclave.NoTEE

		listener := mocks.NewListener(t)
		listener.On("Accept").Return(nil, assert.AnError)

		socket, err := tee.NewSocketWithListener(platform, network, listener)
		require.NoError(t, err)

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
		platform := bearclave.NoTEE

		conn := mocks.NewConn(t)
		conn.On("Read", mock.Anything).Return(len(wantB64), assert.AnError)
		conn.On("Close").Return(nil).Maybe()

		listener := mocks.NewListener(t)
		listener.On("Accept").Return(conn, nil)

		socket, err := tee.NewSocketWithListener(platform, network, listener)
		require.NoError(t, err)

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
		platform := bearclave.NoTEE

		conn := mocks.NewConn(t)
		conn.On("Read", mock.Anything).Return(len(want), nil)
		conn.On("Close").Return(nil).Maybe()

		listener := mocks.NewListener(t)
		listener.On("Accept").Return(conn, nil)

		socket, err := tee.NewSocketWithListener(platform, network, listener)
		require.NoError(t, err)

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
		platform := bearclave.NoTEE

		conn := mocks.NewConn(t)
		conn.On("Read", mock.Anything).Return(len(wantB64), nil).Maybe()
		conn.On("Close").Return(nil).Maybe()

		listener := mocks.NewListener(t)
		listener.On("Accept").Return(conn, nil).Maybe()

		socket, err := tee.NewSocketWithListener(platform, network, listener)
		require.NoError(t, err)

		// when
		cancel()
		_, err = socket.Receive(ctx)

		// then
		assert.ErrorContains(t, err, "context cancelled")
	})
}
