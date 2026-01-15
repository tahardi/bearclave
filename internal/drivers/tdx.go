package drivers

import (
	"errors"
	"fmt"
	"io"

	"github.com/tahardi/bearclave/internal/drivers/controllers"
)

var (
	ErrTDXClient = errors.New("tdx client")
	ErrEmptyReport = fmt.Errorf("%w: received empty report from tdx device", ErrTDXClient)
)

type TDX interface {
	io.Closer

	GetReport0(reportData []byte) (report []byte, err error)
}

type TDXClient struct {
	ioctrl controllers.IOController
}

func NewTDXClient() (*TDXClient, error) {
	ioctrl, err := controllers.NewTDXController()
	if err != nil {
		return nil, err
	}
	return NewTDXClientWithController(ioctrl)
}

func NewTDXClientWithController(
	ioctrl controllers.IOController,
) (*TDXClient, error) {
	return &TDXClient{ioctrl: ioctrl}, nil
}

func (n *TDXClient) Close() error {
	return n.ioctrl.Close()
}

func (t *TDXClient) GetReport0(reportData []byte) ([]byte, error) {
	if reportData == nil {
		reportData = []byte{}
	}

	report, err := t.ioctrl.Send(reportData)
	if err != nil {
		return nil, fmt.Errorf("%w: sending report request: %w", ErrTDXClient, err)
	}

	if len(report) == 0 {
		return nil, ErrEmptyReport
	}
	return report, nil
}
