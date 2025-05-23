package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/tahardi/bearclave/examples/hello-http-multi/examples"
	"github.com/tahardi/bearclave/internal/networking"
	"github.com/tahardi/bearclave/internal/setup"
)

func HelloMultipleServers(client *networking.Client) (string, error) {
	multipleServersResponse := examples.MultipleServersResponse{}
	err := client.Do(
		"GET",
		examples.HelloMultipleServersPath,
		nil,
		&multipleServersResponse,
	)
	if err != nil {
		return "", err
	}
	return multipleServersResponse.Message, nil
}

var configFile string
var host string

func main() {
	flag.StringVar(
		&configFile,
		"config",
		setup.DefaultConfigFile,
		"The Trusted Computing platform to use. Options: "+
			"nitro, sev, tdx, unsafe (default: unsafe)",
	)
	flag.StringVar(
		&host,
		"host",
		"127.0.0.1",
		"The hostname of the enclave gateway to connect to (default: 127.0.0.1)",
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

	urls := make([]string, len(proxyConfig.Services))
	for i, service := range proxyConfig.Services {
		serverConfig, exists := config.Server[service]
		if !exists {
			logger.Error("missing server config", slog.String("service", service))
			return
		}
		urls[i] = fmt.Sprintf("http://%s:%d%s",
			host,
			proxyConfig.Port,
			serverConfig.Route,
		)
	}

	for _, url := range urls {
		logger.Info("sending request to url", slog.String("url", url))
		client := networking.NewClient(url)
		message, err := HelloMultipleServers(client)
		if err != nil {
			logger.Error("making hello multiple servers request", slog.String("error", err.Error()))
			continue
		}
		logger.Info("received message", slog.String("message", message))
	}
}
