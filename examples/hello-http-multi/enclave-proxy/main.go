package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/tahardi/bearclave/pkg/networking"
	"github.com/tahardi/bearclave/pkg/setup"
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

	routes := make([]string, len(proxyConfig.Services))
	cids := make([]int, len(proxyConfig.Services))
	ports := make([]int, len(proxyConfig.Services))
	for i, service := range proxyConfig.Services {
		serverConfig, exists := config.Server[service]
		if !exists {
			logger.Error("missing server config", slog.String("service", service))
			return
		}
		routes[i] = serverConfig.Route
		cids[i] = serverConfig.CID
		ports[i] = serverConfig.Port
	}

	proxy, err := networking.NewMultiProxy(
		config.Platform,
		routes,
		cids,
		ports,
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
		logger.Error("Proxy server error", slog.String("error", err.Error()))
	}
}
