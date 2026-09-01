package httpapi

import (
	"crypto/subtle"
	"errors"
	"net/http"
	"strings"

	"github.com/allenljf/daily-news/backend/internal/identity"
)

// RequireAllowed verifies a Firebase Bearer token and applies the email allowlist.
func RequireAllowed(verifier identity.TokenVerifier, allowedEmail string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		token, ok := bearerToken(request.Header.Get("Authorization"))
		if !ok {
			WriteProblem(writer, Problem{Title: "Authentication required", Status: http.StatusUnauthorized})
			return
		}

		verifiedIdentity, err := verifier.Verify(request.Context(), token)
		if err != nil {
			if errors.Is(err, identity.ErrInvalidToken) {
				WriteProblem(writer, Problem{Title: "Authentication required", Status: http.StatusUnauthorized})
				return
			}
			WriteProblem(writer, Problem{Title: "Authentication unavailable", Status: http.StatusServiceUnavailable})
			return
		}

		if allowedEmail == "" || subtle.ConstantTimeCompare([]byte(verifiedIdentity.Email), []byte(allowedEmail)) != 1 {
			WriteProblem(writer, Problem{Title: "Access forbidden", Status: http.StatusForbidden})
			return
		}

		next.ServeHTTP(writer, request)
	})
}

func bearerToken(header string) (string, bool) {
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", false
	}
	return parts[1], true
}
