// Package api composes Daily News domain HTTP handlers into one API handler.
package api

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/allenljf/daily-news/backend/internal/category"
	"github.com/allenljf/daily-news/backend/internal/identity"
	"github.com/allenljf/daily-news/backend/internal/ingestion"
	"github.com/allenljf/daily-news/backend/internal/news"
)

// NewHandler mounts every versioned Daily News route behind one handler.
func NewHandler(database *sql.DB, verifier identity.TokenVerifier, allowedEmail string, launcher ingestion.JobLauncher) http.Handler {
	categories := category.NewHandler(category.NewStore(database), verifier, allowedEmail)
	articles := news.NewHandler(news.NewStore(database), verifier, allowedEmail)
	runs := ingestion.NewHandler(ingestion.NewRunStore(database), launcher, verifier, allowedEmail)
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch {
		case request.URL.Path == "/v1/categories" || (strings.HasPrefix(request.URL.Path, "/v1/categories/") && strings.Contains(request.URL.Path, "/news")):
			if strings.Contains(request.URL.Path, "/news") {
				articles.ServeHTTP(writer, request)
				return
			}
			categories.ServeHTTP(writer, request)
		case strings.HasPrefix(request.URL.Path, "/v1/news/"):
			articles.ServeHTTP(writer, request)
		case strings.HasPrefix(request.URL.Path, "/v1/ingestion-runs"):
			runs.ServeHTTP(writer, request)
		default:
			http.NotFound(writer, request)
		}
	})
}
