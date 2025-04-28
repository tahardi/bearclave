package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/tahardi/bearclave"
	"github.com/tahardi/bearclave/examples/hello-world/sdk"
)

func MakeVerifier(platform sdk.Platform) (bearclave.Verifier, error) {
	switch platform {
	case sdk.ConfidentialVMs:
		return bearclave.NewCVMSVerifier()
	case sdk.Nitro:
		return bearclave.NewNitroVerifier()
	case sdk.Unsafe:
		return bearclave.NewUnsafeVerifier()
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
			nitroEnclaveCID,
			nitroEnclavePort,
			nitroNonclavePort,
		)
	case sdk.Unsafe:
		return bearclave.NewUnsafeCommunicator(
			unsafeEnclaveAddr,
			unsafeNonclaveAddr,
		)
	default:
		return nil, fmt.Errorf("unsupported platform '%s'", platform)
	}
}

var platform string

var nitroEnclaveCID int
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
		&nitroEnclaveCID,
		"nitroEnclaveCID",
		bearclave.NitroEnclaveCID,
		"The context ID that the enclave should use when using Nitro",
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
		"enclave",
		"127.0.0.1:8080",
		"The address that the enclave should listen on",
	)
	flag.StringVar(
		&unsafeNonclaveAddr,
		"nonclave",
		"127.0.0.1:8081",
		"The address that the non-enclave should listen on",
	)

	verifier, err := MakeVerifier(sdk.Platform(platform))
	if err != nil {
		panic(err)
	}

	communicator, err := MakeCommunicator(sdk.Platform(platform))
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	fmt.Printf("Sending to: %s\n", unsafeEnclaveAddr)
	want := []byte("Hello, world!")
	err = communicator.Send(ctx, want)
	if err != nil {
		panic(err)
	}

	attestation, err := communicator.Receive(ctx)
	if err != nil {
		panic(err)
	}

	got, err := verifier.Verify(attestation)
	if err != nil {
		panic(err)
	}

	if !bytes.Equal(got, want) {
		panic("got != want")
	}
	fmt.Printf("Verified userdata: %s\n", string(got))
}
