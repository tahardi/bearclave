package drivers_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tahardi/bearclave/internal/drivers"
	"github.com/tahardi/bearclave/internal/drivers/controllers"
	"github.com/tahardi/bearclave/mocks"
)

func TestTDXClient_Interfaces(t *testing.T) {
	t.Run("TDX", func(_ *testing.T) {
		var _ drivers.TDX = &drivers.TDXClient{}
	})
}

func TestTDXClient_GetReport0(t *testing.T) {
	t.Run("happy path", func(_ *testing.T) {
		// given
		reportData := []byte("reportData")
		want := []byte("report")
		ioctl := mocks.NewIOController(t)
		ioctl.On("Send", mock.Anything).Return(want, nil)
		client, err := drivers.NewTDXClientWithController(ioctl)
		require.NoError(t, err)

		// when
		got, err := client.GetReport0(reportData)

		// then
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("happy path - nil report data", func(_ *testing.T) {
		// given
		want := []byte("report")
		ioctl := mocks.NewIOController(t)
		ioctl.On("Send", mock.Anything).Return(want, nil)
		client, err := drivers.NewTDXClientWithController(ioctl)
		require.NoError(t, err)

		// when
		got, err := client.GetReport0(nil)

		// then
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("error - ioctl", func(t *testing.T) {
		// given
		reportData := []byte("reportData")
		ioctl := mocks.NewIOController(t)
		ioctl.On("Send", mock.Anything).Return(nil, assert.AnError)

		client, err := drivers.NewTDXClientWithController(ioctl)
		require.NoError(t, err)

		// when
		_, err = client.GetReport0(reportData)

		// then
		require.ErrorIs(t, err, assert.AnError)
	})

	t.Run("error - report data too long", func(t *testing.T) {
		// given
		reportData := make([]byte, controllers.TDXReportDataLen+1)
		ioctl, err := controllers.NewTDXControllerWithFile(nil)
		require.NoError(t, err)

		client, err := drivers.NewTDXClientWithController(ioctl)
		require.NoError(t, err)

		// when
		_, err = client.GetReport0(reportData)

		// then
		require.ErrorIs(t, err, controllers.ErrReportDataTooLong)
	})

	t.Run("error - empty report", func(_ *testing.T) {
		// given
		reportData := []byte("reportData")
		want := []byte{}
		ioctl := mocks.NewIOController(t)
		ioctl.On("Send", mock.Anything).Return(want, nil)
		client, err := drivers.NewTDXClientWithController(ioctl)
		require.NoError(t, err)

		// when
		_, err = client.GetReport0(reportData)

		// then
		require.ErrorIs(t, err, drivers.ErrEmptyReport)
	})
}
