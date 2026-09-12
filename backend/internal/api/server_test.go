package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/allenljf/daily-news/backend/internal/identity"
	"github.com/google/uuid"
)

func TestHandlerServesHealthzWithoutDatabaseOrAuthentication(t *testing.T) {
	handler := NewHandler(nil, verifier{}, "allowed@example.com", noopLauncher{})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if got := response.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", got)
	}
	var body map[string]string
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("status body = %q, want ok", body["status"])
	}
}

func TestHandlerRoutesEveryV1FamilyThroughSharedAuthentication(t *testing.T) {
	handler := NewHandler(nil, verifier{}, "allowed@example.com", noopLauncher{})
	for _, requestCase := range []struct{ method, target string }{
		{http.MethodGet, "/v1/categories"},
		{http.MethodGet, "/v1/categories/00000000-0000-0000-0000-000000000000/news"},
		{http.MethodPatch, "/v1/news/00000000-0000-0000-0000-000000000000"},
		{http.MethodGet, "/v1/ingestion-runs/latest"},
	} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(requestCase.method, requestCase.target, nil))
		if response.Code != http.StatusUnauthorized {
			t.Fatalf("%s %s status = %d, want 401", requestCase.method, requestCase.target, response.Code)
		}
	}
}

type verifier struct{}

func (verifier) Verify(context.Context, string) (identity.Identity, error) {
	return identity.Identity{}, identity.ErrInvalidToken
}

type noopLauncher struct{}

func (noopLauncher) Launch(context.Context, uuid.UUID) error { return nil }
