package main

import (
	"bytes"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/tahardi/bearclave/internal/attestation"
	"github.com/tahardi/bearclave/internal/networking"
	"github.com/tahardi/bearclave/internal/setup"
)

var host string
var configFile string

func main() {
	flag.StringVar(
		&configFile,
		"config",
		setup.DefaultConfigFile,
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

	proxyConfig := config.Proxy
	if len(proxyConfig.Services) == 0 {
		logger.Error("missing proxy services")
		return
	}
	url := fmt.Sprintf("http://%s:%d", host, proxyConfig.Port)

	verifier, err := attestation.NewVerifier(config.Platform)
	if err != nil {
		logger.Error("making verifier", slog.String("error", err.Error()))
		return
	}

	want := []byte("Hello, world!")
	client := networking.NewClient(url)
	att, err := client.AttestUserData(want)
	if err != nil {
		logger.Error("attesting userdata", slog.String("error", err.Error()))
		return
	}

	got, err := verifier.Verify(att)
	if err != nil {
		logger.Error("verifying attestation", slog.String("error", err.Error()))
		return
	}

	if !bytes.Contains(got, want) {
		logger.Error("userdata verification failed")
		return
	}
	logger.Info("verified userdata", slog.String("userdata", string(got)))
}
