package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/tahardi/bearclave/internal/attestation"
	"github.com/tahardi/bearclave/internal/networking"
	"github.com/tahardi/bearclave/internal/setup"
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

func MakeAttestUserDataHandler(
	attester attestation.Attester,
	logger *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := AttestUserDataRequest{}
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			writeError(w, fmt.Errorf("decoding request: %w", err))
			return
		}

		logger.Info("attesting userdata", slog.String("userdata", string(req.Data)))
		att, err := attester.Attest(req.Data)
		if err != nil {
			writeError(w, fmt.Errorf("attesting userdata: %w", err))
			return
		}
		writeResponse(w, AttestUserDataResponse{Attestation: att})
	}
}

const serviceName = "enclave-server"

var configFile string

func main() {
	flag.StringVar(
		&configFile,
		"config",
		setup.DefaultConfigFile,
		"The Trusted Computing platform to use. Options: "+
			"nitro, sev, tdx, notee (default: notee)",
	)
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config, err := setup.LoadConfig(configFile)
	if err != nil {
		logger.Error("loading config", err)
		return
	}
	logger.Info("loaded config", slog.Any(configFile, config))

	attester, err := attestation.NewAttester(config.Platform)
	if err != nil {
		logger.Error("making attester", slog.String("error", err.Error()))
		return
	}

	serverMux := http.NewServeMux()
	serverMux.Handle("POST "+"/attest-user-data", MakeAttestUserDataHandler(attester, logger))

	serverConfig, exists := config.Server[serviceName]
	if !exists {
		logger.Error("missing server config", slog.String("service", serviceName))
		return
	}

	server, err := networking.NewServer(
		config.Platform,
		serverConfig.Port,
		serverMux,
	)
	if err != nil {
		logger.Error("making server", slog.String("error", err.Error()))
		return
	}

	logger.Info("Enclave server started", slog.String("addr", server.Addr()))
	if err = server.Serve(); err != nil {
		logger.Error("Enclave server error", slog.String("error", err.Error()))
	}
}
