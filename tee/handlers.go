package tee

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"time"
)

const (
	DefaultProxyTimeout        = 30 * time.Second
	ProxyConnectionEstablished = "HTTP/1.1 200 Connection Established\r\n\r\n"
	ProxyBadGateway            = "HTTP/1.1 502 Bad Gateway\r\n\r\n"
)

func MakeProxyTLSHandler(
	logger *slog.Logger,
	timeout time.Duration,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("proxy tls request received", slog.String("URL", r.URL.String()))
		if r.Method != http.MethodConnect {
			msg := "method should be CONNECT but got: " + r.Method
			logger.Error(msg)
			WriteError(w, proxyError(msg, nil))
			return
		}

		hijacker, ok := w.(http.Hijacker)
		if !ok {
			msg := "hijacking is not supported"
			logger.Error(msg)
			WriteError(w, proxyError(msg, nil))
			return
		}

		clientConn, _, err := hijacker.Hijack()
		if err != nil {
			msg := "hijacking connection"
			logger.Error(msg, slog.String("error", err.Error()))
			WriteError(w, proxyError(msg, err))
			return
		}
		defer clientConn.Close()

		// NOTE: Do NOT write to ResponseWriter after hijacking connection
		dialCtx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()

		targetAddr := r.RequestURI
		serverConn, err := (&net.Dialer{}).DialContext(dialCtx, NetworkTCP4, targetAddr)
		if err != nil {
			msg := "dialing: " + targetAddr
			logger.Error(msg, slog.String("error", err.Error()))
			_, _ = clientConn.Write([]byte(ProxyBadGateway))
			return
		}
		defer serverConn.Close()

		logger.Info("connection established")
		_, err = clientConn.Write([]byte(ProxyConnectionEstablished))
		if err != nil {
			logger.Error("writing response", slog.String("error", err.Error()))
			return
		}

		connCtx, connCancel := context.WithTimeout(r.Context(), timeout)
		defer connCancel()

		connDone := make(chan error, NumConnDoneChannels)
		go func() {
			_, connErr := copyNoSplice(serverConn, clientConn)
			connDone <- connErr
		}()
		go func() {
			_, connErr := copyNoSplice(clientConn, serverConn)
			connDone <- connErr
		}()

		select {
		case connErr := <-connDone:
			if connErr != nil && !errors.Is(connErr, io.EOF) {
				logger.Error("copy error", slog.String("error", connErr.Error()))
			} else {
				logger.Info("connection closed")
			}
		case <-connCtx.Done():
			logger.Error("proxy timeout", slog.String("error", connCtx.Err().Error()))
		}
	}
}

func MakeProxyHandler(
	client *http.Client,
	logger *slog.Logger,
	timeout time.Duration,
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

		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()
		f, err := http.NewRequestWithContext(
			ctx, r.Method, targetURL.String(), r.Body,
		)
		if err != nil {
			logger.Error(
				"creating request", slog.String("error", err.Error()),
			)
			WriteError(w, proxyError("creating request", err))
			return
		}

		CopyHTTPHeadersForForwarding(f.Header, r.Header)
		SetHTTPHeadersForForwarding(f, r)
		f.RequestURI = ""

		logger.Info("forwarding request", slog.String("url", f.URL.String()))
		resp, err := client.Do(f)
		if err != nil {
			logger.Error(
				"forwarding request", slog.String("error", err.Error()),
			)
			WriteError(w, proxyError("forwarding request", err))
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
			WriteError(w, proxyError("copying response body", err))
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
		WriteError(w, proxyError("marshaling response", err))
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(data)
	if err != nil {
		WriteError(w, proxyError("writing response", err))
		return
	}
}
