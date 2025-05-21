package main

import (
	"flag"
	"log/slog"
	"net/http"
	"os"

	"github.com/tahardi/bearclave/examples/hello-http-multi/examples"
	"github.com/tahardi/bearclave/internal/networking"
	"github.com/tahardi/bearclave/internal/setup"
)

const serviceName = "enclave-server-2"

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
		logger.Error("loading config", err)
		return
	}
	logger.Info("loaded config", slog.Any(configFile, config))

	serverConfig, exists := config.Server[serviceName]
	if !exists {
		logger.Error("missing server config", slog.String("service", serviceName))
		return
	}

	serverMux := http.NewServeMux()
	serverMux.Handle(
		"GET "+examples.HelloMultipleServersPath,
		examples.MakeHelloMultipleServersHandler(logger, serviceName),
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

	logger.Info("Enclave server 2 started", slog.String("addr", server.Addr()))
	if err = server.Serve(); err != nil {
		logger.Error("Enclave server error", slog.String("error", err.Error()))
	}
}
