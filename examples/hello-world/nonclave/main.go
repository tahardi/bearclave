package main

import (
	"bytes"
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
	unsafeDetector, err := bearclave.NewUnsafeDetector()
	if err != nil {
		return nil, fmt.Errorf("creating Unsafe detector: %w", err)
	}
	nitroDetector, err := bearclave.NewNitroDetector()
	if err != nil {
		return nil, fmt.Errorf("creating Nitro detector: %w", err)
	}
	return []bearclave.Detector{
		cvmsDetector,
		unsafeDetector,
		nitroDetector,
	}, nil
}

func MakeVerifier(detectors []bearclave.Detector) (bearclave.Verifier, error) {
	for _, detector := range detectors {
		platform, ok := detector.Detect()
		switch {
		case !ok:
			continue
		case platform == bearclave.CVMSPlatform:
			return bearclave.NewCVMSVerifier()
		case platform == bearclave.UnsafePlatform:
			return bearclave.NewUnsafeVerifier()
		case platform == bearclave.NitroPlatform:
			return bearclave.NewNitroVerifier()
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
		case platform == bearclave.UnsafePlatform:
			return bearclave.NewUnsafeNonclaveCommunicator(
				enclaveAddr,
				nonclaveAddr,
			)
		case platform == bearclave.NitroPlatform:
			return bearclave.NewNitroCommunicator()
		}
	}
	return nil, fmt.Errorf("no supported platforms detected")
}

var enclaveAddr string
var nonclaveAddr string

func main() {
	flag.StringVar(
		&enclaveAddr,
		"enclave",
		"127.0.0.1:8080",
		"The address that the enclave should listen on",
	)
	flag.StringVar(
		&nonclaveAddr,
		"nonclave",
		"127.0.0.1:8081",
		"The address that the non-enclave should listen on",
	)

	detectors, err := MakeDetectors()
	if err != nil {
		panic(err)
	}

	verifier, err := MakeVerifier(detectors)
	if err != nil {
		panic(err)
	}

	communicator, err := MakeCommunicator(detectors)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fmt.Printf("Sending to: %s\n", enclaveAddr)
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
