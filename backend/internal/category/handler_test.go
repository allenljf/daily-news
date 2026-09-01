package category

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/allenljf/daily-news/backend/internal/identity"
)

func TestHandlerRejectsUnauthenticatedCategoryList(t *testing.T) {
	t.Parallel()

	mux := NewHandler(nil, fakeVerifier{}, "allowed@example.com")
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/v1/categories", nil))

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
}

func TestHandlerRejectsBlankCategoryName(t *testing.T) {
	t.Parallel()

	mux := NewHandler(nil, fakeVerifier{identity: identity.Identity{UID: "user", Email: "allowed@example.com"}}, "allowed@example.com")
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/v1/categories", strings.NewReader(`{"name":"  ","source_settings":[]}`))
	request.Header.Set("Authorization", "Bearer token")
	mux.ServeHTTP(response, request)

	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnprocessableEntity)
	}
}

type fakeVerifier struct{ identity identity.Identity }

func (verifier fakeVerifier) Verify(context.Context, string) (identity.Identity, error) {
	return verifier.identity, nil
}
