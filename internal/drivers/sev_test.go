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

func TestSEVClient_Interfaces(t *testing.T) {
	t.Run("SEV", func(_ *testing.T) {
		var _ drivers.SEV = &drivers.SEVClient{}
	})
}

func TestSEVClient_GetReport(t *testing.T) {
	t.Run("happy path", func(_ *testing.T) {
		// given
		want := &controllers.TSMReportResult{
			OutBlob:  []byte("report"),
			AuxBlob:  []byte("cert table"),
			Provider: drivers.SEVProvider,
		}
		tsm := mocks.NewTSMController(t)
		tsm.On("GetReport", mock.Anything).Return(want, nil)

		client, err := drivers.NewSEVClientWithTSM(tsm)
		require.NoError(t, err)

		// when
		got, err := client.GetReport(drivers.WithSEVReportCertTable(true))

		// then
		require.NoError(t, err)
		assert.Equal(t, want.OutBlob, got.Report)
		assert.Equal(t, want.AuxBlob, got.CertTable)
	})

	t.Run("error - tsm", func(_ *testing.T) {
		// given
		tsm := mocks.NewTSMController(t)
		tsm.On("GetReport", mock.Anything).Return(nil, assert.AnError)

		client, err := drivers.NewSEVClientWithTSM(tsm)
		require.NoError(t, err)

		// when
		_, err = client.GetReport()

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

		client, err := drivers.NewSEVClientWithTSM(tsm)
		require.NoError(t, err)

		// when
		_, err = client.GetReport()

		// then
		require.ErrorIs(t, err, drivers.ErrSEVClient)
	})
}
