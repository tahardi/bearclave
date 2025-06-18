package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/tahardi/bearclave/examples/hello-http-multi/examples"
	"github.com/tahardi/bearclave/pkg/attestation"
	"github.com/tahardi/bearclave/pkg/networking"
	"github.com/tahardi/bearclave/pkg/setup"
)

func HelloMultipleServers(client *networking.Client) ([]byte, error) {
	multipleServersResponse := examples.MultipleServersResponse{}
	err := client.Do(
		"GET",
		examples.HelloMultipleServersPath,
		nil,
		&multipleServersResponse,
	)
	if err != nil {
		return nil, err
	}
	return multipleServersResponse.Report, nil
}

var configFile string
var host string

func main() {
	flag.StringVar(
		&configFile,
		"config",
		"configs/nonclave/notee.yaml",
		"The Trusted Computing platform to use. Options: "+
			"nitro, sev, tdx, notee (default: notee)",
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

	verifier, err := attestation.NewVerifier(config.Platform)
	if err != nil {
		logger.Error("making verifier", slog.String("error", err.Error()))
		return
	}

	urls := make([]string, 0)
	measurements := make([]string, 0)
	for service, server := range config.Servers {
		urls = append(
			urls,
			fmt.Sprintf("http://%s:%d%s",
				host,
				config.Proxy.Port,
				server.Route,
			),
		)

		attConfig, exists := config.Attestations[service]
		if !exists {
			logger.Error(
				"missing attestation config",
				slog.String("service", service),
			)
			return
		}
		measurements = append(measurements, attConfig.Measurement)
	}

	for i, url := range urls {
		logger.Info("sending request to url", slog.String("url", url))
		client := networking.NewClient(url)
		report, err := HelloMultipleServers(client)
		if err != nil {
			logger.Error("making request", slog.String("error", err.Error()))
			continue
		}

		userdata, err := verifier.Verify(
			report,
			attestation.WithMeasurement(measurements[i]),
		)
		if err != nil {
			logger.Error("verifying report", slog.String("error", err.Error()))
			continue
		}
		logger.Info("verified userdata", slog.String("userdata", string(userdata)))
	}
}
