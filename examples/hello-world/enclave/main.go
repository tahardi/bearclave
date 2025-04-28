package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/tahardi/bearclave"
	"github.com/tahardi/bearclave/examples/hello-world/sdk"
)

func MakeAttester(platform sdk.Platform) (bearclave.Attester, error) {
	switch platform {
	case sdk.ConfidentialVMs:
		return bearclave.NewCVMSAttester()
	case sdk.Nitro:
		return bearclave.NewNitroAttester()
	case sdk.Unsafe:
		return bearclave.NewUnsafeAttester()
	default:
		return nil, fmt.Errorf("unsupported platform '%s'", platform)
	}
}

func MakeCommunicator(platform sdk.Platform) (bearclave.Communicator, error) {
	switch platform {
	case sdk.ConfidentialVMs:
		return bearclave.NewCVMSCommunicator()
	case sdk.Nitro:
		return bearclave.NewNitroCommunicator(
			nitroNonclaveCID,
			nitroNonclavePort,
			nitroEnclavePort,
		)
	case sdk.Unsafe:
		return bearclave.NewUnsafeCommunicator(
			unsafeNonclaveAddr,
			unsafeEnclaveAddr,
		)
	default:
		return nil, fmt.Errorf("unsupported platform '%s'", platform)
	}
}

var platform string

var nitroNonclaveCID int
var nitroEnclavePort int
var nitroNonclavePort int

var unsafeEnclaveAddr string
var unsafeNonclaveAddr string

func main() {
	flag.StringVar(
		&platform,
		"platform",
		"unsafe",
		"The Trusted Computing platform to use. Options: "+
			"cvms, nitro, unsafe (default: unsafe)",
	)

	flag.IntVar(
		&nitroNonclaveCID,
		"nitroNonclaveCID",
		bearclave.NitroNonclaveCID,
		"The context ID that the non-enclave should use when using Nitro",
	)
	flag.IntVar(
		&nitroEnclavePort,
		"nitroEnclavePort",
		8080,
		"The port that the enclave should listen on when using Nitro",
	)
	flag.IntVar(
		&nitroNonclavePort,
		"nitroNonclavePort",
		8081,
		"The port that the non-enclave should listen on when using Nitro",
	)

	flag.StringVar(
		&unsafeEnclaveAddr,
		"unsafeEnclaveAddr",
		"127.0.0.1:8080",
		"The address that the enclave should listen on when using Unsafe",
	)
	flag.StringVar(
		&unsafeNonclaveAddr,
		"unsafeNonclaveAddr",
		"127.0.0.1:8081",
		"The address that the non-enclave should listen on when using Unsafe",
	)

	attester, err := MakeAttester(sdk.Platform(platform))
	if err != nil {
		panic(err)
	}

	communicator, err := MakeCommunicator(sdk.Platform(platform))
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	fmt.Printf("Listening on: %s\n", unsafeEnclaveAddr)
	userdata, err := communicator.Receive(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Attesting userdata: %s\n", userdata)
	attestation, err := attester.Attest(userdata)
	if err != nil {
		panic(err)
	}

	err = communicator.Send(ctx, attestation)
	if err != nil {
		panic(err)
	}
}
