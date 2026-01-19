package drivers

import (
	"errors"
	"fmt"
	"strings"

	"github.com/tahardi/bearclave/internal/drivers/controllers"
)

const (
	SEVProvider = "sev_guest"
)

var (
	ErrSEVClient = errors.New("sev client")
)

type SEV interface {
	GetReport(options ...SEVReportOption) (result *SEVReportResult, err error)
}

type SEVReportResult struct {
	Report    []byte `json:"report"`
	CertTable []byte `json:"certtable"`
}

type SEVReportOption func(*SEVReportOptions)
type SEVReportOptions struct {
	UserData  []byte
	CertTable bool
}

func MakeDefaultSEVReportOptions() SEVReportOptions {
	return SEVReportOptions{
		UserData:  nil,
		CertTable: false,
	}
}

func WithSEVReportUserData(userData []byte) SEVReportOption {
	return func(opts *SEVReportOptions) {
		opts.UserData = userData
	}
}

func WithSEVReportCertTable(certTable bool) SEVReportOption {
	return func(opts *SEVReportOptions) {
		opts.CertTable = certTable
	}
}

type SEVClient struct {
	tsm controllers.TSMController
}

func NewSEVClient() (*SEVClient, error) {
	tsm, err := controllers.NewTSM()
	if err != nil {
		return nil, fmt.Errorf("%w: making tsm controller: %w", ErrSEVClient, err)
	}
	return NewSEVClientWithTSM(tsm)
}

func NewSEVClientWithTSM(tsm controllers.TSMController) (*SEVClient, error) {
	return &SEVClient{tsm: tsm}, nil
}

// GetReport Information on auxillary blob and privilege levels taken from
// Linux TSM documentation (or from AMD docs reference in the TSM doc):
//
// https://www.kernel.org/doc/Documentation/ABI/testing/configfs-tsm
func (s *SEVClient) GetReport(options ...SEVReportOption) (*SEVReportResult, error) {
	opts := MakeDefaultSEVReportOptions()
	for _, opt := range options {
		opt(&opts)
	}

	result, err := s.tsm.GetReport(
		controllers.WithTSMReportInBlob(opts.UserData),
		controllers.WithTSMReportAuxBlob(opts.CertTable),
	)
	switch {
	case err != nil:
		return nil, fmt.Errorf("%w: getting report: %w", ErrSEVClient, err)
	case !strings.Contains(result.Provider, SEVProvider):
		return nil, fmt.Errorf("%w: unexpected provider '%s'", ErrSEVClient, result.Provider)
	}

	return &SEVReportResult{
		Report:    result.OutBlob,
		CertTable: result.AuxBlob,
	}, nil
}
