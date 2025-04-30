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
	case sdk.Nitro:
		return bearclave.NewNitroAttester()
	case sdk.SEV:
		return bearclave.NewSEVAttester()
	case sdk.TDX:
		return bearclave.NewTDXAttester()
	case sdk.Unsafe:
		return bearclave.NewUnsafeAttester()
	default:
		return nil, fmt.Errorf("unsupported platform '%s'", platform)
	}
}

func MakeCommunicator(platform sdk.Platform) (bearclave.Communicator, error) {
	switch platform {
	case sdk.Nitro:
		return bearclave.NewNitroCommunicator(
			nonclaveCID,
			nonclavePort,
			enclavePort,
		)
	case sdk.SEV:
		return bearclave.NewSEVCommunicator(
			nonclaveCID,
			nonclavePort,
			enclavePort,
		)
	case sdk.TDX:
		return bearclave.NewTDXCommunicator(
			nonclaveCID,
			nonclavePort,
			enclavePort,
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

var nonclaveCID int
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
		&enclavePort,
		"enclavePort",
		8080,
		"The port that the enclave should listen on",
	)
	flag.IntVar(
		&nonclaveCID,
		"nonclaveCID",
		bearclave.NitroNonclaveCID,
		"The context ID of the non-enclave (Nitro: 3, SEV: 2, TDX: 2)",
	)
	flag.IntVar(
		&nonclavePort,
		"nonclavePort",
		8081,
		"The port of the non-enclave that the enclave should connect to",
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

	fmt.Printf("Using platform: %s\n", platform)
	fmt.Printf("Using non-enclave CID: %d\n", nonclaveCID)
	fmt.Printf("Using enclave port: %d\n", enclavePort)
	fmt.Printf("Using non-enclave port: %d\n", nonclavePort)

	attester, err := MakeAttester(sdk.Nitro)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Made attester\n")

	communicator, err := MakeCommunicator(sdk.Nitro)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Made communicator\n")

	ctx := context.Background()
	fmt.Printf("Listening on: %s\n", unsafeEnclaveAddr)
	userdata, err := communicator.Receive(ctx)
	if err != nil {
		fmt.Printf("Error receiving: %s\n", err.Error())
		panic(err)
	}

	fmt.Printf("Attesting userdata: %s\n", userdata)
	attestation, err := attester.Attest(userdata)
	if err != nil {
		fmt.Printf("Error attesting: %s\n", err.Error())
		panic(err)
	}

	err = communicator.Send(ctx, attestation)
	if err != nil {
		fmt.Printf("Error sending: %s\n", err.Error())
		panic(err)
	}
}
