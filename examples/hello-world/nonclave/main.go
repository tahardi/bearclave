package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/tahardi/bearclave"
	"github.com/tahardi/bearclave/examples/hello-world/sdk"
)

func MakeVerifier(config *sdk.Config) (bearclave.Verifier, error) {
	switch config.Platform {
	case sdk.Nitro:
		return bearclave.NewNitroVerifier()
	case sdk.SEV:
		return bearclave.NewSEVVerifier()
	case sdk.TDX:
		return bearclave.NewTDXVerifier()
	case sdk.Unsafe:
		return bearclave.NewUnsafeVerifier()
	default:
		return nil, fmt.Errorf("unsupported platform '%s'", config.Platform)
	}
}

func MakeCommunicator(config *sdk.Config) (bearclave.Communicator, error) {
	switch config.Platform {
	case sdk.Nitro:
		return bearclave.NewNitroCommunicator(
			config.EnclaveCID,
			config.EnclavePort,
			config.NonclavePort,
		)
	case sdk.SEV:
		return bearclave.NewSEVCommunicator(
			config.EnclaveCID,
			config.EnclavePort,
			config.NonclavePort,
		)
	case sdk.TDX:
		return bearclave.NewTDXCommunicator(
			config.EnclaveCID,
			config.EnclavePort,
			config.NonclavePort,
		)
	case sdk.Unsafe:
		return bearclave.NewUnsafeCommunicator(
			config.EnclaveAddr,
			config.NonclaveAddr,
		)
	default:
		return nil, fmt.Errorf("unsupported platform '%s'", config.Platform)
	}
}

var configFile string

func main() {
	flag.StringVar(
		&configFile,
		"config",
		sdk.DefaultConfigFile,
		"The Trusted Computing platform to use. Options: "+
			"nitro, sev, tdx, unsafe (default: unsafe)",
	)
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config, err := sdk.LoadConfig(configFile)
	if err != nil {
		logger.Error("loading config", slog.String("error", err.Error()))
		return
	}
	logger.Info("loaded config", slog.Any(configFile, config))

	verifier, err := MakeVerifier(config)
	if err != nil {
		logger.Error("making verifier", slog.String("error", err.Error()))
		return
	}

	communicator, err := MakeCommunicator(config)
	if err != nil {
		logger.Error("making communicator", slog.String("error", err.Error()))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	want := []byte("Hello, world!")
	logger.Info("Sending userdata to enclave...", slog.String("userdata", string(want)))
	err = communicator.Send(ctx, want)
	if err != nil {
		logger.Error("sending userdata", slog.String("error", err.Error()))
		return
	}

	attestation, err := communicator.Receive(ctx)
	if err != nil {
		logger.Error("receiving attestation", slog.String("error", err.Error()))
		return
	}

	got, err := verifier.Verify(attestation)
	if err != nil {
		logger.Error("verifying attestation", slog.String("error", err.Error()))
		return
	}

	if !bytes.Equal(got, want) {
		logger.Error("userdata verification failed")
		return
	}
	logger.Info("verified userdata", slog.String("userdata", string(got)))
}
