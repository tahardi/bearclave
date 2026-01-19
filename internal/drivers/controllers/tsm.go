package controllers

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Documentation for TSM and Report API taken from:
// https://www.kernel.org/doc/Documentation/ABI/testing/configfs-tsm
//
// Useful types and constants take from Linux kernel:
// https://github.com/torvalds/linux/blob/944aacb68baf7624ab8d277d0ebf07f025ca137c/include/linux/tsm.h
const (
	TSMPath                            = "/tsm"
	TSMReportPath                      = TSMPath + "/report"
	TSMReportPattern                   = "bearclave-report-*"
	TSMReportInBlob                    = "/inblob"
	TSMReportInBlobSize                = 64
	TSMReportOutBlob                   = "/outblob"
	TSMReportAuxBlob                   = "/auxblob"
	TSMReportManifestBlob              = "/manifestblob"
	TSMReportProvider                  = "/provider"
	TSMReportGeneration                = "/generation"
	TSMReportPrivLevel                 = "/privlevel"
	TSMReportPrivLevelFloor            = "/privlevel_floor"
	TSMReportPrivLevelNil              = -1
	TSMReportPrivLevelMax              = 3
	TSMReportServiceProvider           = "/service_provider"
	TSMReportServiceGUID               = "/service_guid"
	TSMReportServiceManifestVersion    = "/service_manifest_version"
	TSMReportServiceManifestVersionNil = -1
	TSMReportUintBase                  = 10
	TSMReportUintSize                  = 64
)

var (
	ErrTSM = errors.New("tsm")
	ErrTSMReport = fmt.Errorf("%w: report", ErrTSM)
)

type TSMController interface {
	GetReport(options ...TSMReportOption) (tsmResult *TSMReportResult, err error)
}

type TSMReportResult struct {
	OutBlob      []byte `json:"outblob"`
	AuxBlob      []byte `json:"auxblob,omitempty"`
	ManifestBlob []byte `json:"manifestblob,omitempty"`
	Provider     string `json:"provider"`
}

type TSMReportOption func(*TSMReportOptions)
type TSMReportOptions struct {
	InBlob                 []byte
	AuxBlob                bool
	ManifestBlob           bool
	PrivLevel              int
	ServiceProvider        []byte
	ServiceGUID            []byte
	ServiceManifestVersion int
}

func MakeDefaultTSMReportOptions() TSMReportOptions {
	return TSMReportOptions{
		InBlob:                 nil,
		AuxBlob:                false,
		ManifestBlob:           false,
		PrivLevel:              TSMReportPrivLevelNil,
		ServiceProvider:        nil,
		ServiceGUID:            nil,
		ServiceManifestVersion: TSMReportServiceManifestVersionNil,
	}
}

func WithTSMReportInBlob(inBlob []byte) TSMReportOption {
	return func(opts *TSMReportOptions) {
		opts.InBlob = inBlob
	}
}

func WithTSMReportAuxBlob() TSMReportOption {
	return func(opts *TSMReportOptions) {
		opts.AuxBlob = true
	}
}

func WithTSMReportManifestBlob() TSMReportOption {
	return func(opts *TSMReportOptions) {
		opts.ManifestBlob = true
	}
}

func WithTSMReportPrivLevel(privLevel int) TSMReportOption {
	return func(opts *TSMReportOptions) {
		opts.PrivLevel = privLevel
	}
}

func WithTSMReportServiceProvider(serviceProvider string) TSMReportOption {
	return func(opts *TSMReportOptions) {
		opts.ServiceProvider = []byte(serviceProvider)
	}
}

func WithTSMReportServiceGUID(serviceGUID []byte) TSMReportOption {
	return func(opts *TSMReportOptions) {
		opts.ServiceGUID = serviceGUID
	}
}

func WithTSMReportServiceManifestVersion(serviceManifestVersion int) TSMReportOption {
	return func(opts *TSMReportOptions) {
		opts.ServiceManifestVersion = serviceManifestVersion
	}
}

type TSM struct {
	configFS CFSController
}

func NewTSM() (*TSM, error) {
	configFS, err := NewConfigFS()
	if err != nil {
		return nil, err
	}
	return NewTSMWithConfigFS(configFS)
}

func NewTSMWithConfigFS(configFS CFSController) (*TSM, error) {
	tmsPath := configFS.Path() + "/" + TSMPath
	info, err := os.Stat(tmsPath)
	switch {
	case err == nil && info.IsDir():
		break
	case errors.Is(err, os.ErrNotExist):
		return nil, fmt.Errorf("%w: tsm path '%s' does not exist: %w", ErrTSM, tmsPath, err)
	case info != nil && !info.IsDir():
		return nil, fmt.Errorf("%w: tsm path '%s' is not a directory", ErrTSM, tmsPath)
	default:
		return nil, fmt.Errorf("%w: unexpected error checking tsm path '%s': %w", ErrTSM, tmsPath, err)
	}
	return &TSM{configFS: configFS}, nil
}

func (t *TSM) GetReport(options ...TSMReportOption) (*TSMReportResult, error) {
	opts := MakeDefaultTSMReportOptions()
	for _, opt := range options {
		opt(&opts)
	}

	reportPath, err := t.configFS.MkdirTemp(TSMReportPath, TSMReportPattern)
	if err != nil {
		return nil, fmt.Errorf(
			"%w: making dir '%s': %w",
			ErrTSMReport, reportPath, err,
		)
	}
	defer t.configFS.RemoveAll(reportPath)

	err = t.setReportAttributes(reportPath, opts)
	if err != nil {
		return nil, err
	}

	return t.getReportAttributes(reportPath, opts)
}

func (t *TSM) getReportAttributes(
	reportPath string,
	opts TSMReportOptions,
) (*TSMReportResult, error) {
	result := &TSMReportResult{}

	// Always read the outblob because it contains the attestation report.
	outBlobPath := reportPath + TSMReportOutBlob
	outBlob, err := t.configFS.ReadFile(outBlobPath)
	if err != nil {
		return nil, fmt.Errorf(
			"%w: reading outblob from '%s': %w",
			ErrTSMReport, outBlobPath, err,
		)
	}
	result.OutBlob = outBlob

	// Always read the provider so users can verify they are on the right platform.
	providerPath := reportPath + TSMReportProvider
	provider, err := t.configFS.ReadFile(providerPath)
	if err != nil {
		return nil, fmt.Errorf(
			"%w: reading provider from '%s': %w",
			ErrTSMReport, providerPath, err,
		)
	}
	result.Provider = string(provider)

	if opts.AuxBlob {
		auxBlobPath := reportPath + TSMReportAuxBlob
		auxBlob, err := t.configFS.ReadFile(auxBlobPath)
		if err != nil {
			return nil, fmt.Errorf(
				"%w: reading auxblob from '%s': %w",
				ErrTSMReport, auxBlobPath, err,
			)
		}
		result.AuxBlob = auxBlob
	}

	if opts.ManifestBlob {
		manifestBlobPath := reportPath + TSMReportManifestBlob
		manifestBlob, err := t.configFS.ReadFile(manifestBlobPath)
		if err != nil {
			return nil, fmt.Errorf(
				"%w: reading manifestblob from '%s': %w",
				ErrTSMReport, manifestBlobPath, err,
			)
		}
		result.ManifestBlob = manifestBlob
	}
	return result, nil
}

func (t *TSM) setReportAttributes(
	reportPath string,
	opts TSMReportOptions,
) error {
	// The generation file contains a number that gets incremented every time
	// one of the write-only attribute files gets modified (e.g., inblob).
	// Check the number at the beginning and end of our configuration to ensure
	// it has been incremented the expected number of times. If not, then
	// something went wrong (like someone else editing our attribute files)
	generationPath := reportPath + TSMReportGeneration
	expectedGeneration, err := t.readUint64(generationPath)
	if err != nil {
		return fmt.Errorf(
			"%w: reading generation from '%s': %w",
			ErrTSMReport, generationPath, err,
		)
	}

	if opts.PrivLevel != TSMReportPrivLevelNil {
		privLevelFloorPath := reportPath + TSMReportPrivLevelFloor
		privLevelFloor, err := t.readUint64(privLevelFloorPath)
		switch {
		case err != nil:
			return fmt.Errorf(
				"%w: reading priv level floor from '%s': %w",
				ErrTSMReport, privLevelFloorPath, err,
			)
		case opts.PrivLevel < int(privLevelFloor):
			return fmt.Errorf(
				"%w: priv level %d is below floor %d",
				ErrTSMReport, opts.PrivLevel, privLevelFloor,
			)
		case opts.PrivLevel > TSMReportPrivLevelMax:
			return fmt.Errorf(
				"%w: priv level %d is above max %d",
				ErrTSMReport, opts.PrivLevel, TSMReportPrivLevelMax,
			)
		}

		privLevelPath := reportPath + TSMReportPrivLevel
		err = t.writeUint64(privLevelPath, uint64(opts.PrivLevel))
		if err != nil {
			return fmt.Errorf(
				"%w: writing priv level to '%s': %w",
				ErrTSMReport, privLevelPath, err,
			)
		}
		expectedGeneration++
	}

	if opts.ServiceProvider != nil {
		serviceProviderPath := reportPath + TSMReportServiceProvider
		err = t.configFS.WriteFile(serviceProviderPath, opts.ServiceProvider)
		if err != nil {
			return fmt.Errorf(
				"%w: writing service provider to '%s': %w",
				ErrTSMReport, serviceProviderPath, err,
			)
		}
		expectedGeneration++
	}

	if opts.ServiceGUID != nil {
		serviceGUIDPath := reportPath + TSMReportServiceGUID
		err = t.configFS.WriteFile(serviceGUIDPath, opts.ServiceGUID)
		if err != nil {
			return fmt.Errorf(
				"%w: writing service GUID to '%s': %w",
				ErrTSMReport, serviceGUIDPath, err,
			)
		}
		expectedGeneration++
	}

	if opts.ServiceManifestVersion != TSMReportServiceManifestVersionNil {
		serviceManifestVersionPath := reportPath + TSMReportServiceManifestVersion
		err = t.writeUint64(serviceManifestVersionPath, uint64(opts.ServiceManifestVersion))
		if err != nil {
			return fmt.Errorf(
				"%w: writing service manifest version to '%s': %w",
				ErrTSMReport, serviceManifestVersionPath, err,
			)
		}
		expectedGeneration++
	}

	if opts.InBlob != nil {
		inBlobFixedSize := make([]byte, TSMReportInBlobSize)
		copy(inBlobFixedSize[:], opts.InBlob)

		inBlobPath := reportPath + TSMReportInBlob
		err = t.configFS.WriteFile(inBlobPath, inBlobFixedSize)
		if err != nil {
			return fmt.Errorf(
				"%w: writing inblob to '%s': %w",
				ErrTSMReport, inBlobPath, err,
			)
		}
		expectedGeneration++
	}

	gotGeneration, err := t.readUint64(generationPath)
	switch {
	case err != nil:
		return fmt.Errorf(
			"%w: reading generation from '%s': %w",
			ErrTSMReport, generationPath, err,
		)
	case gotGeneration != expectedGeneration:
		return fmt.Errorf(
			"%w: generation mismatch, expected %d, got %d",
			ErrTSMReport, expectedGeneration, gotGeneration,
		)
	}
	return nil
}

func (t *TSM) readUint64(path string) (uint64, error) {
	bytes, err := t.configFS.ReadFile(path)
	if err != nil {
		return 0, err
	}

	trimmed := strings.TrimRight(string(bytes), "\n")
	return strconv.ParseUint(trimmed, TSMReportUintBase, TSMReportUintSize)
}

func (t *TSM) writeUint64(path string, value uint64) error {
	bytes := []byte(strconv.FormatUint(value, TSMReportUintBase))
	return t.configFS.WriteFile(path, bytes)
}
