package ipc_test

import (
	"context"
	"encoding/base64"
	"net"
	"testing"

	"github.com/mdlayher/vsock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/tahardi/bearclave/internal/ipc"
	"github.com/tahardi/bearclave/internal/mocks"
)

func TestVSocketIPC_Send(t *testing.T) {
	t.Run("error - parsing endpoint", func(t *testing.T) {
		// given
		ctx := context.Background()
		want := []byte("hello world")
		endpoint := ":8080"
		socketIPC, err := ipc.NewVSocketIPCWithDialAndListener(
			func(uint32, uint32, *vsock.Config) (net.Conn, error) {
				return nil, assert.AnError
			},
			nil,
		)
		require.NoError(t, err)

		// when
		err = socketIPC.Send(ctx, endpoint, want)

		// then
		assert.ErrorContains(t, err, "parsing endpoint")
	})

	t.Run("error - dialing", func(t *testing.T) {
		// given
		ctx := context.Background()
		want := []byte("hello world")
		endpoint := "4:8080"
		socketIPC, err := ipc.NewVSocketIPCWithDialAndListener(
			func(uint32, uint32, *vsock.Config) (net.Conn, error) {
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
		endpoint := "4:8080"

		conn := mocks.NewConn(t)
		conn.On("Write", []byte(wantB64)).Return(len(wantB64), assert.AnError)
		conn.On("Close").Return(nil).Maybe()

		socketIPC, err := ipc.NewVSocketIPCWithDialAndListener(
			func(uint32, uint32, *vsock.Config) (net.Conn, error) {
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
		endpoint := "4:8080"

		conn := mocks.NewConn(t)
		conn.On("Write", []byte(wantB64)).Return(len(wantB64)-1, nil)
		conn.On("Close").Return(nil).Maybe()

		socketIPC, err := ipc.NewVSocketIPCWithDialAndListener(
			func(uint32, uint32, *vsock.Config) (net.Conn, error) {
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
		endpoint := "4:8080"

		conn := mocks.NewConn(t)
		conn.On("Write", []byte(wantB64)).Return(len(wantB64), nil).Maybe()
		conn.On("Close").Return(nil).Maybe()

		socketIPC, err := ipc.NewVSocketIPCWithDialAndListener(
			func(uint32, uint32, *vsock.Config) (net.Conn, error) {
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

func TestVSocketIPC_Receive(t *testing.T) {
	t.Run("error - accepting connection", func(t *testing.T) {
		// given
		ctx := context.Background()

		listener := mocks.NewListener(t)
		listener.On("Accept").Return(nil, assert.AnError)

		socketIPC, err := ipc.NewVSocketIPCWithDialAndListener(nil, listener)
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

		socketIPC, err := ipc.NewVSocketIPCWithDialAndListener(nil, listener)
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

		socketIPC, err := ipc.NewVSocketIPCWithDialAndListener(nil, listener)
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

		socketIPC, err := ipc.NewVSocketIPCWithDialAndListener(nil, listener)
		require.NoError(t, err)

		// when
		cancel()
		_, err = socketIPC.Receive(ctx)

		// then
		assert.ErrorContains(t, err, "context cancelled")
	})
}

func TestParseEndpoint(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// given
		wantCID := uint32(4)
		wantPort := uint32(8080)
		endpoint := "4:8080"

		// when
		cid, port, err := ipc.ParseEndpoint(endpoint)

		// then
		assert.NoError(t, err)
		assert.Equal(t, wantCID, cid)
		assert.Equal(t, wantPort, port)
	})

	t.Run("error - invalid format", func(t *testing.T) {
		// given
		endpoint := "string:8080"

		// when
		_, _, err := ipc.ParseEndpoint(endpoint)

		// then
		assert.ErrorContains(t, err, "invalid format")
	})
}
