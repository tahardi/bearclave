package tdx_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tahardi/bearclave/internal/drivers"
	"github.com/tahardi/bearclave/internal/drivers/controllers"
)

func TestTDX_Drivers(t *testing.T) {
	t.Run("happy path - new tdx client", func(_ *testing.T) {
		client, err := drivers.NewTDXClient()
		require.NoError(t, err)
		require.NotNil(t, client)
		err = client.Close()
		require.NoError(t, err)
	})

	t.Run("happy path - get report", func(_ *testing.T) {
		// given
		reportData := []byte("reportData")
		client, err := drivers.NewTDXClient()
		require.NoError(t, err)
		require.NotNil(t, client)
		defer client.Close()

		// when
		report, err := client.GetReport0(reportData)

		// then
		require.NoError(t, err)
		require.NotEmpty(t, report)
	})

	t.Run("error - report data too long", func(_ *testing.T) {
		// given
		reportData := make([]byte, controllers.TDXReportDataLen+1)
		client, err := drivers.NewTDXClient()
		require.NoError(t, err)
		require.NotNil(t, client)
		defer client.Close()

		// when
		_, err = client.GetReport0(reportData)

		// then
		require.ErrorIs(t, err, controllers.ErrReportDataTooLong)
	})
}
