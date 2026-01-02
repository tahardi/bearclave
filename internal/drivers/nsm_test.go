package drivers_test

import (
	"testing"

	"github.com/fxamacker/cbor/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tahardi/bearclave/internal/drivers"
	"github.com/tahardi/bearclave/mocks"
)

func unmarshalCBORFromMockArgs[T any](
	t *testing.T,
	args mock.Arguments,
	value *T,
) {
	t.Helper()
	data, ok := args.Get(0).([]byte)
	require.True(t, ok)
	err := cbor.Unmarshal(data, value)
	require.NoError(t, err)
}

func unmarshalSerdeCBORFromMockArgs[T any](
	t *testing.T,
	args mock.Arguments,
	key string,
	value *T,
) {
	t.Helper()
	data, ok := args.Get(0).([]byte)
	require.True(t, ok)
	err := drivers.UnmarshalSerdeCBOR(key, data, value)
	require.NoError(t, err)
}

func TestNSMClient_Interfaces(t *testing.T) {
	t.Run("NSM", func(_ *testing.T) {
		var _ drivers.NSM = &drivers.NSMClient{}
	})
}

func TestNSMClient_DescribePCR(t *testing.T) {
	wantIndex := uint16(1)
	wantLock := true
	wantData := []byte("data")
	resp := &drivers.DescribePCRResponse{Lock: wantLock, Data: wantData}
	respBytes, err := drivers.MarshalSerdeCBOR(drivers.NSMDescribePCR, resp)
	require.NoError(t, err)

	t.Run("happy path", func(t *testing.T) {
		// given
		ioctl := mocks.NewIOController(t)
		ioctl.On("Send", mock.Anything).
			Run(func(args mock.Arguments) {
				req := &drivers.DescribePCRRequest{}
				unmarshalSerdeCBORFromMockArgs(t, args, drivers.NSMDescribePCR, req)
				assert.Equal(t, wantIndex, req.Index)
			}).
			Return(respBytes, nil)

		client, err := drivers.NewNSMClientWithController(ioctl)
		require.NoError(t, err)

		// when
		gotData, gotLock, err := client.DescribePCR(wantIndex)

		// then
		require.NoError(t, err)
		assert.Equal(t, wantLock, gotLock)
		assert.Equal(t, wantData, gotData)
	})

	t.Run("error - ioctl", func(t *testing.T) {
		// given
		ioctl := mocks.NewIOController(t)
		ioctl.On("Send", mock.Anything).Return(nil, assert.AnError)

		client, err := drivers.NewNSMClientWithController(ioctl)
		require.NoError(t, err)

		// when
		_, _, err = client.DescribePCR(wantIndex)

		// then
		require.Error(t, err)
		require.ErrorIs(t, err, assert.AnError)
	})

	t.Run("error - missing response", func(t *testing.T) {
		// given
		missingBytes, err := drivers.MarshalSerdeCBOR("wrongresponsetype", resp)
		require.NoError(t, err)

		ioctl := mocks.NewIOController(t)
		ioctl.On("Send", mock.Anything).
			Run(func(args mock.Arguments) {
				req := &drivers.DescribePCRRequest{}
				unmarshalSerdeCBORFromMockArgs(t, args, drivers.NSMDescribePCR, req)
				assert.Equal(t, wantIndex, req.Index)
			}).
			Return(missingBytes, nil)

		client, err := drivers.NewNSMClientWithController(ioctl)
		require.NoError(t, err)

		// when
		_, _, err = client.DescribePCR(wantIndex)

		// then
		require.Error(t, err)
		require.ErrorIs(t, err, drivers.ErrNSMClient)
		assert.ErrorContains(t, err, "missing value for key")
	})

	t.Run("error - missing pcr value", func(t *testing.T) {
		// given
		missingResp := &drivers.DescribePCRResponse{Lock: true, Data: nil}
		missingBytes, err := drivers.MarshalSerdeCBOR(drivers.NSMDescribePCR, missingResp)
		require.NoError(t, err)

		ioctl := mocks.NewIOController(t)
		ioctl.On("Send", mock.Anything).
			Run(func(args mock.Arguments) {
				req := &drivers.DescribePCRRequest{}
				unmarshalSerdeCBORFromMockArgs(t, args, drivers.NSMDescribePCR, req)
				assert.Equal(t, wantIndex, req.Index)
			}).
			Return(missingBytes, nil)

		client, err := drivers.NewNSMClientWithController(ioctl)
		require.NoError(t, err)

		// when
		_, _, err = client.DescribePCR(wantIndex)

		// then
		require.Error(t, err)
		require.ErrorIs(t, err, drivers.ErrNSMClient)
		assert.ErrorContains(t, err, "missing value for pcr")
	})
}

func TestNSMClient_ExtendPCR(t *testing.T) {
	wantIndex := uint16(1)
	wantData := []byte("data")
	resp := &drivers.ExtendPCRResponse{Data: wantData}
	respBytes, err := drivers.MarshalSerdeCBOR(drivers.NSMExtendPCR, resp)
	require.NoError(t, err)

	t.Run("happy path", func(t *testing.T) {
		// given
		ioctl := mocks.NewIOController(t)
		ioctl.On("Send", mock.Anything).
			Run(func(args mock.Arguments) {
				req := &drivers.ExtendPCRRequest{}
				unmarshalSerdeCBORFromMockArgs(t, args, drivers.NSMExtendPCR, req)
				assert.Equal(t, wantIndex, req.Index)
				assert.Equal(t, wantData, req.Data)
			}).
			Return(respBytes, nil)

		client, err := drivers.NewNSMClientWithController(ioctl)
		require.NoError(t, err)

		// when
		gotData, err := client.ExtendPCR(wantIndex, wantData)

		// then
		require.NoError(t, err)
		assert.Equal(t, wantData, gotData)
	})

	t.Run("error - ioctl", func(t *testing.T) {
		// given
		ioctl := mocks.NewIOController(t)
		ioctl.On("Send", mock.Anything).Return(nil, assert.AnError)

		client, err := drivers.NewNSMClientWithController(ioctl)
		require.NoError(t, err)

		// when
		_, err = client.ExtendPCR(wantIndex, wantData)

		// then
		require.Error(t, err)
		require.ErrorIs(t, err, assert.AnError)
	})

	t.Run("error - missing response", func(t *testing.T) {
		// given
		missingBytes, err := drivers.MarshalSerdeCBOR("wrongresponsetype", resp)
		require.NoError(t, err)

		ioctl := mocks.NewIOController(t)
		ioctl.On("Send", mock.Anything).
			Run(func(args mock.Arguments) {
				req := &drivers.ExtendPCRRequest{}
				unmarshalSerdeCBORFromMockArgs(t, args, drivers.NSMExtendPCR, req)
				assert.Equal(t, wantIndex, req.Index)
				assert.Equal(t, wantData, req.Data)
			}).
			Return(missingBytes, nil)

		client, err := drivers.NewNSMClientWithController(ioctl)
		require.NoError(t, err)

		// when
		_, err = client.ExtendPCR(wantIndex, wantData)

		// then
		require.Error(t, err)
		require.ErrorIs(t, err, drivers.ErrNSMClient)
		assert.ErrorContains(t, err, "missing value for key")
	})

	t.Run("error - missing pcr value", func(t *testing.T) {
		// given
		missingResp := &drivers.ExtendPCRResponse{Data: nil}
		missingBytes, err := drivers.MarshalSerdeCBOR(drivers.NSMExtendPCR, missingResp)
		require.NoError(t, err)

		ioctl := mocks.NewIOController(t)
		ioctl.On("Send", mock.Anything).
			Run(func(args mock.Arguments) {
				req := &drivers.ExtendPCRRequest{}
				unmarshalSerdeCBORFromMockArgs(t, args, drivers.NSMExtendPCR, req)
				assert.Equal(t, wantIndex, req.Index)
				assert.Equal(t, wantData, req.Data)
			}).
			Return(missingBytes, nil)

		client, err := drivers.NewNSMClientWithController(ioctl)
		require.NoError(t, err)

		// when
		_, err = client.ExtendPCR(wantIndex, wantData)

		// then
		require.Error(t, err)
		require.ErrorIs(t, err, drivers.ErrNSMClient)
		assert.ErrorContains(t, err, "missing value for pcr")
	})
}

func TestNSMClient_LockPCR(t *testing.T) {
	wantIndex := uint16(1)
	respBytes, err := cbor.Marshal(drivers.NSMLockPCR)
	require.NoError(t, err)

	t.Run("happy path", func(t *testing.T) {
		// given
		ioctl := mocks.NewIOController(t)
		ioctl.On("Send", mock.Anything).
			Run(func(args mock.Arguments) {
				req := &drivers.LockPCRRequest{}
				unmarshalSerdeCBORFromMockArgs(t, args, drivers.NSMLockPCR, req)
				assert.Equal(t, wantIndex, req.Index)
			}).
			Return(respBytes, nil)

		client, err := drivers.NewNSMClientWithController(ioctl)
		require.NoError(t, err)

		// when
		err = client.LockPCR(wantIndex)

		// then
		require.NoError(t, err)
	})

	t.Run("error - ioctl", func(t *testing.T) {
		// given
		ioctl := mocks.NewIOController(t)
		ioctl.On("Send", mock.Anything).Return(nil, assert.AnError)

		client, err := drivers.NewNSMClientWithController(ioctl)
		require.NoError(t, err)

		// when
		err = client.LockPCR(wantIndex)

		// then
		require.Error(t, err)
		require.ErrorIs(t, err, assert.AnError)
	})

	t.Run("error - missing response", func(t *testing.T) {
		// given
		missingBytes, err := cbor.Marshal("wrongresponsetype")
		require.NoError(t, err)

		ioctl := mocks.NewIOController(t)
		ioctl.On("Send", mock.Anything).
			Run(func(args mock.Arguments) {
				req := &drivers.LockPCRRequest{}
				unmarshalSerdeCBORFromMockArgs(t, args, drivers.NSMLockPCR, req)
				assert.Equal(t, wantIndex, req.Index)
			}).
			Return(missingBytes, nil)

		client, err := drivers.NewNSMClientWithController(ioctl)
		require.NoError(t, err)

		// when
		err = client.LockPCR(wantIndex)

		// then
		require.Error(t, err)
		require.ErrorIs(t, err, drivers.ErrNSMClient)
		assert.ErrorContains(t, err, "invalid lock pcr response")
	})
}

func TestNSMClient_LockPCRs(t *testing.T) {
	wantRange := uint16(1)
	respBytes, err := cbor.Marshal(drivers.NSMLockPCRs)
	require.NoError(t, err)

	t.Run("happy path", func(t *testing.T) {
		// given
		ioctl := mocks.NewIOController(t)
		ioctl.On("Send", mock.Anything).
			Run(func(args mock.Arguments) {
				req := &drivers.LockPCRsRequest{}
				unmarshalSerdeCBORFromMockArgs(t, args, drivers.NSMLockPCRs, req)
				assert.Equal(t, wantRange, req.Range)
			}).
			Return(respBytes, nil)

		client, err := drivers.NewNSMClientWithController(ioctl)
		require.NoError(t, err)

		// when
		err = client.LockPCRs(wantRange)

		// then
		require.NoError(t, err)
	})

	t.Run("error - ioctl", func(t *testing.T) {
		// given
		ioctl := mocks.NewIOController(t)
		ioctl.On("Send", mock.Anything).Return(nil, assert.AnError)

		client, err := drivers.NewNSMClientWithController(ioctl)
		require.NoError(t, err)

		// when
		err = client.LockPCRs(wantRange)

		// then
		require.Error(t, err)
		require.ErrorIs(t, err, assert.AnError)
	})

	t.Run("error - missing response", func(t *testing.T) {
		// given
		missingBytes, err := cbor.Marshal("wrongresponsetype")
		require.NoError(t, err)

		ioctl := mocks.NewIOController(t)
		ioctl.On("Send", mock.Anything).
			Run(func(args mock.Arguments) {
				req := &drivers.LockPCRsRequest{}
				unmarshalSerdeCBORFromMockArgs(t, args, drivers.NSMLockPCRs, req)
				assert.Equal(t, wantRange, req.Range)
			}).
			Return(missingBytes, nil)

		client, err := drivers.NewNSMClientWithController(ioctl)
		require.NoError(t, err)

		// when
		err = client.LockPCRs(wantRange)

		// then
		require.Error(t, err)
		require.ErrorIs(t, err, drivers.ErrNSMClient)
		assert.ErrorContains(t, err, "invalid lock pcrs response")
	})
}

func TestNSMClient_GetAttestation(t *testing.T) {
	wantAttestation := []byte("attestation")
	wantNonce := []byte("nonce")
	wantPublicKey := []byte("publickey")
	wantUserData := []byte("userdata")
	resp := &drivers.GetAttestationResponse{Document: wantAttestation}
	respBytes, err := drivers.MarshalSerdeCBOR(drivers.NSMGetAttestation, resp)
	require.NoError(t, err)

	t.Run("happy path", func(t *testing.T) {
		// given
		ioctl := mocks.NewIOController(t)
		ioctl.On("Send", mock.Anything).
			Run(func(args mock.Arguments) {
				req := &drivers.GetAttestationRequest{}
				unmarshalSerdeCBORFromMockArgs(t, args, drivers.NSMGetAttestation, req)
				assert.Equal(t, wantNonce, req.Nonce)
				assert.Equal(t, wantPublicKey, req.PublicKey)
				assert.Equal(t, wantUserData, req.UserData)
			}).
			Return(respBytes, nil)

		client, err := drivers.NewNSMClientWithController(ioctl)
		require.NoError(t, err)

		// when
		gotAttestation, err := client.GetAttestation(
			wantNonce,
			wantPublicKey,
			wantUserData,
		)

		// then
		require.NoError(t, err)
		assert.Equal(t, wantAttestation, gotAttestation)
	})

	t.Run("error - ioctl", func(t *testing.T) {
		// given
		ioctl := mocks.NewIOController(t)
		ioctl.On("Send", mock.Anything).Return(nil, assert.AnError)

		client, err := drivers.NewNSMClientWithController(ioctl)
		require.NoError(t, err)

		// when
		_, err = client.GetAttestation(wantNonce, wantPublicKey, wantUserData)

		// then
		require.Error(t, err)
		require.ErrorIs(t, err, assert.AnError)
	})

	t.Run("error - missing response", func(t *testing.T) {
		// given
		missingBytes, err := drivers.MarshalSerdeCBOR("wrongresponsetype", resp)
		require.NoError(t, err)

		ioctl := mocks.NewIOController(t)
		ioctl.On("Send", mock.Anything).
			Run(func(args mock.Arguments) {
				req := &drivers.GetAttestationRequest{}
				unmarshalSerdeCBORFromMockArgs(t, args, drivers.NSMGetAttestation, req)
				assert.Equal(t, wantNonce, req.Nonce)
				assert.Equal(t, wantPublicKey, req.PublicKey)
				assert.Equal(t, wantUserData, req.UserData)
			}).
			Return(missingBytes, nil)

		client, err := drivers.NewNSMClientWithController(ioctl)
		require.NoError(t, err)

		// when
		_, err = client.GetAttestation(wantNonce, wantPublicKey, wantUserData)

		// then
		require.Error(t, err)
		require.ErrorIs(t, err, drivers.ErrNSMClient)
		assert.ErrorContains(t, err, "missing value for key")
	})

	t.Run("error - missing pcr value", func(t *testing.T) {
		// given
		missingResp := &drivers.GetAttestationResponse{Document: nil}
		missingBytes, err := drivers.MarshalSerdeCBOR(drivers.NSMGetAttestation, missingResp)
		require.NoError(t, err)

		ioctl := mocks.NewIOController(t)
		ioctl.On("Send", mock.Anything).
			Run(func(args mock.Arguments) {
				req := &drivers.GetAttestationRequest{}
				unmarshalSerdeCBORFromMockArgs(t, args, drivers.NSMGetAttestation, req)
				assert.Equal(t, wantNonce, req.Nonce)
				assert.Equal(t, wantPublicKey, req.PublicKey)
				assert.Equal(t, wantUserData, req.UserData)
			}).
			Return(missingBytes, nil)

		client, err := drivers.NewNSMClientWithController(ioctl)
		require.NoError(t, err)

		// when
		_, err = client.GetAttestation(wantNonce, wantPublicKey, wantUserData)

		// then
		require.Error(t, err)
		require.ErrorIs(t, err, drivers.ErrNSMClient)
		assert.ErrorContains(t, err, "missing attestation")
	})
}

func TestNSMClient_GetDescription(t *testing.T) {
	wantDescription := &drivers.NSMDescription{
		VersionMajor: uint16(1),
		VersionMinor: uint16(2),
		VersionPatch: uint16(3),
		ModuleID:     "moduleID",
		MaxPCRs:      uint16(4),
		LockedPCRs:   []uint16{5, 6},
		Digest:       "SHA384",
	}
	respBytes, err := drivers.MarshalSerdeCBOR(drivers.NSMGetDescription, &wantDescription)
	require.NoError(t, err)

	t.Run("happy path", func(t *testing.T) {
		// given
		ioctl := mocks.NewIOController(t)
		ioctl.On("Send", mock.Anything).
			Run(func(args mock.Arguments) {
				req := ""
				unmarshalCBORFromMockArgs(t, args, &req)
				assert.Equal(t, drivers.NSMGetDescription, req)
			}).
			Return(respBytes, nil)

		client, err := drivers.NewNSMClientWithController(ioctl)
		require.NoError(t, err)

		// when
		gotDescription, err := client.GetDescription()

		// then
		require.NoError(t, err)
		assert.Equal(t, wantDescription, gotDescription)
	})

	t.Run("error - ioctl", func(t *testing.T) {
		// given
		ioctl := mocks.NewIOController(t)
		ioctl.On("Send", mock.Anything).Return(nil, assert.AnError)

		client, err := drivers.NewNSMClientWithController(ioctl)
		require.NoError(t, err)

		// when
		_, err = client.GetDescription()

		// then
		require.Error(t, err)
		require.ErrorIs(t, err, assert.AnError)
	})

	t.Run("error - missing response", func(t *testing.T) {
		// given
		missingBytes, err := drivers.MarshalSerdeCBOR("wrongresponsetype", &wantDescription)
		require.NoError(t, err)

		ioctl := mocks.NewIOController(t)
		ioctl.On("Send", mock.Anything).
			Run(func(args mock.Arguments) {
				req := ""
				unmarshalCBORFromMockArgs(t, args, &req)
				assert.Equal(t, drivers.NSMGetDescription, req)
				assert.Equal(t, drivers.NSMGetDescription, req)
			}).
			Return(missingBytes, nil)

		client, err := drivers.NewNSMClientWithController(ioctl)
		require.NoError(t, err)

		// when
		_, err = client.GetDescription()

		// then
		require.Error(t, err)
		require.ErrorIs(t, err, drivers.ErrNSMClient)
		assert.ErrorContains(t, err, "missing value for key")
	})

	t.Run("error - missing description", func(t *testing.T) {
		// given
		missingDesc := &drivers.NSMDescription{
			VersionMajor: uint16(1),
			VersionMinor: uint16(2),
			VersionPatch: uint16(3),
			ModuleID:     "",
			MaxPCRs:      uint16(4),
			LockedPCRs:   []uint16{5, 6},
			Digest:       "",
		}
		missingBytes, err := drivers.MarshalSerdeCBOR(drivers.NSMGetDescription, &missingDesc)
		require.NoError(t, err)

		ioctl := mocks.NewIOController(t)
		ioctl.On("Send", mock.Anything).
			Run(func(args mock.Arguments) {
				req := ""
				unmarshalCBORFromMockArgs(t, args, &req)
				assert.Equal(t, drivers.NSMGetDescription, req)
				assert.Equal(t, drivers.NSMGetDescription, req)
			}).
			Return(missingBytes, nil)

		client, err := drivers.NewNSMClientWithController(ioctl)
		require.NoError(t, err)

		// when
		_, err = client.GetDescription()

		// then
		require.Error(t, err)
		require.ErrorIs(t, err, drivers.ErrNSMClient)
		assert.ErrorContains(t, err, "missing description")
	})
}

func TestNSMClient_GetRandom(t *testing.T) {
	wantLength := uint16(5)
	wantRandom := []byte("aaaaa")
	resp := &drivers.GetRandomResponse{Random: []byte("a")}
	respBytes, err := drivers.MarshalSerdeCBOR(drivers.NSMGetRandom, resp)
	require.NoError(t, err)

	t.Run("happy path", func(t *testing.T) {
		// given
		ioctl := mocks.NewIOController(t)
		ioctl.On("Send", mock.Anything).
			Run(func(args mock.Arguments) {
				req := ""
				unmarshalCBORFromMockArgs(t, args, &req)
				assert.Equal(t, drivers.NSMGetRandom, req)
			}).
			Return(respBytes, nil)

		client, err := drivers.NewNSMClientWithController(ioctl)
		require.NoError(t, err)

		// when
		gotRandom, err := client.GetRandom(wantLength)

		// then
		require.NoError(t, err)
		assert.Equal(t, wantRandom, gotRandom)
	})

	t.Run("error - ioctl", func(t *testing.T) {
		// given
		ioctl := mocks.NewIOController(t)
		ioctl.On("Send", mock.Anything).Return(nil, assert.AnError)

		client, err := drivers.NewNSMClientWithController(ioctl)
		require.NoError(t, err)

		// when
		_, err = client.GetRandom(wantLength)

		// then
		require.Error(t, err)
		require.ErrorIs(t, err, assert.AnError)
	})

	t.Run("error - missing response", func(t *testing.T) {
		// given
		missingBytes, err := drivers.MarshalSerdeCBOR("wrongresponsetype", resp)
		require.NoError(t, err)

		ioctl := mocks.NewIOController(t)
		ioctl.On("Send", mock.Anything).
			Run(func(args mock.Arguments) {
				req := ""
				unmarshalCBORFromMockArgs(t, args, &req)
				assert.Equal(t, drivers.NSMGetRandom, req)
			}).
			Return(missingBytes, nil)

		client, err := drivers.NewNSMClientWithController(ioctl)
		require.NoError(t, err)

		// when
		_, err = client.GetRandom(wantLength)

		// then
		require.Error(t, err)
		require.ErrorIs(t, err, drivers.ErrNSMClient)
		assert.ErrorContains(t, err, "missing value for key")
	})

	t.Run("error - missing random bytes", func(t *testing.T) {
		// given
		missingResp := &drivers.GetRandomResponse{Random: nil}
		MissingBytes, err := drivers.MarshalSerdeCBOR(drivers.NSMGetRandom, missingResp)
		require.NoError(t, err)

		ioctl := mocks.NewIOController(t)
		ioctl.On("Send", mock.Anything).
			Run(func(args mock.Arguments) {
				req := ""
				unmarshalCBORFromMockArgs(t, args, &req)
				assert.Equal(t, drivers.NSMGetRandom, req)
			}).
			Return(MissingBytes, nil)

		client, err := drivers.NewNSMClientWithController(ioctl)
		require.NoError(t, err)

		// when
		_, err = client.GetRandom(wantLength)

		// then
		require.Error(t, err)
		require.ErrorIs(t, err, drivers.ErrNSMClient)
		assert.ErrorContains(t, err, "missing random bytes")
	})
}
