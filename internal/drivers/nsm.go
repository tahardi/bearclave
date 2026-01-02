package drivers

import (
	"errors"
	"fmt"
	"io"

	"github.com/fxamacker/cbor/v2"
	"github.com/tahardi/bearclave/internal/drivers/controllers"
)

const (
	// https://github.com/aws/aws-nitro-enclaves-nsm-api/blob/8ec7eac72bbb2097f1058ee32c13e1ff232f13e8/src/driver/mod.rs#L29
	NSMRequestMaxSize = 0x1000

	// https://github.com/aws/aws-nitro-enclaves-nsm-api/blob/main/src/api/mod.rs#L82
	NSMDescribePCR    = "DescribePCR"
	NSMExtendPCR      = "ExtendPCR"
	NSMLockPCR        = "LockPCR"
	NSMLockPCRs       = "LockPCRs"
	NSMGetAttestation = "Attestation"
	NSMGetDescription = "DescribeNSM"
	NSMGetRandom      = "GetRandom"

	// https://github.com/aws/aws-nitro-enclaves-nsm-api/blob/8ec7eac72bbb2097f1058ee32c13e1ff232f13e8/src/api/mod.rs#L49
	NSMDeviceErrorSuccess          = "Success"
	NSMDeviceErrorInvalidArgument  = "InvalidArgument"
	NSMDeviceErrorInvalidIndex     = "InvalidIndex"
	NSMDeviceErrorInvalidResponse  = "InvalidResponse"
	NSMDeviceErrorReadOnlyIndex    = "ReadOnlyIndex"
	NSMDeviceErrorInvalidOperation = "InvalidOperation"
	NSMDeviceErrorBufferTooSmall   = "BufferTooSmall"
	NSMDeviceErrorInputTooLarge    = "InputTooLarge"
	NSMDeviceErrorInternalError    = "InternalError"
)

var (
	ErrNSMClient                 = errors.New("nsm client")
	ErrNSMDevice                 = errors.New("nsm device")
	ErrNSMDeviceInvalidArgument  = fmt.Errorf("%w: %s", ErrNSMDevice, NSMDeviceErrorInvalidArgument)
	ErrNSMDeviceInvalidIndex     = fmt.Errorf("%w: %s", ErrNSMDevice, NSMDeviceErrorInvalidIndex)
	ErrNSMDeviceInvalidResponse  = fmt.Errorf("%w: %s", ErrNSMDevice, NSMDeviceErrorInvalidResponse)
	ErrNSMDeviceReadOnlyIndex    = fmt.Errorf("%w: %s", ErrNSMDevice, NSMDeviceErrorReadOnlyIndex)
	ErrNSMDeviceInvalidOperation = fmt.Errorf("%w: %s", ErrNSMDevice, NSMDeviceErrorInvalidOperation)
	ErrNSMDeviceBufferTooSmall   = fmt.Errorf("%w: %s", ErrNSMDevice, NSMDeviceErrorBufferTooSmall)
	ErrNSMDeviceInputTooLarge    = fmt.Errorf("%w: %s", ErrNSMDevice, NSMDeviceErrorInputTooLarge)
	ErrNSMDeviceInternalError    = fmt.Errorf("%w: %s", ErrNSMDevice, NSMDeviceErrorInternalError)
)

type NSM interface {
	io.Closer

	DescribePCR(index uint16) (pcr []byte, lock bool, err error)
	ExtendPCR(index uint16, data []byte) (pcr []byte, err error)
	LockPCR(index uint16) (err error)
	LockPCRs(end uint16) (err error)
	GetAttestation(nonce []byte, publicKey []byte, userData []byte) (attestation []byte, err error)
	GetDescription() (description *NSMDescription, err error)
	GetRandom(length uint16) (random []byte, err error)
}

type NSMClient struct {
	ioctrl controllers.IOController
}

func NewNSMClient() (*NSMClient, error) {
	ioctrl, err := controllers.NewNSMController()
	if err != nil {
		return nil, err
	}
	return NewNSMClientWithController(ioctrl)
}

func NewNSMClientWithController(
	ioctrl controllers.IOController,
) (*NSMClient, error) {
	return &NSMClient{ioctrl: ioctrl}, nil
}

func (n *NSMClient) Close() error {
	return n.ioctrl.Close()
}

type DescribePCRRequest struct {
	Index uint16 `cbor:"index"`
}

type DescribePCRResponse struct {
	Lock bool   `cbor:"lock"`
	Data []byte `cbor:"data"`
}

func (n *NSMClient) DescribePCR(index uint16) ([]byte, bool, error) {
	req := &DescribePCRRequest{Index: index}
	reqBytes, err := MarshalSerdeCBOR(NSMDescribePCR, req)
	if err != nil {
		return nil, false, fmt.Errorf(
			"%w: marshaling describe pcr request: %w",
			ErrNSMClient,
			err,
		)
	}

	respBytes, err := n.ioctrl.Send(reqBytes)
	if err != nil {
		return nil, false, fmt.Errorf(
			"%w: sending describe pcr request: %w",
			ErrNSMClient,
			err,
		)
	}

	resp := &DescribePCRResponse{}
	err = UnmarshalSerdeResponse(NSMDescribePCR, respBytes, resp)
	switch {
	case err != nil:
		return nil, false, fmt.Errorf(
			"%w: unmarshaling describe pcr response '%x': %w",
			ErrNSMClient,
			respBytes,
			err,
		)
	case len(resp.Data) == 0:
		return nil, false, fmt.Errorf(
			"%w: missing value for pcr: %d",
			ErrNSMClient,
			index,
		)
	}
	return resp.Data, resp.Lock, nil
}

type ExtendPCRRequest struct {
	Index uint16 `cbor:"index"`
	Data  []byte `cbor:"data"`
}

type ExtendPCRResponse struct {
	Data []byte `cbor:"data"`
}

func (n *NSMClient) ExtendPCR(index uint16, data []byte) ([]byte, error) {
	req := &ExtendPCRRequest{Index: index, Data: data}
	reqBytes, err := MarshalSerdeCBOR(NSMExtendPCR, req)
	if err != nil {
		return nil, fmt.Errorf(
			"%w: marshaling extend pcr request: %w",
			ErrNSMClient,
			err,
		)
	}

	respBytes, err := n.ioctrl.Send(reqBytes)
	if err != nil {
		return nil, fmt.Errorf(
			"%w: sending extend pcr request: %w",
			ErrNSMClient,
			err,
		)
	}

	resp := &ExtendPCRResponse{}
	err = UnmarshalSerdeResponse(NSMExtendPCR, respBytes, resp)
	switch {
	case err != nil:
		return nil, fmt.Errorf(
			"%w: unmarshaling extend pcr response '%x': %w",
			ErrNSMClient,
			respBytes,
			err,
		)
	case len(resp.Data) == 0:
		return nil, fmt.Errorf(
			"%w: missing value for pcr: %x",
			ErrNSMClient,
			respBytes,
		)
	}
	return resp.Data, nil
}

type LockPCRRequest struct {
	Index uint16 `cbor:"index"`
}

func (n *NSMClient) LockPCR(index uint16) error {
	req := &LockPCRRequest{Index: index}
	reqBytes, err := MarshalSerdeCBOR(NSMLockPCR, req)
	if err != nil {
		return fmt.Errorf(
			"%w: marshaling lock pcr request: %w",
			ErrNSMClient,
			err,
		)
	}

	respBytes, err := n.ioctrl.Send(reqBytes)
	if err != nil {
		return fmt.Errorf(
			"%w: sending lock pcr request: %w",
			ErrNSMClient,
			err,
		)
	}

	resp := ""
	err = UnmarshalResponse(respBytes, &resp)
	switch {
	case err != nil:
		return fmt.Errorf(
			"%w: unmarshaling lock pcr response '%x': %w",
			ErrNSMClient,
			respBytes,
			err,
		)
	case resp != NSMLockPCR:
		return fmt.Errorf(
			"%w: invalid lock pcr response: %x",
			ErrNSMClient,
			respBytes,
		)
	}
	return nil
}

type LockPCRsRequest struct {
	Range uint16 `cbor:"range"`
}

// LockPCRs locks all PCRs from 0 to end (exclusive).
func (n *NSMClient) LockPCRs(end uint16) error {
	req := &LockPCRsRequest{Range: end}
	reqBytes, err := MarshalSerdeCBOR(NSMLockPCRs, req)
	if err != nil {
		return fmt.Errorf(
			"%w: marshaling lock pcrs request: %w",
			ErrNSMClient,
			err,
		)
	}

	respBytes, err := n.ioctrl.Send(reqBytes)
	if err != nil {
		return fmt.Errorf(
			"%w: sending lock pcrs request: %w",
			ErrNSMClient,
			err,
		)
	}

	resp := ""
	err = UnmarshalResponse(respBytes, &resp)
	switch {
	case err != nil:
		return fmt.Errorf(
			"%w: unmarshaling lock pcrs response '%x': %w",
			ErrNSMClient,
			respBytes,
			err,
		)
	case resp != NSMLockPCRs:
		return fmt.Errorf(
			"%w: invalid lock pcrs response: %x",
			ErrNSMClient,
			respBytes,
		)
	}
	return nil
}

type GetAttestationRequest struct {
	Nonce     []byte `cbor:"nonce"`
	PublicKey []byte `cbor:"public_key"`
	UserData  []byte `cbor:"user_data"`
}

type GetAttestationResponse struct {
	Document []byte `cbor:"document"`
}

func (n *NSMClient) GetAttestation(
	nonce []byte,
	publicKey []byte,
	userData []byte,
) ([]byte, error) {
	req := &GetAttestationRequest{Nonce: nonce, PublicKey: publicKey, UserData: userData}
	reqBytes, err := MarshalSerdeCBOR(NSMGetAttestation, req)
	if err != nil {
		return nil, fmt.Errorf(
			"%w: marshaling get attestation request: %w",
			ErrNSMClient,
			err,
		)
	}

	respBytes, err := n.ioctrl.Send(reqBytes)
	if err != nil {
		return nil, fmt.Errorf(
			"%w: sending get attestation request: %w",
			ErrNSMClient,
			err,
		)
	}

	resp := &GetAttestationResponse{}
	err = UnmarshalSerdeResponse(NSMGetAttestation, respBytes, resp)
	switch {
	case err != nil:
		return nil, fmt.Errorf(
			"%w: unmarshaling get attestation response '%x': %w",
			ErrNSMClient,
			respBytes,
			err,
		)
	case len(resp.Document) == 0:
		return nil, fmt.Errorf(
			"%w: missing attestation: %x",
			ErrNSMClient,
			respBytes,
		)
	}
	return resp.Document, nil
}

type NSMDescription struct {
	VersionMajor uint16   `cbor:"version_major" json:"version_major,omitempty"`
	VersionMinor uint16   `cbor:"version_minor" json:"version_minor,omitempty"`
	VersionPatch uint16   `cbor:"version_patch" json:"version_patch,omitempty"`
	ModuleID     string   `cbor:"module_id"     json:"module_id,omitempty"`
	MaxPCRs      uint16   `cbor:"max_pcrs"      json:"max_pcrs,omitempty"`
	LockedPCRs   []uint16 `cbor:"locked_pcrs"   json:"locked_pcrs,omitempty"`
	Digest       string   `cbor:"digest"        json:"digest,omitempty"`
}

func (n *NSMClient) GetDescription() (*NSMDescription, error) {
	reqBytes, err := cbor.Marshal(NSMGetDescription)
	if err != nil {
		return nil, fmt.Errorf(
			"%w: marshaling get description request: %w",
			ErrNSMClient,
			err,
		)
	}

	respBytes, err := n.ioctrl.Send(reqBytes)
	if err != nil {
		return nil, fmt.Errorf(
			"%w: sending get description request: %w",
			ErrNSMClient,
			err,
		)
	}

	resp := &NSMDescription{}
	err = UnmarshalSerdeResponse(NSMGetDescription, respBytes, resp)
	switch {
	case err != nil:
		return nil, fmt.Errorf(
			"%w: unmarshaling get description response '%x': %w",
			ErrNSMClient,
			respBytes,
			err,
		)
	case resp.Digest == "" || resp.ModuleID == "":
		return nil, fmt.Errorf(
			"%w: missing description: %x",
			ErrNSMClient,
			respBytes,
		)
	}
	return resp, nil
}

type GetRandomResponse struct {
	Random []byte `cbor:"random"`
}

func (n *NSMClient) GetRandom(length uint16) ([]byte, error) {
	random := make([]byte, length)
	if length == 0 {
		return random, nil
	}

	numBytes := 0
	for numBytes < int(length) {
		bytes, err := n.getRandom()
		if err != nil {
			return nil, err
		}
		numBytes += copy(random[numBytes:], bytes)
	}
	return random, nil
}

func (n *NSMClient) getRandom() ([]byte, error) {
	reqBytes, err := cbor.Marshal(NSMGetRandom)
	if err != nil {
		return nil, fmt.Errorf(
			"%w: marshaling get random request: %w",
			ErrNSMClient,
			err,
		)
	}

	respBytes, err := n.ioctrl.Send(reqBytes)
	if err != nil {
		return nil, fmt.Errorf(
			"%w: sending get random request: %w",
			ErrNSMClient,
			err,
		)
	}

	resp := &GetRandomResponse{}
	err = UnmarshalSerdeResponse(NSMGetRandom, respBytes, resp)
	switch {
	case err != nil:
		return nil, fmt.Errorf(
			"%w: unmarshaling get random response '%x': %w",
			ErrNSMClient,
			respBytes,
			err,
		)
	case len(resp.Random) == 0:
		return nil, fmt.Errorf(
			"%w: missing random bytes: %x",
			ErrNSMClient,
			respBytes,
		)
	}
	return resp.Random, nil
}

type NSMDeviceError struct {
	Error string `cbor:"error"`
}

func UnmarshalNSMDeviceError(data []byte) error {
	nsmError := &NSMDeviceError{Error: ""}
	_ = cbor.Unmarshal(data, nsmError)
	switch nsmError.Error {
	case "", NSMDeviceErrorSuccess:
		return nil
	case NSMDeviceErrorInvalidArgument:
		return ErrNSMDeviceInvalidArgument
	case NSMDeviceErrorInvalidIndex:
		return ErrNSMDeviceInvalidIndex
	case NSMDeviceErrorInvalidResponse:
		return ErrNSMDeviceInvalidResponse
	case NSMDeviceErrorReadOnlyIndex:
		return ErrNSMDeviceReadOnlyIndex
	case NSMDeviceErrorInvalidOperation:
		return ErrNSMDeviceInvalidOperation
	case NSMDeviceErrorBufferTooSmall:
		return ErrNSMDeviceBufferTooSmall
	case NSMDeviceErrorInputTooLarge:
		return ErrNSMDeviceInputTooLarge
	case NSMDeviceErrorInternalError:
		return ErrNSMDeviceInternalError
	default:
		return fmt.Errorf("%w: %s", ErrNSMDevice, nsmError.Error)
	}
}

// UnmarshalSerdeResponse our NSMClient.X functions assume that they will
// get the proper response type for a given request (e.g., GetRandomRequest
// returns GetRandomResponse). The NSM device will return an error, however,
// if the operation failed. We use these convenience functions to first
// check if we got an NSMDeviceError. If not, we try to unmarshal to the
// expected response type.
func UnmarshalSerdeResponse[T any](key string, data []byte, value *T) error {
	err := UnmarshalNSMDeviceError(data)
	if err != nil {
		return err
	}
	return UnmarshalSerdeCBOR(key, data, value)
}

func UnmarshalResponse[T any](data []byte, value *T) error {
	err := UnmarshalNSMDeviceError(data)
	if err != nil {
		return err
	}
	return cbor.Unmarshal(data, value)
}

func MarshalSerdeCBOR[T any](key string, value T) ([]byte, error) {
	serde := map[string]T{key: value}
	return cbor.Marshal(serde)
}

func UnmarshalSerdeCBOR[T any](key string, data []byte, value *T) error {
	serde := map[string]T{}
	err := cbor.Unmarshal(data, &serde)
	if err != nil {
		return err
	}

	result, ok := serde[key]
	if !ok {
		return fmt.Errorf("%w: missing value for key: %s", ErrNSMClient, key)
	}
	*value = result
	return nil
}
