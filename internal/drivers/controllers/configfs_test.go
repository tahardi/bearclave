package controllers_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tahardi/bearclave/internal/drivers/controllers"
)

func TestConfigFS_Interfaces(t *testing.T) {
	t.Run("CFSController", func(_ *testing.T) {
		var _ controllers.CFSController = &controllers.ConfigFS{}
	})
}

func TestNewConfigFSWithPath(t *testing.T) {
	t.Run("happy path", func(_ *testing.T) {
		// given
		cfsPath := os.TempDir()

		// when
		_, err := controllers.NewConfigFSWithPath(cfsPath)

		// then
		require.NoError(t, err)
	})

	t.Run("error - directory does not exist", func(_ *testing.T) {
		// given
		cfsPath := os.TempDir() + "/nonexistent"

		// when
		_, err := controllers.NewConfigFSWithPath(cfsPath)

		// then
		require.ErrorIs(t, err, os.ErrNotExist)
	})

	t.Run("error - path is not a directory", func(_ *testing.T) {
		// given
		cfsPath := os.TempDir() + "/file"
		require.NoError(t, os.WriteFile(cfsPath, []byte{}, 0600))

		// when
		_, err := controllers.NewConfigFSWithPath(cfsPath)

		// then
		require.ErrorIs(t, err, controllers.ErrConfigFS)
	})
}

func TestConfigFS_MkdirTemp(t *testing.T) {
	t.Run("happy path", func(_ *testing.T) {
		// given
		cfsPath := os.TempDir() + "/configfs"
		reportPath := "/tsm/report"
		err := os.MkdirAll(cfsPath+reportPath, os.FileMode(0700))
		require.NoError(t, err)
		defer os.RemoveAll(cfsPath)

		pattern := "bearclave-report-*"
		cfs, err := controllers.NewConfigFSWithPath(cfsPath)
		require.NoError(t, err)

		// when
		name, err := cfs.MkdirTemp(reportPath, pattern)

		// then
		require.NoError(t, err)
		require.DirExists(t, name)
	})
}

func TestConfigFS_RemoveAll(t *testing.T) {
	t.Run("happy path", func(_ *testing.T) {
		// given
		cfsPath := os.TempDir() + "/configfs"
		reportPath := "/tsm/report"
		err := os.MkdirAll(cfsPath+reportPath, os.FileMode(0700))
		require.NoError(t, err)
		defer os.RemoveAll(cfsPath)

		cfs, err := controllers.NewConfigFSWithPath(cfsPath)
		require.NoError(t, err)

		// when
		err = cfs.RemoveAll(reportPath)

		// then
		require.NoError(t, err)
		require.NoDirExists(t, cfsPath+reportPath)
	})
}

func TestConfigFS_ReadFile(t *testing.T) {
	t.Run("happy path", func(_ *testing.T) {
		// given
		cfsPath := os.TempDir() + "/configfs"
		err := os.MkdirAll(cfsPath, os.FileMode(0700))
		require.NoError(t, err)
		defer os.RemoveAll(cfsPath)

		reportPath := "/report"
		want := []byte("hello world")
		err = os.WriteFile(cfsPath+reportPath, want, os.FileMode(0600))
		require.NoError(t, err)

		cfs, err := controllers.NewConfigFSWithPath(cfsPath)
		require.NoError(t, err)

		// when
		got, err := cfs.ReadFile(reportPath)

		// then
		require.NoError(t, err)
		require.Equal(t, want, got)
	})
}

func TestConfigFS_WriteFile(t *testing.T) {
	t.Run("happy path", func(_ *testing.T) {
		// given
		cfsPath := os.TempDir() + "/configfs"
		err := os.MkdirAll(cfsPath, os.FileMode(0700))
		require.NoError(t, err)
		defer os.RemoveAll(cfsPath)

		reportPath := "/report"
		want := []byte("hello world")

		cfs, err := controllers.NewConfigFSWithPath(cfsPath)
		require.NoError(t, err)

		// when
		err = cfs.WriteFile(reportPath, want)

		// then
		require.NoError(t, err)

		// Our ConfigFS should make write-only files. Make sure we can't read
		_, err = cfs.ReadFile(reportPath)
		require.Error(t, err)
	})
}
