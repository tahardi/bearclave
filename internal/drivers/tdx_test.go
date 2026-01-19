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

func TestTDXClient_GetReport(t *testing.T) {
	t.Run("happy path - no user data", func(_ *testing.T) {
		// given
		want := &controllers.TSMReportResult{
			OutBlob:  []byte("report"),
			Provider: drivers.TDXProvider,
		}
		tsm := mocks.NewTSMController(t)
		tsm.On("GetReport", mock.Anything).Return(want, nil)

		client, err := drivers.NewTDXClientWithTSM(tsm)
		require.NoError(t, err)

		// when
		got, err := client.GetReport(nil)

		// then
		require.NoError(t, err)
		assert.Equal(t, want.OutBlob, got)
	})

	t.Run("error - tsm", func(_ *testing.T) {
		// given
		tsm := mocks.NewTSMController(t)
		tsm.On("GetReport", mock.Anything).Return(nil, assert.AnError)

		client, err := drivers.NewTDXClientWithTSM(tsm)
		require.NoError(t, err)

		// when
		_, err = client.GetReport(nil)

		// then
		require.ErrorIs(t, err, assert.AnError)
	})

	t.Run("error - wrong provider", func(_ *testing.T) {
		// given
		want := &controllers.TSMReportResult{
			OutBlob:  []byte("report"),
			Provider: "wrong-provider",
		}
		tsm := mocks.NewTSMController(t)
		tsm.On("GetReport", mock.Anything).Return(want, nil)

		client, err := drivers.NewTDXClientWithTSM(tsm)
		require.NoError(t, err)

		// when
		_, err = client.GetReport(nil)

		// then
		require.ErrorIs(t, err, drivers.ErrTDXClient)
	})
}
