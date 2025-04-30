package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/tahardi/bearclave"
	"github.com/tahardi/bearclave/examples/hello-world/sdk"
)

func MakeAttester(
	platform sdk.Platform,
	config *sdk.Config,
) (bearclave.Attester, error) {
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

func MakeCommunicator(
	platform sdk.Platform,
	config *sdk.Config,
) (bearclave.Communicator, error) {
	switch platform {
	case sdk.Nitro:
		return bearclave.NewNitroCommunicator(
			config.NonclaveCID,
			config.NonclavePort,
			config.EnclavePort,
		)
	case sdk.SEV:
		return bearclave.NewSEVCommunicator(
			config.NonclaveCID,
			config.NonclavePort,
			config.EnclavePort,
		)
	case sdk.TDX:
		return bearclave.NewTDXCommunicator(
			config.NonclaveCID,
			config.NonclavePort,
			config.EnclavePort,
		)
	case sdk.Unsafe:
		return bearclave.NewUnsafeCommunicator(
			config.NonclaveAddr,
			config.EnclaveAddr,
		)
	default:
		return nil, fmt.Errorf("unsupported platform '%s'", platform)
	}
}

var platform string
var configFile string

func main() {
	flag.StringVar(
		&platform,
		"platform",
		"unsafe",
		"The Trusted Computing platform to use. Options: "+
			"nitro, sev, tdx, unsafe (default: unsafe)",
	)
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
		logger.Error("loading config", err)
		return
	}
	logger.Info("loaded config", slog.Any(configFile, config))

	attester, err := MakeAttester(sdk.Platform(platform), config)
	if err != nil {
		logger.Error("making attester", err)
		return
	}

	communicator, err := MakeCommunicator(sdk.Platform(platform), config)
	if err != nil {
		logger.Error("making communicator", err)
		return
	}

	logger.Info("Waiting to receive userdata from non-enclave...")
	ctx := context.Background()
	userdata, err := communicator.Receive(ctx)
	if err != nil {
		logger.Error("receiving userdata", err)
		return
	}

	logger.Info("Attesting userdata", slog.String("userdata", string(userdata)))
	attestation, err := attester.Attest(userdata)
	if err != nil {
		logger.Error("attesting userdata", err)
		return
	}

	err = communicator.Send(ctx, attestation)
	if err != nil {
		logger.Error("sending attestation", err)
		return
	}
}
