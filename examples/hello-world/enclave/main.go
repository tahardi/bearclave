package main

import (
	"context"
	"flag"
	"log/slog"
	"os"

	"github.com/tahardi/bearclave/pkg/attestation"
	"github.com/tahardi/bearclave/pkg/ipc"
	"github.com/tahardi/bearclave/pkg/setup"
)

const enclaveName = "enclave"
const proxyName = "enclave-proxy"

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
		logger.Error("loading config", slog.Any("error", err))
		return
	}
	logger.Info("loaded config", slog.Any(configFile, config))

	attester, err := attestation.NewAttester(config.Platform)
	if err != nil {
		logger.Error("making attester", slog.String("error", err.Error()))
		return
	}

	enclaveIPC, exists := config.IPCs[enclaveName]
	if !exists {
		logger.Error("missing IPC config", slog.String("service", enclaveName))
		return
	}

	communicator, err := ipc.NewIPC(config.Platform, enclaveIPC.Endpoint)
	if err != nil {
		logger.Error("making ipc", slog.String("error", err.Error()))
		return
	}

	proxyIPC, exists := config.IPCs[proxyName]
	if !exists {
		logger.Error("missing IPC config", slog.String("service", proxyName))
	}

	for {
		logger.Info("waiting to receive userdata from enclave-proxy...")
		ctx := context.Background()
		userdata, err := communicator.Receive(ctx)
		if err != nil {
			logger.Error("receiving userdata", slog.String("error", err.Error()))
			return
		}

		logger.Info("attesting userdata", slog.String("userdata", string(userdata)))
		att, err := attester.Attest(userdata)
		if err != nil {
			logger.Error("attesting userdata", slog.String("error", err.Error()))
			return
		}

		logger.Info("sending attestation to enclave-proxy...")
		err = communicator.Send(ctx, proxyIPC.Endpoint, att)
		if err != nil {
			logger.Error("sending attestation", slog.String("error", err.Error()))
			return
		}
	}
}
