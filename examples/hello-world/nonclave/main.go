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
	case sdk.Nitro:
		return bearclave.NewNitroVerifier()
	case sdk.SEV:
		return bearclave.NewSEVVerifier()
	case sdk.TDX:
		return bearclave.NewTDXVerifier()
	case sdk.Unsafe:
		return bearclave.NewUnsafeVerifier()
	default:
		return nil, fmt.Errorf("unsupported platform '%s'", platform)
	}
}

func MakeCommunicator(platform sdk.Platform) (bearclave.Communicator, error) {
	switch platform {
	case sdk.Nitro:
		return bearclave.NewNitroCommunicator(
			enclaveCID,
			enclavePort,
			nonclavePort,
		)
	case sdk.SEV:
		return bearclave.NewSEVCommunicator(
			enclaveCID,
			enclavePort,
			nonclavePort,
		)
	case sdk.TDX:
		return bearclave.NewTDXCommunicator(
			enclaveCID,
			enclavePort,
			nonclavePort,
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

var enclaveCID int
var enclavePort int
var nonclavePort int

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
		&enclaveCID,
		"enclaveCID",
		bearclave.NitroEnclaveCID,
		"The context ID of the enclave that the non-enclave should connect to",
	)
	flag.IntVar(
		&enclavePort,
		"enclavePort",
		8080,
		"The port of the enclave that the non-enclave should connect to",
	)
	flag.IntVar(
		&nonclavePort,
		"nonclavePort",
		8081,
		"The port that the non-enclave should listen on",
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
