package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/tahardi/bearclave/pkg/ipc"
	"github.com/tahardi/bearclave/pkg/networking"
	"github.com/tahardi/bearclave/pkg/setup"
)

func MakeAttestUserDataHandler(
	communicator ipc.IPC,
	enclaveEndpoint string,
	logger *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := networking.AttestUserDataRequest{}
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			networking.WriteError(w, fmt.Errorf("decoding request: %w", err))
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		logger.Info("sending userdata to enclave...", slog.String("userdata", string(req.Data)))
		err = communicator.Send(ctx, enclaveEndpoint, req.Data)
		if err != nil {
			networking.WriteError(w, fmt.Errorf("sending userdata to enclave: %w", err))
			return
		}

		logger.Info("waiting for attestation from enclave...")
		attestation, err := communicator.Receive(ctx)
		if err != nil {
			networking.WriteError(w, fmt.Errorf("receiving attestation from enclave: %w", err))
			return
		}

		resp := networking.AttestUserDataResponse{Attestation: attestation}
		networking.WriteResponse(w, resp)
	}
}

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
		logger.Error("loading config", slog.String("error", err.Error()))
		return
	}
	logger.Info("loaded config", slog.Any(configFile, config))

	proxyIPC, exists := config.IPCs[proxyName]
	if !exists {
		logger.Error("missing IPC config", slog.String("service", proxyName))
		return
	}

	communicator, err := ipc.NewIPC(config.Platform, proxyIPC.Endpoint)
	if err != nil {
		logger.Error("making ipc", slog.String("error", err.Error()))
		return
	}

	enclaveIPC, exists := config.IPCs[enclaveName]
	if !exists {
		logger.Error("missing IPC config", slog.String("service", enclaveName))
		return
	}

	serverMux := http.NewServeMux()
	serverMux.Handle(
		"POST "+networking.AttestUserDataPath,
		MakeAttestUserDataHandler(communicator, enclaveIPC.Endpoint, logger),
	)

	proxyAddr := fmt.Sprintf("0.0.0.0:%d", config.Proxy.Port)
	server := &http.Server{
		Addr:    proxyAddr,
		Handler: serverMux,
	}

	logger.Info("HTTP server started", slog.String("addr", server.Addr))
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("HTTP server error", slog.String("error", err.Error()))
	}
}
