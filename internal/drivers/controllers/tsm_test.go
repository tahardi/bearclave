package controllers_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tahardi/bearclave/internal/drivers/controllers"
	"github.com/tahardi/bearclave/mocks"
)

func TestTSM_Interfaces(t *testing.T) {
	t.Run("TSMController", func(_ *testing.T) {
		var _ controllers.TSMController = &controllers.TSM{}
	})
}

func TestNewTMSWithConfigFS(t *testing.T) {
	t.Run("happy path", func(_ *testing.T) {
		// given
		cfsPath := os.TempDir() + "/configfs"
		tmsPath := controllers.TSMPath
		defer os.RemoveAll(cfsPath)

		err := os.MkdirAll(cfsPath+tmsPath, os.FileMode(0700))
		require.NoError(t, err)

		cfs := mocks.NewCFSController(t)
		cfs.On("Path").Return(cfsPath)

		// when
		_, err = controllers.NewTSMWithConfigFS(cfs)

		// then
		require.NoError(t, err)
	})

	t.Run("error - directory does not exist", func(_ *testing.T) {
		// given
		cfsPath := os.TempDir() + "/configfs"
		defer os.RemoveAll(cfsPath)

		err := os.MkdirAll(cfsPath, os.FileMode(0700))
		require.NoError(t, err)

		cfs := mocks.NewCFSController(t)
		cfs.On("Path").Return(cfsPath)

		// when
		_, err = controllers.NewTSMWithConfigFS(cfs)

		// then
		require.ErrorIs(t, err, os.ErrNotExist)
	})

	t.Run("error - path is not a directory", func(_ *testing.T) {
		// given
		cfsPath := os.TempDir()
		require.NoError(t, os.WriteFile(cfsPath+controllers.TSMPath, []byte{}, 0600))

		cfs := mocks.NewCFSController(t)
		cfs.On("Path").Return(cfsPath)

		// when
		_, err := controllers.NewTSMWithConfigFS(cfs)

		// then
		require.ErrorIs(t, err, controllers.ErrTSM)
	})
}
