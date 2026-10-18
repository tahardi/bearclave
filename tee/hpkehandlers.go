package tee

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

const (
	HPKENonceParam          = "nonce"
	HPKENonceSize           = 32
	MaxHPKENonceSize        = 64
	DefaultHPKEMaxBodyBytes = Megabyte
)

type HPKEHandlerFunc func(ctx context.Context, plaintext []byte) ([]byte, error)

func HPKEAAD(method string, path string) []byte {
	return []byte(method + " " + path)
}

func MakeFetchHPKEKeyHandler(attester *Attester, key *HPKEKey, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		encoded := r.URL.Query().Get(HPKENonceParam)
		nonce, err := base64.RawURLEncoding.DecodeString(encoded)
		switch {
		case encoded == "":
			http.Error(w, "missing nonce", http.StatusBadRequest)
			return
		case err != nil:
			http.Error(w, "decoding nonce", http.StatusBadRequest)
			return
		case len(nonce) > MaxHPKENonceSize:
			http.Error(w, "nonce too long", http.StatusBadRequest)
			return
		}

		binding := KeyBinding{
			Purpose:   HPKEKeyBindingPurpose,
			PublicKey: key.PublicKey(),
			Nonce:     nonce,
			IssuedAt:  time.Now().Unix(),
		}
		att, err := AttestKeyBinding(attester, binding)
		if err != nil {
			logger.Error("attesting hpke key", slog.String("error", err.Error()))
			http.Error(w, "attesting hpke key", http.StatusInternalServerError)
			return
		}
		WriteResponse(w, att)
	}
}

func MakeSendSealedRequestHandler(key *HPKEKey, handler HPKEHandlerFunc, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		body := http.MaxBytesReader(w, r.Body, DefaultHPKEMaxBodyBytes)
		var msg SealedMessage
		err := json.NewDecoder(body).Decode(&msg)
		if err != nil {
			http.Error(w, "decoding request", http.StatusBadRequest)
			return
		}

		plaintext, responseKey, err := key.Open(msg, HPKEAAD(r.Method, r.URL.Path))
		if err != nil {
			logger.Error("opening request", slog.String("error", err.Error()))
			http.Error(w, "opening request", http.StatusBadRequest)
			return
		}

		reply, err := handler(r.Context(), plaintext)
		if err != nil {
			logger.Error("handling request", slog.String("error", err.Error()))
			http.Error(w, "handling request", http.StatusInternalServerError)
			return
		}
		WriteResponse(w, responseKey.Seal(reply))
	}
}
