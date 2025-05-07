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

func MakeTransporter(config *sdk.Config) (bearclave.Transporter, error) {
	switch config.Platform {
	case sdk.Nitro:
		return bearclave.NewNitroTransporter(
			config.EnclaveProxyCID,
			config.EnclaveProxyPort,
			config.EnclavePort,
		)
	case sdk.SEV:
		return bearclave.NewSEVTransporter(
			config.EnclaveProxyPort,
			config.EnclavePort,
		)
	case sdk.TDX:
		return bearclave.NewTDXTransporter(
			config.EnclaveProxyPort,
			config.EnclavePort,
		)
	case sdk.Unsafe:
		return bearclave.NewUnsafeTransporter(
			config.EnclaveProxyPort,
			config.EnclavePort,
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

	transporter, err := MakeTransporter(config)
	if err != nil {
		logger.Error("making transporter", slog.String("error", err.Error()))
		return
	}

	for {
		logger.Info("Waiting to receive userdata from non-enclave...")
		ctx := context.Background()
		userdata, err := transporter.Receive(ctx)
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

		attestation := []byte("Hello from the enclave! Received userdata: ")
		attestation = append(attestation, userdata...)
		err = transporter.Send(ctx, attestation)
		if err != nil {
			logger.Error("sending attestation", slog.String("error", err.Error()))
			return
		}
	}
}
