package main

import (
	"flag"
	"log/slog"
	"net/http"
	"os"

	"github.com/tahardi/bearclave/internal/attestation"
	"github.com/tahardi/bearclave/internal/networking"
	"github.com/tahardi/bearclave/internal/setup"
)

const serviceName = "enclave-server"

var configFile string

func main() {
	flag.StringVar(
		&configFile,
		"config",
		setup.DefaultConfigFile,
		"The Trusted Computing platform to use. Options: "+
			"nitro, sev, tdx, notee (default: notee)",
	)
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config, err := setup.LoadConfig(configFile)
	if err != nil {
		logger.Error("loading config", slog.Any("error", err))
		return
	}
	logger.Info("loaded config", slog.Any(configFile, config))

	attester, err := attestation.NewAttester(config.Platform)
	if err != nil {
		logger.Error("making attester", slog.String("error", err.Error()))
		return
	}

	serverConfig, exists := config.Server[serviceName]
	if !exists {
		logger.Error(
			"missing server config",
			slog.String("service", serviceName),
		)
		return
	}

	serverMux := http.NewServeMux()
	serverMux.Handle(
		"POST "+networking.AttestUserDataPath,
		networking.MakeAttestUserDataHandler(attester, logger),
	)

	server, err := networking.NewServer(
		config.Platform,
		serverConfig.Port,
		serverMux,
	)
	if err != nil {
		logger.Error("making server", slog.String("error", err.Error()))
		return
	}

	logger.Info("Enclave server started", slog.String("addr", server.Addr()))
	if err = server.Serve(); err != nil {
		logger.Error(
			"Enclave server error",
			slog.String("error", err.Error()),
		)
	}
}
