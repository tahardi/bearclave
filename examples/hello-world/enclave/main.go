package main

import (
	"context"
	"flag"
	"log/slog"
	"os"

	"github.com/tahardi/bearclave/examples/hello-world/sdk"
)

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

	attester, err := sdk.MakeAttester(config.Platform)
	if err != nil {
		logger.Error("making attester", slog.String("error", err.Error()))
		return
	}

	transporter, err := sdk.MakeTransporter(
		config.Platform,
		config.SendCID,
		config.SendPort,
		config.ReceivePort,
	)
	if err != nil {
		logger.Error("making transporter", slog.String("error", err.Error()))
		return
	}

	for {
		logger.Info("waiting to receive userdata from enclave-proxy...")
		ctx := context.Background()
		userdata, err := transporter.Receive(ctx)
		if err != nil {
			logger.Error("receiving userdata", slog.String("error", err.Error()))
			return
		}

		logger.Info("attesting userdata", slog.String("userdata", string(userdata)))
		attestation, err := attester.Attest(userdata)
		if err != nil {
			logger.Error("attesting userdata", slog.String("error", err.Error()))
			return
		}

		logger.Info("sending attestation to enclave-proxy...")
		err = transporter.Send(ctx, attestation)
		if err != nil {
			logger.Error("sending attestation", slog.String("error", err.Error()))
			return
		}
	}
}
