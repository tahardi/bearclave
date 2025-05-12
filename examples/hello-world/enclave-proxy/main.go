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

func MakeAttestUserDataHandler(transporter bearclave.Transporter, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := AttestUserDataRequest{}
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			writeError(w, fmt.Errorf("decoding request: %w", err))
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		logger.Info("sending userdata to enclave...", slog.String("userdata", string(req.Data)))
		err = transporter.Send(ctx, req.Data)
		if err != nil {
			writeError(w, fmt.Errorf("sending userdata to enclave: %w", err))
			return
		}

		logger.Info("waiting for attestation from enclave...")
		attestation, err := transporter.Receive(ctx)
		if err != nil {
			writeError(w, fmt.Errorf("receiving attestation from enclave: %w", err))
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

	transporter, err := sdk.MakeTransporter(
		config.Platform,
		config.SendCID,
		config.SendPort,
		config.ReceivePort,
	)
	if err != nil {
		logger.Error("making transporter", slog.String("error", err.Error()))
		return
	}

	serverMux := http.NewServeMux()
	serverMux.Handle("POST "+"/attest-user-data", MakeAttestUserDataHandler(transporter, logger))

	// TODO: Do I want to set other options? Take in port from config?
	server := &http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: serverMux,
	}

	logger.Info("HTTP server started", slog.String("addr", server.Addr))
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("HTTP server error", slog.String("error", err.Error()))
	}
}
