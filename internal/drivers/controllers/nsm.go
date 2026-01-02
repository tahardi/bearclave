package controllers

import (
	"errors"
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

const (
	// https://github.com/aws/aws-nitro-enclaves-nsm-api/blob/8ec7eac72bbb2097f1058ee32c13e1ff232f13e8/src/driver/mod.rs#L12
	NSMDevFile         = "/dev/nsm"
	NSMIOControlMagic  = 0x0A
	NSMIOControlNumber = 0
	NSMMaxRequestSize  = 0x1000
	NSMMaxResponseSize = 0x3000
)

var (
	ErrNSMController = errors.New("nsm controller")
)

// NSMMessage structure taken from the aws-nitro-enclaves-nsm-api rust library:
// https://github.com/aws/aws-nitro-enclaves-nsm-api/blob/8ec7eac72bbb2097f1058ee32c13e1ff232f13e8/src/driver/mod.rs#L34
type NSMMessage struct {
	Request  syscall.Iovec
	Response syscall.Iovec
}

type NSMController struct {
	file *os.File
}

func NewNSMController() (*NSMController, error) {
	file, err := os.Open(NSMDevFile)
	if err != nil {
		return nil, err
	}
	return NewNSMControllerWithFile(file)
}

func NewNSMControllerWithFile(file *os.File) (*NSMController, error) {
	return &NSMController{file: file}, nil
}

func (n *NSMController) Close() error {
	return n.file.Close()
}

func (n *NSMController) Send(request []byte) ([]byte, error) {
	if len(request) < 1 || len(request) > NSMMaxRequestSize {
		return nil, fmt.Errorf(
			"%w: invalid request size: %d",
			ErrNSMController,
			len(request),
		)
	}

	response := make([]byte, NSMMaxResponseSize)
	message := NSMMessage{
		Request: syscall.Iovec{
			Base: &request[0],
		},
		Response: syscall.Iovec{
			Base: &response[0],
		},
	}
	message.Request.SetLen(len(request))
	message.Response.SetLen(len(response))

	// The direction, magic, and number are taken from the
	// aws-nitro-enclaves-nsm-api rust library:
	// https://github.com/aws/aws-nitro-enclaves-nsm-api/blob/8ec7eac72bbb2097f1058ee32c13e1ff232f13e8/src/driver/mod.rs#L66
	command := MakeIOControlCommand(
		IOControlRead|IOControlWrite,
		NSMIOControlMagic,
		NSMIOControlNumber,
		uint(unsafe.Sizeof(message)),
	)

	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		n.file.Fd(),
		uintptr(command),
		uintptr(unsafe.Pointer(&message)),
	)
	if errno != 0 {
		return nil, fmt.Errorf("%w: making syscall: %w", ErrNSMController, errno)
	}
	return response[:message.Response.Len], nil
}
