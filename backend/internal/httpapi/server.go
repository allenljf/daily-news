// Package httpapi owns the Daily News HTTP transport boundary.
package httpapi

import (
	"encoding/json"
	"net/http"
	"time"
)

// NewMux creates the API route multiplexer.
func NewMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthz)
	return mux
}

// NewServer creates the Cloud Run HTTP server with bounded connection handling.
func NewServer(address string) *http.Server {
	return NewServerWithHandler(address, NewMux())
}

// NewServerWithHandler creates the Cloud Run server for an explicit API handler.
func NewServerWithHandler(address string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              address,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}

func healthz(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(writer).Encode(map[string]string{"status": "ok"})
}
