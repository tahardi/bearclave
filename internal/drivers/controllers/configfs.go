package controllers

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

const (
	ConfigFSPath = "/sys/kernel/config"
	WriteOnly    = os.FileMode(0220)
)

var (
	ErrConfigFS = errors.New("configfs")
)

type CFSController interface {
	MkdirTemp(path string, pattern string) (name string, err error)
	RemoveAll(path string) (err error)
	ReadFile(path string) (data []byte, err error)
	WriteFile(path string, data []byte) (err error)
	Path() string
}

type MkdirTempFunc func(path string, pattern string) (string, error)

type ConfigFS struct {
	cfsPath       string
}

func NewConfigFS() (*ConfigFS, error) {
	return NewConfigFSWithPath(ConfigFSPath)
}

func NewConfigFSWithPath(cfsPath string) (*ConfigFS, error) {
	info, err := os.Stat(cfsPath)
	switch {
	case err == nil && info.IsDir():
		break
	case errors.Is(err, os.ErrNotExist):
		return nil, fmt.Errorf("%w: configfs path '%s' does not exist: %w", ErrConfigFS, cfsPath, err)
	case info != nil && !info.IsDir():
		return nil, fmt.Errorf("%w: configfs path '%s' is not a directory", ErrConfigFS, cfsPath)
	default:
		return nil, fmt.Errorf("%w: stat configfs path '%s': %w", ErrConfigFS, cfsPath, err)
	}
	return &ConfigFS{
		cfsPath: cfsPath,
	}, nil
}

func (c *ConfigFS) Path() string {
	return c.cfsPath
}

func (c *ConfigFS) MkdirTemp(path string, pattern string) (string, error) {
	if !strings.HasPrefix(path, c.cfsPath) {
		if !strings.HasPrefix(path, "/") {
			path = c.cfsPath + "/" + path
		} else {
			path = c.cfsPath + path
		}
	}
	name, err := os.MkdirTemp(path, pattern)
	if err != nil {
		return "", fmt.Errorf("%w: making dir '%s': %w", ErrConfigFS, path, err)
	}
	return name, nil
}

func (c *ConfigFS) RemoveAll(path string) error {
	if !strings.HasPrefix(path, c.cfsPath) {
		path = c.cfsPath + "/" + path
	}
	err := os.RemoveAll(path)
	if err != nil {
		return fmt.Errorf("%w: removing dir '%s': %w", ErrConfigFS, path, err)
	}
	return nil
}

func (c *ConfigFS) ReadFile(path string) ([]byte, error) {
	if !strings.HasPrefix(path, c.cfsPath) {
		path = c.cfsPath + "/" + path
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("%w: reading file '%s': %w", ErrConfigFS, path, err)
	}
	return data, nil
}

func (c *ConfigFS) WriteFile(path string, data []byte) error {
	if !strings.HasPrefix(path, c.cfsPath) {
		path = c.cfsPath + "/" + path
	}

	err := os.WriteFile(path, data, WriteOnly)
	if err != nil {
		return fmt.Errorf("%w: writing file '%s': %w", ErrConfigFS, path, err)
	}
	return nil
}
