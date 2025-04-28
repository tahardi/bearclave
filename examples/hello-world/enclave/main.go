package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/tahardi/bearclave"
)

func MakeDetectors() ([]bearclave.Detector, error) {
	cvmsDetector, err := bearclave.NewCVMSDetector()
	if err != nil {
		return nil, fmt.Errorf("creating CVMS detector: %w", err)
	}
	nitroDetector, err := bearclave.NewNitroDetector()
	if err != nil {
		return nil, fmt.Errorf("creating Nitro detector: %w", err)
	}
	unsafeDetector, err := bearclave.NewUnsafeDetector()
	if err != nil {
		return nil, fmt.Errorf("creating Unsafe detector: %w", err)
	}
	return []bearclave.Detector{
		cvmsDetector,
		nitroDetector,
		unsafeDetector,
	}, nil
}

func MakeAttester(detectors []bearclave.Detector) (bearclave.Attester, error) {
	for _, detector := range detectors {
		platform, ok := detector.Detect()
		switch {
		case !ok:
			continue
		case platform == bearclave.CVMSPlatform:
			return bearclave.NewCVMSAttester()
		case platform == bearclave.NitroPlatform:
			return bearclave.NewNitroAttester()
		case platform == bearclave.UnsafePlatform:
			return bearclave.NewUnsafeAttester()
		}
	}
	return nil, fmt.Errorf("no supported platforms detected")
}

func MakeCommunicator(detectors []bearclave.Detector) (bearclave.Communicator, error) {
	for _, detector := range detectors {
		platform, ok := detector.Detect()
		switch {
		case !ok:
			continue
		case platform == bearclave.CVMSPlatform:
			return bearclave.NewCVMSCommunicator()
		case platform == bearclave.NitroPlatform:
			return bearclave.NewNitroCommunicator(
				nitroNonclaveCID,
				nitroNonclavePort,
				nitroEnclavePort,
			)
		case platform == bearclave.UnsafePlatform:
			return bearclave.NewUnsafeCommunicator(
				unsafeNonclaveAddr,
				unsafeEnclaveAddr,
			)
		}
	}
	return nil, fmt.Errorf("no supported platforms detected")
}

var nitroEnclaveCID int
var nitroNonclaveCID int
var nitroEnclavePort int
var nitroNonclavePort int

var unsafeEnclaveAddr string
var unsafeNonclaveAddr string

func main() {
	flag.IntVar(
		&nitroEnclaveCID,
		"nitroEnclaveCID",
		bearclave.NitroEnclaveCID,
		"The context ID that the enclave should use when using Nitro",
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

	detectors, err := MakeDetectors()
	if err != nil {
		panic(err)
	}

	attester, err := MakeAttester(detectors)
	if err != nil {
		panic(err)
	}

	communicator, err := MakeCommunicator(detectors)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

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
