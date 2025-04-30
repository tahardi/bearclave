package tdx

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/mdlayher/vsock"
)

// TODO: This might be 2 and 3 on other systems. In fact, I think it's:
//   - 0 for hypervisor
//   - 1 is reserved
//   - 2 for any process running on the host (i.e., nonclave)
//   - 3 and above. You can sometimes control the CID when you start the VM
//
// TODO: You might be able to determine the CID in the enclave program at
// runtime with IOCTL_VM_SOCKETS_GET_LOCAL_CID and then pass that in to
// this module when initializing the Communicator

// NonclaveCID In AWS Nitro the "nonclave" program runs on the host (i.e.,
// the parent EC2 instance), which, according to documentation, is always 3.
const NonclaveCID = 2

// EnclaveCID In AWS Nitro the "enclave" program runs on the guest (i.e., the
// VM), which can be any value between 4 and 1023. We use 4 here because it's
// the default value for the `cid` argument to `nitro-cli run-enclave`.
const EnclaveCID = 3

type Communicator struct {
	sendContextID   uint32
	sendPort        uint32
	receiveListener *vsock.Listener
}

func NewCommunicator(
	sendContextID int,
	sendPort int,
	receivePort int,
) (*Communicator, error) {
	receiveListener, err := vsock.Listen(uint32(receivePort), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to set up vsock listener: %w", err)
	}

	return &Communicator{
		sendContextID:   uint32(sendContextID),
		sendPort:        uint32(sendPort),
		receiveListener: receiveListener,
	}, nil
}

func (c *Communicator) Close() error {
	return fmt.Errorf("not implemented")
}

func (c *Communicator) Send(ctx context.Context, data []byte) error {
	errChan := make(chan error, 1)
	go func() {
		conn, err := vsock.Dial(c.sendContextID, c.sendPort, nil)
		if err != nil {
			errChan <- fmt.Errorf("failed to connect to %v: %w", c.sendPort, err)
			return
		}
		defer conn.Close()

		base64Data := base64.StdEncoding.EncodeToString(data)

		n, writeErr := conn.Write([]byte(base64Data))
		switch {
		case writeErr != nil:
			errChan <- fmt.Errorf("failed to write data: %w", writeErr)
		case n != len(base64Data):
			errChan <- fmt.Errorf("failed to write all data: %w", writeErr)
		default:
			errChan <- nil
		}
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("send context cancelled: %w", ctx.Err())
	case err := <-errChan:
		return err
	}
}

func (c *Communicator) Receive(ctx context.Context) ([]byte, error) {
	dataChan := make(chan []byte, 1)
	errChan := make(chan error, 1)
	go func() {
		conn, err := c.receiveListener.Accept()
		if err != nil {
			errChan <- fmt.Errorf("failed to accept connection: %w", err)
			return
		}
		defer conn.Close()

		buf := make([]byte, 10000)
		n, readErr := conn.Read(buf)
		if readErr != nil {
			errChan <- fmt.Errorf("failed to read data: %w", readErr)
			return
		}

		base64Data := buf[:n]
		data, err := base64.StdEncoding.DecodeString(string(base64Data))
		if err != nil {
			errChan <- fmt.Errorf("failed to decode data: %w", err)
			return
		}
		dataChan <- data
	}()

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("receive context cancelled: %w", ctx.Err())
	case err := <-errChan:
		return nil, err
	case data := <-dataChan:
		return data, nil
	}
}
