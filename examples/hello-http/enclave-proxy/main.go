package main

import (
	"flag"
	"fmt"
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

	proxyConfig := config.Proxy
	if len(proxyConfig.Services) == 0 {
		logger.Error("missing proxy services")
		return
	}

	serverConfig, exists := config.Server[proxyConfig.Services[0]]
	if !exists {
		logger.Error(
			"missing server config",
			slog.String("service", proxyConfig.Services[0]),
		)
		return
	}

	proxy, err := networking.NewProxy(
		config.Platform,
		serverConfig.Route,
		serverConfig.CID,
		serverConfig.Port,
	)
	if err != nil {
		logger.Error("making proxy", slog.String("error", err.Error()))
		return
	}

	proxyAddr := fmt.Sprintf("0.0.0.0:%d", proxyConfig.Port)
	server := &http.Server{
		Addr:    proxyAddr,
		Handler: proxy,
	}

	logger.Info("Proxy server started", slog.String("addr", server.Addr))
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error(
			"Proxy server error",
			slog.String("error", err.Error()),
		)
	}
}
