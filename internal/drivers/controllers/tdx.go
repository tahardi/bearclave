package controllers

import (
	"errors"
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

const (
	TDXDevFile               = "/dev/tdx-guest"
	TDXIOControlMagic        = uint('T')
	TDXReportIOControlNumber = 1
	TDXReportDataLen         = 64
	TDXReportLen             = 1014
)

var (
	ErrTDXController     = errors.New("tdx controller")
	ErrReportDataTooLong = fmt.Errorf("%w: report data too long", ErrTDXController)
	ErrReportDataNil     = fmt.Errorf("%w: report data cannot be nil", ErrTDXController)
)

// TDXReportReq mirrors the Linux kernel's struct tdx_report_req from
// include/uapi/linux/tdx-guest.h. The kernel driver expects this exact
// structure and will copy the resulting report back into the request struct.
// Since there is only one TDCALL function (i.e., TDG.MR.REPORT) that the Linux
// Kernel exposes to guest applications, our message structure is specific to
// reports, unlike the general NSMMessage we see in AWS Nitro Enclaves.
//
// https://github.com/torvalds/linux/blob/e0d4140e804380ae898da1e4c58c21e6323415a4/include/uapi/linux/tdx-guest.h
type TDXReportReq struct {
	ReportData [TDXReportDataLen]uint8
	Report     [TDXReportLen]uint8
}

type TDXController struct {
	file *os.File
}

func NewTDXController() (*TDXController, error) {
	file, err := os.Open(TDXDevFile)
	if err != nil {
		return nil, err
	}
	return NewTDXControllerWithFile(file)
}

func NewTDXControllerWithFile(file *os.File) (*TDXController, error) {
	return &TDXController{file: file}, nil
}

func (n *TDXController) Close() error {
	return n.file.Close()
}

func (n *TDXController) Send(request []byte) ([]byte, error) {
	switch {
	case request == nil:
		return nil, ErrReportDataNil
	case len(request) > TDXReportDataLen:
		return nil, fmt.Errorf(
			"%w: max '%d' got '%d' bytes",
			ErrReportDataTooLong,
			TDXReportDataLen,
			len(request),
		)
	}

	message := TDXReportReq{}
	copy(message.ReportData[:], request)

	command := MakeIOControlCommand(
		IOControlRead|IOControlWrite,
		TDXIOControlMagic,
		TDXReportIOControlNumber,
		uint(unsafe.Sizeof(message)),
	)

	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		n.file.Fd(),
		uintptr(command),
		uintptr(unsafe.Pointer(&message)),
	)
	if errno != 0 {
		return nil, fmt.Errorf("%w: making syscall: %w", ErrTDXController, errno)
	}
	return message.Report[:], nil
}
