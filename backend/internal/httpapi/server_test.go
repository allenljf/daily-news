package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewMuxHealthzReturnsExistingContract(t *testing.T) {
	t.Parallel()

	response := httptest.NewRecorder()
	NewMux().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if got, want := response.Header().Get("Content-Type"), "application/json"; got != want {
		t.Fatalf("Content-Type = %q, want %q", got, want)
	}
	if got, want := response.Body.String(), "{\"status\":\"ok\"}\n"; got != want {
		t.Fatalf("body = %q, want %q", got, want)
	}
}

func TestNewServerShutsDownWithoutRequests(t *testing.T) {
	t.Parallel()

	server := NewServer("127.0.0.1:0")
	if err := server.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}
}
