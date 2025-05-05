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

	"github.com/tahardi/bearclave"
	"github.com/tahardi/bearclave/examples/hello-world/sdk"
)

func MakeCommunicator(config *sdk.Config) (bearclave.Communicator, error) {
	switch config.Platform {
	case sdk.Nitro:
		return bearclave.NewNitroCommunicator(
			config.EnclaveCID,
			config.EnclavePort,
			config.NonclavePort,
		)
	case sdk.SEV:
		return bearclave.NewSEVCommunicator(
			config.EnclaveAddr,
			config.NonclaveAddr,
		)
	case sdk.TDX:
		return bearclave.NewTDXCommunicator(
			config.EnclaveCID,
			config.EnclavePort,
			config.NonclavePort,
		)
	case sdk.Unsafe:
		return bearclave.NewUnsafeCommunicator(
			config.EnclaveAddr,
			config.NonclaveAddr,
		)
	default:
		return nil, fmt.Errorf("unsupported platform '%s'", config.Platform)
	}
}

func writeError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func writeResponse(w http.ResponseWriter, out any) {
	data, err := json.Marshal(out)
	if err != nil {
		writeError(w, fmt.Errorf("marshaling response: %w", err))
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(data)
	if err != nil {
		writeError(w, fmt.Errorf("writing response: %w", err))
		return
	}
}

type AttestUserDataRequest struct {
	Data []byte `json:"data"`
}
type AttestUserDataResponse struct {
	Attestation []byte `json:"attestation"`
}

func MakeAttestUserDataHandler(communicator bearclave.Communicator, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := AttestUserDataRequest{}
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			writeError(w, fmt.Errorf("decoding request: %w", err))
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		logger.Info("Sending userdata to enclave...", slog.String("userdata", string(req.Data)))
		err = communicator.Send(ctx, req.Data)
		if err != nil {
			logger.Error("sending userdata", slog.String("error", err.Error()))
			return
		}

		attestation, err := communicator.Receive(ctx)
		if err != nil {
			logger.Error("receiving attestation", slog.String("error", err.Error()))
			return
		}
		writeResponse(w, AttestUserDataResponse{Attestation: attestation})
	}
}

var configFile string

func main() {
	flag.StringVar(
		&configFile,
		"config",
		sdk.DefaultConfigFile,
		"The Trusted Computing platform to use. Options: "+
			"nitro, sev, tdx, unsafe (default: unsafe)",
	)
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config, err := sdk.LoadConfig(configFile)
	if err != nil {
		logger.Error("loading config", slog.String("error", err.Error()))
		return
	}
	logger.Info("loaded config", slog.Any(configFile, config))

	communicator, err := MakeCommunicator(config)
	if err != nil {
		logger.Error("making communicator", slog.String("error", err.Error()))
		return
	}

	serverMux := http.NewServeMux()
	serverMux.Handle("POST "+"/attest-user-data", MakeAttestUserDataHandler(communicator, logger))

	// TODO: Do I want to set other options?
	server := &http.Server{
		Addr:    ":8080",
		Handler: serverMux,
	}

	logger.Info("Starting HTTP server on '%s'", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("HTTP server error", slog.String("error", err.Error()))
	}
}
