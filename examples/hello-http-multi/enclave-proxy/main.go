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
		"configs/enclave/notee.yaml",
		"The Trusted Computing platform to use. Options: "+
			"nitro, sev, tdx, notee (default: notee)",
	)
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config, err := setup.LoadConfig(configFile)
	if err != nil {
		logger.Error("loading config", slog.String("error", err.Error()))
		return
	}
	logger.Info("loaded config", slog.Any(configFile, config))

	routes := make([]string, 0)
	cids := make([]int, 0)
	ports := make([]int, 0)
	for _, server := range config.Servers {
		routes = append(routes, server.Route)
		cids = append(cids, server.CID)
		ports = append(ports, server.Port)
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

	proxyAddr := fmt.Sprintf("0.0.0.0:%d", config.Proxy.Port)
	server := &http.Server{
		Addr:    proxyAddr,
		Handler: proxy,
	}

	logger.Info("Proxy server started", slog.String("addr", server.Addr))
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("Proxy server error", slog.String("error", err.Error()))
	}
}
