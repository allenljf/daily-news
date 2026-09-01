package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/allenljf/daily-news/backend/internal/identity"
)

func TestRequireAllowedRejectsMissingBearerToken(t *testing.T) {
	t.Parallel()

	response := serveProtected(t, fakeVerifier{}, "allowed@example.com", "")
	assertProblem(t, response, http.StatusUnauthorized, "Authentication required")
}

func TestRequireAllowedRejectsMalformedBearerToken(t *testing.T) {
	t.Parallel()

	response := serveProtected(t, fakeVerifier{}, "allowed@example.com", "Token malformed")
	assertProblem(t, response, http.StatusUnauthorized, "Authentication required")
}

func TestRequireAllowedRejectsInvalidFirebaseToken(t *testing.T) {
	t.Parallel()

	response := serveProtected(t, fakeVerifier{err: identity.ErrInvalidToken}, "allowed@example.com", "Bearer invalid")
	assertProblem(t, response, http.StatusUnauthorized, "Authentication required")
}

func TestRequireAllowedDoesNotMisclassifyFirebaseOperationalFailure(t *testing.T) {
	t.Parallel()

	response := serveProtected(t, fakeVerifier{err: errors.New("certificate service unavailable")}, "allowed@example.com", "Bearer valid")
	assertProblem(t, response, http.StatusServiceUnavailable, "Authentication unavailable")
}

func TestRequireAllowedRejectsNonAllowlistedEmail(t *testing.T) {
	t.Parallel()

	response := serveProtected(t, fakeVerifier{identity: identity.Identity{UID: "user", Email: "other@example.com"}}, "allowed@example.com", "Bearer valid")
	assertProblem(t, response, http.StatusForbidden, "Access forbidden")
}

func TestDecodeJSONRejectsUnknownFieldsWithProblemDetails(t *testing.T) {
	t.Parallel()

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/v1/test", strings.NewReader(`{"known":"value","unknown":true}`))
	handler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		var body struct {
			Known string `json:"known"`
		}
		if err := DecodeJSON(request, &body); err != nil {
			WriteProblem(writer, ValidationProblem(err))
			return
		}
		writer.WriteHeader(http.StatusNoContent)
	})

	handler.ServeHTTP(response, request)
	assertProblem(t, response, http.StatusUnprocessableEntity, "Validation failed")
}

type fakeVerifier struct {
	identity identity.Identity
	err      error
}

func (verifier fakeVerifier) Verify(_ context.Context, _ string) (identity.Identity, error) {
	if verifier.err != nil {
		return identity.Identity{}, verifier.err
	}
	return verifier.identity, nil
}

func serveProtected(t *testing.T, verifier identity.TokenVerifier, allowedEmail string, authorization string) *httptest.ResponseRecorder {
	t.Helper()

	request := httptest.NewRequest(http.MethodGet, "/v1/test", nil)
	if authorization != "" {
		request.Header.Set("Authorization", authorization)
	}
	response := httptest.NewRecorder()
	RequireAllowed(verifier, allowedEmail, http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusNoContent)
	})).ServeHTTP(response, request)
	return response
}

func assertProblem(t *testing.T, response *httptest.ResponseRecorder, status int, title string) {
	t.Helper()

	if response.Code != status {
		t.Fatalf("status = %d, want %d", response.Code, status)
	}
	if got, want := response.Header().Get("Content-Type"), ProblemJSONMediaType; got != want {
		t.Fatalf("Content-Type = %q, want %q", got, want)
	}
	want := `{"detail":{"title":"` + title + `","status":` + strconv.Itoa(status) + `}}` + "\n"
	if got := response.Body.String(); got != want {
		t.Fatalf("body = %q, want %q", got, want)
	}
}
