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
var port int
var platform string

func main() {
	flag.StringVar(
		&host,
		"host",
		"127.0.0.1",
		"The hostname of the enclave gateway to connect to (default: 127.0.0.1)",
	)
	flag.IntVar(
		&port,
		"port",
		8080,
		"The port of the enclave gateway to connect to (default: 8080)",
	)
	flag.StringVar(
		&platform,
		"platform",
		"notee",
		"The Trusted Computing platform the enclave is running on. Options: "+
			"nitro, sev, tdx, notee (default: notee)",
	)
	flag.Parse()

	url := fmt.Sprintf("http://%s:%d", host, port)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	logger.Info("nonclave configuration",
		slog.String("platform", platform),
		slog.String("url", url),
	)

	verifier, err := attestation.NewVerifier(setup.Platform(platform))
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
