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

func MakeAttester(config *sdk.Config) (bearclave.Attester, error) {
	switch config.Platform {
	case sdk.Nitro:
		return bearclave.NewNitroAttester()
	case sdk.SEV:
		return bearclave.NewSEVAttester()
	case sdk.TDX:
		return bearclave.NewTDXAttester()
	case sdk.Unsafe:
		return bearclave.NewUnsafeAttester()
	default:
		return nil, fmt.Errorf("unsupported platform '%s'", config.Platform)
	}
}

func MakeCommunicator(config *sdk.Config) (bearclave.Communicator, error) {
	switch config.Platform {
	case sdk.Nitro:
		return bearclave.NewNitroCommunicator(
			config.NonclaveCID,
			config.NonclavePort,
			config.EnclavePort,
		)
	case sdk.SEV:
		return bearclave.NewSEVCommunicator(
			"0.0.0.0:8081",
			"0.0.0.0:8082",
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
		logger.Error("loading config", err)
		return
	}
	logger.Info("loaded config", slog.Any(configFile, config))

	//attester, err := MakeAttester(config)
	//if err != nil {
	//	logger.Error("making attester", slog.String("error", err.Error()))
	//	return
	//}

	communicator, err := MakeCommunicator(config)
	if err != nil {
		logger.Error("making communicator", slog.String("error", err.Error()))
		return
	}

	for {
		logger.Info("Waiting to receive userdata from non-enclave...")
		ctx := context.Background()
		userdata, err := communicator.Receive(ctx)
		if err != nil {
			logger.Error("receiving userdata", slog.String("error", err.Error()))
			return
		}

		//logger.Info("Attesting userdata", slog.String("userdata", string(userdata)))
		//attestation, err := attester.Attest(userdata)
		//if err != nil {
		//	logger.Error("attesting userdata", slog.String("error", err.Error()))
		//	return
		//}

		attestation := userdata
		err = communicator.Send(ctx, attestation)
		if err != nil {
			logger.Error("sending attestation", slog.String("error", err.Error()))
			return
		}
	}
}
