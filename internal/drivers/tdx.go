package drivers

import (
	"errors"
	"fmt"
	"strings"

	"github.com/tahardi/bearclave/internal/drivers/controllers"
)

const (
	TDXProvider = "tdx_guest"
)

var (
	ErrTDXClient = errors.New("tdx client")
)

type TDX interface {
	GetReport(data []byte) (report []byte, err error)
}

type TDXClient struct {
	tsm controllers.TSMController
}

func NewTDXClient() (*TDXClient, error) {
	tsm, err := controllers.NewTSM()
	if err != nil {
		return nil, fmt.Errorf("%w: making tsm controller: %w", ErrTDXClient, err)
	}
	return NewTDXClientWithTSM(tsm)
}

func NewTDXClientWithTSM(tsm controllers.TSMController) (*TDXClient, error) {
	return &TDXClient{tsm: tsm}, nil
}

func (t *TDXClient) GetReport(data []byte) (report []byte, err error) {
	result, err := t.tsm.GetReport(controllers.WithTSMReportInBlob(data))
	switch {
	case err != nil:
		return nil, fmt.Errorf("%w: getting report: %w", ErrTDXClient, err)
	case !strings.Contains(result.Provider, TDXProvider):
		return nil, fmt.Errorf("%w: unexpected provider '%s'", ErrTDXClient, result.Provider)
	}
	return result.OutBlob, nil
}
