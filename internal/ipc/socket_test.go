package ipc_test

import (
	"context"
	"encoding/base64"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/tahardi/bearclave/internal/ipc"
	"github.com/tahardi/bearclave/internal/mocks"
)

func TestSocketIPCRoundTrip(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// given
		ctx := context.Background()
		want := []byte("hello world")
		endpoint1 := "127.0.0.1:8080"
		endpoint2 := "127.0.0.1:8081"
		socketIPC1, err := ipc.NewSocketIPC(endpoint1)
		require.NoError(t, err)
		socketIPC2, err := ipc.NewSocketIPC(endpoint2)
		require.NoError(t, err)

		// when
		err = socketIPC1.Send(ctx, endpoint2, want)
		require.NoError(t, err)

		got, err := socketIPC2.Receive(ctx)
		require.NoError(t, err)

		// then
		assert.Equal(t, want, got)
	})
}

func TestSocketIPC_Send(t *testing.T) {
	t.Run("error - dialing", func(t *testing.T) {
		// given
		ctx := context.Background()
		want := []byte("hello world")
		endpoint := "127.0.0.1:8080"
		socketIPC, err := ipc.NewSocketIPCWithDialAndListener(
			func(string, string) (net.Conn, error) {
				return nil, assert.AnError
			},
			nil,
		)
		require.NoError(t, err)

		// when
		err = socketIPC.Send(ctx, endpoint, want)

		// then
		assert.ErrorContains(t, err, "dialing")
	})

	t.Run("error - writing data", func(t *testing.T) {
		// given
		ctx := context.Background()
		want := []byte("hello world")
		wantB64 := base64.StdEncoding.EncodeToString(want)
		endpoint := "127.0.0.1:8080"

		conn := mocks.NewConn(t)
		conn.On("Write", []byte(wantB64)).Return(len(wantB64), assert.AnError)
		conn.On("Close").Return(nil).Maybe()

		socketIPC, err := ipc.NewSocketIPCWithDialAndListener(
			func(string, string) (net.Conn, error) {
				return conn, nil
			},
			nil,
		)
		require.NoError(t, err)

		// when
		err = socketIPC.Send(ctx, endpoint, want)

		// then
		assert.ErrorContains(t, err, "writing data")
	})

	t.Run("error - writing all data", func(t *testing.T) {
		// given
		ctx := context.Background()
		want := []byte("hello world")
		wantB64 := base64.StdEncoding.EncodeToString(want)
		endpoint := "127.0.0.1:8080"

		conn := mocks.NewConn(t)
		conn.On("Write", []byte(wantB64)).Return(len(wantB64)-1, nil)
		conn.On("Close").Return(nil).Maybe()

		socketIPC, err := ipc.NewSocketIPCWithDialAndListener(
			func(string, string) (net.Conn, error) {
				return conn, nil
			},
			nil,
		)
		require.NoError(t, err)

		// when
		err = socketIPC.Send(ctx, endpoint, want)

		// then
		assert.ErrorContains(t, err, "failed to write all data")
	})

	t.Run("error - context cancelled", func(t *testing.T) {
		// given
		ctx, cancel := context.WithCancel(context.Background())
		want := []byte("hello world")
		wantB64 := base64.StdEncoding.EncodeToString(want)
		endpoint := "127.0.0.1:8080"

		conn := mocks.NewConn(t)
		conn.On("Write", []byte(wantB64)).Return(len(wantB64), nil).Maybe()
		conn.On("Close").Return(nil).Maybe()

		socketIPC, err := ipc.NewSocketIPCWithDialAndListener(
			func(string, string) (net.Conn, error) {
				return conn, nil
			},
			nil,
		)
		require.NoError(t, err)

		// when
		cancel()
		err = socketIPC.Send(ctx, endpoint, want)

		// then
		assert.ErrorContains(t, err, "context cancelled")
	})
}

func TestSocketIPC_Receive(t *testing.T) {
	t.Run("error - accepting connection", func(t *testing.T) {
		// given
		ctx := context.Background()

		listener := mocks.NewListener(t)
		listener.On("Accept").Return(nil, assert.AnError)

		socketIPC, err := ipc.NewSocketIPCWithDialAndListener(nil, listener)
		require.NoError(t, err)

		// when
		_, err = socketIPC.Receive(ctx)

		// then
		assert.ErrorContains(t, err, "accepting connection")
	})

	t.Run("error - reading data", func(t *testing.T) {
		// given
		ctx := context.Background()
		want := []byte("hello world")
		wantB64 := base64.StdEncoding.EncodeToString(want)

		conn := mocks.NewConn(t)
		conn.On("Read", mock.Anything).Return(len(wantB64), assert.AnError)
		conn.On("Close").Return(nil).Maybe()

		listener := mocks.NewListener(t)
		listener.On("Accept").Return(conn, nil)

		socketIPC, err := ipc.NewSocketIPCWithDialAndListener(nil, listener)
		require.NoError(t, err)

		// when
		_, err = socketIPC.Receive(ctx)

		// then
		assert.ErrorContains(t, err, "reading data")
	})

	t.Run("error - decoding data", func(t *testing.T) {
		// given
		ctx := context.Background()
		want := []byte("hello world")

		conn := mocks.NewConn(t)
		conn.On("Read", mock.Anything).Return(len(want), nil)
		conn.On("Close").Return(nil).Maybe()

		listener := mocks.NewListener(t)
		listener.On("Accept").Return(conn, nil)

		socketIPC, err := ipc.NewSocketIPCWithDialAndListener(nil, listener)
		require.NoError(t, err)

		// when
		_, err = socketIPC.Receive(ctx)

		// then
		assert.ErrorContains(t, err, "decoding data")
	})

	t.Run("error - context cancelled", func(t *testing.T) {
		// given
		ctx, cancel := context.WithCancel(context.Background())
		want := []byte("hello world")
		wantB64 := base64.StdEncoding.EncodeToString(want)

		conn := mocks.NewConn(t)
		conn.On("Read", mock.Anything).Return(len(wantB64), nil).Maybe()
		conn.On("Close").Return(nil).Maybe()

		listener := mocks.NewListener(t)
		listener.On("Accept").Return(conn, nil).Maybe()

		socketIPC, err := ipc.NewSocketIPCWithDialAndListener(nil, listener)
		require.NoError(t, err)

		// when
		cancel()
		_, err = socketIPC.Receive(ctx)

		// then
		assert.ErrorContains(t, err, "context cancelled")
	})
}
