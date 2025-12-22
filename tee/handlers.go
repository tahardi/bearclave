package tee

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

const (
	ForwardPath = "/"
	DefaultForwardHTTPRequestTimeout = 30 * time.Second
)

func MakeForwardHTTPRequestHandler(
	client *http.Client,
	logger *slog.Logger,
	ctxTimeout time.Duration,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		targetURL := &url.URL{
			Scheme:   "http",
			Host:     r.Host,
			Path:     r.URL.Path,
			RawPath:  r.URL.RawPath,
			RawQuery: r.URL.RawQuery,
		}

		ctx, cancel := context.WithTimeout(r.Context(), ctxTimeout)
		defer cancel()
		f, err := http.NewRequestWithContext(
			ctx, r.Method, targetURL.String(), r.Body,
		)
		if err != nil {
			logger.Error(
				"creating request", slog.String("error", err.Error()),
			)
			WriteError(w, fmt.Errorf("creating request: %w", err))
			return
		}

		CopyHTTPHeadersForForwarding(f.Header, r.Header)
		SetHTTPHeadersForForwarding(f, r)
		f.Header.Set("X-Forwarded-Host", r.Host)
		f.Header.Set("X-Forwarded-For", r.RemoteAddr)
		f.Header.Set("X-Forwarded-Proto", r.Proto)
		f.RequestURI = ""

		logger.Info("forwarding request", slog.String("url", f.URL.String()))
		resp, err := client.Do(f)
		if err != nil {
			logger.Error(
				"forwarding request", slog.String("error", err.Error()),
			)
			WriteError(w, fmt.Errorf("forwarding request: %w", err))
			return
		}
		defer resp.Body.Close()

		for key, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}

		w.WriteHeader(resp.StatusCode)
		if _, err := io.Copy(w, resp.Body); err != nil {
			logger.Error(
				"copying response body", slog.String("error", err.Error()),
			)
			WriteError(w, fmt.Errorf("copying response body: %w", err))
		}
	}
}

func CopyHTTPHeadersForForwarding(f http.Header, r http.Header) {
	ignoredHeaders := map[string]bool{
		"Connection":          true,
		"Keep-Alive":          true,
		"Proxy-Authenticate":  true,
		"Proxy-Authorization": true,
		"Te":                  true,
		"Trailers":            true,
		"Transfer-Encoding":   true,
		"Upgrade":             true,
	}
	for header, values := range r {
		if !ignoredHeaders[header] {
			for _, value := range values {
				f.Add(header, value)
			}
		}
	}
}

func SetHTTPHeadersForForwarding(f *http.Request, r *http.Request) {
	f.Header.Set("X-Forwarded-Host", r.Host)
	f.Header.Set("X-Forwarded-For", r.RemoteAddr)
	f.Header.Set("X-Forwarded-Proto", r.Proto)
}

func WriteError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func WriteResponse(w http.ResponseWriter, out any) {
	data, err := json.Marshal(out)
	if err != nil {
		fmt.Printf("marshaling response: %s", err.Error())
		WriteError(w, fmt.Errorf("marshaling response: %w", err))
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(data)
	if err != nil {
		WriteError(w, fmt.Errorf("writing response: %w", err))
		return
	}
}
