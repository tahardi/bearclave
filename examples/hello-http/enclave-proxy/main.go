package main

import (
	"flag"
	"log/slog"
	"net/http"
	"os"

	"github.com/tahardi/bearclave/internal/networking"
	"github.com/tahardi/bearclave/internal/setup"
)

var configFile string

func main() {
	flag.StringVar(
		&configFile,
		"config",
		setup.DefaultConfigFile,
		"The Trusted Computing platform to use. Options: "+
			"nitro, sev, tdx, unsafe (default: unsafe)",
	)
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config, err := setup.LoadConfig(configFile)
	if err != nil {
		logger.Error("loading config", slog.String("error", err.Error()))
		return
	}
	logger.Info("loaded config", slog.Any(configFile, config))

	proxy, err := networking.NewProxy(
		config.Platform,
		config.SendPort,
		config.SendCID,
	)
	if err != nil {
		logger.Error("making proxy server", slog.String("error", err.Error()))
		return
	}

	server := &http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: proxy.Handler(),
	}

	logger.Info("Proxy server started", slog.String("addr", server.Addr))
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("Proxy server error", slog.String("error", err.Error()))
	}
}
