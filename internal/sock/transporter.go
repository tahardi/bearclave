package sock

import (
	"context"
	"encoding/base64"
	"fmt"
	"net"
)

type Transporter struct {
	receiveListener net.Listener
	sendAddr        string
}

func NewTransporter(sendPort int, receivePort int) (*Transporter, error) {
	sendAddr := fmt.Sprintf("127.0.0.1:%d", sendPort)
	receiveAddr := fmt.Sprintf("127.0.0.1:%d", receivePort)
	//sendAddr := fmt.Sprintf("0.0.0.0:%d", sendPort)
	//receiveAddr := fmt.Sprintf("0.0.0.0:%d", receivePort)
	receiveListener, err := net.Listen("tcp", receiveAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to set up TCP listener on %s: %w", receiveAddr, err)
	}

	return &Transporter{
		sendAddr:        sendAddr,
		receiveListener: receiveListener,
	}, nil
}

func (c *Transporter) Close() error {
	if c.receiveListener != nil {
		c.receiveListener.Close()
	}
	return nil
}

func (c *Transporter) Send(ctx context.Context, data []byte) error {
	errChan := make(chan error, 1)
	go func() {
		conn, err := net.Dial("tcp", c.sendAddr)
		if err != nil {
			errChan <- fmt.Errorf("failed to connect to %s: %w", c.sendAddr, err)
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

func (c *Transporter) Receive(ctx context.Context) ([]byte, error) {
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
