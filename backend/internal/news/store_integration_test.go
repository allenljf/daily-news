package news

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/allenljf/daily-news/backend/internal/identity"
	"github.com/allenljf/daily-news/backend/internal/platform"
	"github.com/google/uuid"
)

func TestStorePreservesFeedAndArticleLifecycleParity(t *testing.T) {
	database := newsIntegrationDatabase(t)
	defer database.Close()
	ctx := context.Background()
	categoryID, firstSourceID, secondSourceID := seedCategory(t, database)
	store := NewStore(database)
	articleIDs := make([]uuid.UUID, 0, 21)
	for position := 1; position <= 21; position++ {
		articleIDs = append(articleIDs, insertArticle(t, database, categoryID, firstSourceID, position, time.Now().AddDate(0, 0, 1), nil))
	}

	first, err := store.List(ctx, categoryID, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(first.Items) != pageSize || first.NextCursor == nil {
		t.Fatalf("first page = %#v", first)
	}
	second, err := store.List(ctx, categoryID, *first.NextCursor, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(second.Items) != 1 || second.NextCursor != nil {
		t.Fatalf("second page = %#v", second)
	}

	filteredID := insertArticle(t, database, categoryID, secondSourceID, 22, time.Now().AddDate(0, 0, 1), nil)
	_ = filteredID
	insertArticle(t, database, categoryID, secondSourceID, 23, time.Now().AddDate(0, 0, -1), nil)
	insertArticle(t, database, categoryID, secondSourceID, 24, time.Now().AddDate(0, 0, 1), ptrTime(time.Now()))
	filtered, err := store.List(ctx, categoryID, "", &secondSourceID)
	if err != nil {
		t.Fatal(err)
	}
	if len(filtered.Items) != 1 || filtered.Items[0].SourceTagID != secondSourceID {
		t.Fatalf("filtered = %#v", filtered.Items)
	}

	permanent, err := store.SetPermanent(ctx, articleIDs[0], true)
	if err != nil || permanent.ExpiresAt != nil {
		t.Fatalf("SetPermanent() = %#v, %v", permanent, err)
	}
	otherCategoryID, otherSourceID := seedOtherCategory(t, database)
	if _, err := database.Exec(`INSERT INTO category_articles (category_id,article_id,source_setting_id) VALUES ($1,$2,$3)`, otherCategoryID, articleIDs[0], otherSourceID); err != nil {
		t.Fatal(err)
	}
	if err := store.Delete(ctx, articleIDs[0]); err != nil {
		t.Fatal(err)
	}
	for _, visibleCategoryID := range []uuid.UUID{categoryID, otherCategoryID} {
		visible, err := store.List(ctx, visibleCategoryID, "", nil)
		if err != nil {
			t.Fatal(err)
		}
		for _, item := range visible.Items {
			if item.ID == articleIDs[0] {
				t.Fatal("globally deleted Article remains visible")
			}
		}
	}
	if _, err := store.Detail(ctx, categoryID, uuid.New()); err != ErrNotFound {
		t.Fatalf("Detail() error = %v, want ErrNotFound", err)
	}
	if _, err := store.List(ctx, categoryID, "invalid", nil); err != ErrInvalidCursor {
		t.Fatalf("List() error = %v, want ErrInvalidCursor", err)
	}
	handler := NewHandler(store, newsVerifier{}, "allowed@example.com")
	response := serveNews(handler, http.MethodGet, "/v1/categories/"+categoryID.String()+"/news?sourceTagId="+secondSourceID.String(), "")
	if response.Code != http.StatusOK {
		t.Fatalf("filtered HTTP status = %d, body = %s", response.Code, response.Body.String())
	}
	var listResponse struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &listResponse); err != nil {
		t.Fatal(err)
	}
	assertNewsJSONFields(t, listResponse.Items[0], false)
	response = serveNews(handler, http.MethodGet, "/v1/categories/"+categoryID.String()+"/news/"+articleIDs[1].String(), "")
	if response.Code != http.StatusOK {
		t.Fatalf("detail HTTP status = %d, body = %s", response.Code, response.Body.String())
	}
	var detailResponse map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &detailResponse); err != nil {
		t.Fatal(err)
	}
	assertNewsJSONFields(t, detailResponse, true)
	response = serveNews(handler, http.MethodPatch, "/v1/news/"+articleIDs[1].String(), `{"permanent":true}`)
	if response.Code != http.StatusOK {
		t.Fatalf("permanent HTTP status = %d, body = %s", response.Code, response.Body.String())
	}
	var permanentResponse map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &permanentResponse); err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"id", "title", "summary", "canonical_url", "published_at", "first_seen_at", "expires_at"} {
		if _, ok := permanentResponse[field]; !ok {
			t.Fatalf("permanent response missing %q", field)
		}
	}
	response = serveNews(handler, http.MethodDelete, "/v1/news/"+articleIDs[1].String(), "")
	if response.Code != http.StatusNoContent {
		t.Fatalf("delete HTTP status = %d", response.Code)
	}
	response = serveNews(handler, http.MethodGet, "/v1/categories/"+categoryID.String()+"/news?cursor=invalid", "")
	if response.Code != http.StatusBadRequest {
		t.Fatalf("invalid cursor HTTP status = %d", response.Code)
	}
	response = serveNews(handler, http.MethodGet, "/v1/categories/"+uuid.NewString()+"/news", "")
	if response.Code != http.StatusNotFound {
		t.Fatalf("missing category HTTP status = %d", response.Code)
	}
	response = serveNews(handler, http.MethodPatch, "/v1/news/"+articleIDs[1].String(), `{}`)
	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("missing permanent HTTP status = %d", response.Code)
	}
}

func TestStoreKeepsArticlesAfterSourceSettingReplacement(t *testing.T) {
	database := newsIntegrationDatabase(t)
	defer database.Close()
	ctx := context.Background()
	categoryID, sourceID, _ := seedCategory(t, database)
	articleID := insertArticle(t, database, categoryID, sourceID, 1, time.Now().AddDate(0, 0, 1), nil)

	// A Category edit soft-deletes its existing Source Settings and inserts new
	// ones; existing Category Articles must remain visible.
	if _, err := database.Exec(`UPDATE source_settings SET deleted_at=now() WHERE id=$1`, sourceID); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`INSERT INTO source_settings (id,category_id,label,website_input,kind,position) VALUES ($1,$2,'New','https://new.example','website',0)`, uuid.New(), categoryID); err != nil {
		t.Fatal(err)
	}

	store := NewStore(database)
	page, err := store.List(ctx, categoryID, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 1 || page.Items[0].ID != articleID {
		t.Fatalf("items = %#v", page.Items)
	}
	if _, err := store.Detail(ctx, categoryID, articleID); err != nil {
		t.Fatalf("Detail() error = %v", err)
	}
}

func assertNewsJSONFields(t *testing.T, value map[string]any, detail bool) {
	t.Helper()
	fields := []string{"id", "title", "summary", "canonical_url", "published_at", "inserted_at", "expires_at", "source_tag_id", "source_tag_label"}
	if detail {
		fields = append(fields, "first_seen_at")
	}
	for _, field := range fields {
		if _, ok := value[field]; !ok {
			t.Fatalf("news response missing %q", field)
		}
	}
}

type newsVerifier struct{}

func (newsVerifier) Verify(context.Context, string) (identity.Identity, error) {
	return identity.Identity{UID: "user", Email: "allowed@example.com"}, nil
}

func serveNews(handler http.Handler, method, target, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, target, strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer valid")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func seedCategory(t *testing.T, database *sql.DB) (uuid.UUID, uuid.UUID, uuid.UUID) {
	t.Helper()
	categoryID, firstSourceID, secondSourceID := uuid.New(), uuid.New(), uuid.New()
	if _, err := database.Exec(`INSERT INTO categories (id,name) VALUES ($1,'News')`, categoryID); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`INSERT INTO source_settings (id,category_id,label,website_input,kind,position) VALUES ($1,$2,'Open web','','unspecified',0),($3,$2,'GitHub','github.com','website',1)`, firstSourceID, categoryID, secondSourceID); err != nil {
		t.Fatal(err)
	}
	return categoryID, firstSourceID, secondSourceID
}

func seedOtherCategory(t *testing.T, database *sql.DB) (uuid.UUID, uuid.UUID) {
	t.Helper()
	categoryID, sourceID := uuid.New(), uuid.New()
	if _, err := database.Exec(`INSERT INTO categories (id,name) VALUES ($1,'Other')`, categoryID); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`INSERT INTO source_settings (id,category_id,label,website_input,kind,position) VALUES ($1,$2,'Other source','','unspecified',0)`, sourceID, categoryID); err != nil {
		t.Fatal(err)
	}
	return categoryID, sourceID
}

func insertArticle(t *testing.T, database *sql.DB, categoryID, sourceID uuid.UUID, position int, expiresAt time.Time, deletedAt *time.Time) uuid.UUID {
	t.Helper()
	id := uuid.New()
	insertedAt := time.Date(2026, 9, 1, 0, 0, position, 0, time.UTC)
	if _, err := database.Exec(`INSERT INTO articles (id,title,normalized_title_hash,canonical_url,canonical_url_hash,first_seen_at,expires_at,deleted_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`, id, fmt.Sprintf("Article %d", position), fmt.Sprintf("%064x", position), fmt.Sprintf("https://example.test/%d", position), fmt.Sprintf("%064x", position+1000), insertedAt, expiresAt, deletedAt); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`INSERT INTO category_articles (category_id,article_id,source_setting_id,inserted_at) VALUES ($1,$2,$3,$4)`, categoryID, id, sourceID, insertedAt); err != nil {
		t.Fatal(err)
	}
	return id
}
func ptrTime(value time.Time) *time.Time { return &value }

func newsIntegrationDatabase(t *testing.T) *sql.DB {
	t.Helper()
	name := "daily-news-go-news-test"
	newsDocker(t, "run", "--rm", "--detach", "--name", name, "-e", "POSTGRES_USER=daily_news", "-e", "POSTGRES_PASSWORD=daily_news", "-e", "POSTGRES_DB=daily_news", "-p", "127.0.0.1::5432", "postgres:16-alpine")
	t.Cleanup(func() { exec.Command("docker", "rm", "-f", name).Run() })
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		if exec.Command("docker", "exec", name, "pg_isready", "-U", "daily_news", "-d", "daily_news").Run() == nil {
			break
		}
		time.Sleep(time.Second)
	}
	port := strings.TrimSpace(newsDocker(t, "port", name, "5432/tcp"))
	port = port[strings.LastIndex(port, ":")+1:]
	url := "postgres://daily_news:daily_news@127.0.0.1:" + port + "/daily_news?sslmode=disable"
	waitForNewsPostgreSQLTCP(t, url)
	path, err := filepath.Abs("../../migrations")
	if err != nil {
		t.Fatal(err)
	}
	if err = platform.ApplyMigrations("file://"+path, url); err != nil {
		t.Fatal(err)
	}
	database, err := platform.OpenDB(context.Background(), url)
	if err != nil {
		t.Fatal(err)
	}
	return database
}

func waitForNewsPostgreSQLTCP(t *testing.T, databaseURL string) {
	t.Helper()
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		connectionContext, cancel := context.WithTimeout(context.Background(), time.Second)
		database, err := platform.OpenDB(connectionContext, databaseURL)
		cancel()
		if err == nil {
			_ = database.Close()
			return
		}
		time.Sleep(250 * time.Millisecond)
	}
	t.Fatalf("PostgreSQL did not accept TCP connections within 30 seconds")
}

func newsDocker(t *testing.T, arguments ...string) string {
	t.Helper()
	output, err := exec.Command("docker", arguments...).CombinedOutput()
	if err != nil {
		t.Fatalf("docker %s: %v\n%s", strings.Join(arguments, " "), err, output)
	}
	return string(output)
}
