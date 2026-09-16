package category

import (
	"context"
	"database/sql"
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

func TestStorePreservesCategoryLifecycleSemantics(t *testing.T) {
	database := integrationDatabase(t)
	defer database.Close()
	store := NewStore(database)
	created, err := store.Create(context.Background(), Request{Name: "AI research", SourceSettings: []SourceSettingInput{{Label: "Open web"}, {Label: "OpenAI", WebsiteInput: "openai.com", Kind: "website"}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(created.SourceSettings) != 2 || created.SourceSettings[1].NormalizedHost == nil || *created.SourceSettings[1].NormalizedHost != "openai.com" {
		t.Fatalf("created settings = %#v", created.SourceSettings)
	}
	updated, err := store.Update(context.Background(), created.ID, Request{Name: "AI updates", SourceSettings: []SourceSettingInput{{Label: "GitHub", WebsiteInput: "https://github.com", Kind: "website"}}})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != "AI updates" || len(updated.SourceSettings) != 1 || updated.SourceSettings[0].Position != 0 {
		t.Fatalf("updated = %#v", updated)
	}
	articleID := uuid.New()
	if _, err := database.Exec(`INSERT INTO articles (id,title,normalized_title_hash,canonical_url,canonical_url_hash) VALUES ($1,'kept',$2,'https://example.test/a',$3)`, articleID, strings.Repeat("a", 64), strings.Repeat("b", 64)); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`INSERT INTO category_articles (category_id,article_id,source_setting_id) VALUES ($1,$2,$3)`, created.ID, articleID, updated.SourceSettings[0].ID); err != nil {
		t.Fatal(err)
	}
	if err := store.Delete(context.Background(), created.ID); err != nil {
		t.Fatal(err)
	}
	var state string
	if err := database.QueryRow(`SELECT (SELECT deleted_at IS NOT NULL FROM categories WHERE id=$1)::text || ',' || (SELECT deleted_at IS NOT NULL FROM category_articles WHERE category_id=$1 AND article_id=$2)::text || ',' || (SELECT count(*) FROM articles WHERE id=$2)::text`, created.ID, articleID).Scan(&state); err != nil {
		t.Fatal(err)
	}
	if state != "true,true,1" {
		t.Fatalf("delete state = %q", state)
	}
	if _, err := store.Update(context.Background(), created.ID, Request{Name: "no", SourceSettings: []SourceSettingInput{}}); err != ErrNotFound {
		t.Fatalf("Update() error = %v, want ErrNotFound", err)
	}
}

func TestStorePreservesDefaultContentLanguageAndDatabaseConstraint(t *testing.T) {
	database := integrationDatabase(t)
	defer database.Close()
	store := NewStore(database)

	created, err := store.Create(context.Background(), Request{Name: "Taiwan news", SourceSettings: []SourceSettingInput{}})
	if err != nil {
		t.Fatal(err)
	}
	if created.ContentLanguage != "zh-Hant" {
		t.Fatalf("created content language = %q, want %q", created.ContentLanguage, "zh-Hant")
	}
	var persisted string
	if err := database.QueryRow(`SELECT content_language FROM categories WHERE id=$1`, created.ID).Scan(&persisted); err != nil {
		t.Fatal(err)
	}
	if persisted != "zh-Hant" {
		t.Fatalf("persisted content language = %q, want %q", persisted, "zh-Hant")
	}
	listed, err := store.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(listed) != 1 || listed[0].ContentLanguage != "zh-Hant" {
		t.Fatalf("listed categories = %#v, want one zh-Hant category", listed)
	}

	var databaseDefault string
	if err := database.QueryRow(`INSERT INTO categories (id,name) VALUES ($1,$2) RETURNING content_language`, uuid.New(), "Default language").Scan(&databaseDefault); err != nil {
		t.Fatal(err)
	}
	if databaseDefault != "zh-Hant" {
		t.Fatalf("database default content language = %q, want %q", databaseDefault, "zh-Hant")
	}

	english, err := store.Create(context.Background(), Request{
		Name:               "English news",
		ContentLanguage:    EnglishContentLanguage,
		contentLanguageSet: true,
	})
	if err != nil {
		t.Fatalf("create English category: %v", err)
	}
	if english.ContentLanguage != EnglishContentLanguage {
		t.Fatalf("English content language = %q, want %q", english.ContentLanguage, EnglishContentLanguage)
	}
	if _, err := database.Exec(`INSERT INTO categories (id,name,content_language) VALUES ($1,$2,$3)`, uuid.New(), "Another English category", EnglishContentLanguage); err != nil {
		t.Fatalf("database rejected English content language: %v", err)
	}
	if _, err := database.Exec(`INSERT INTO categories (id,name,content_language) VALUES ($1,$2,$3)`, uuid.New(), "Unsupported language", "fr"); err == nil {
		t.Fatal("INSERT with unsupported content language succeeded")
	}
}

func TestHandlerListsNoCategoriesAsJSONArray(t *testing.T) {
	database := integrationDatabase(t)
	defer database.Close()
	mux := NewHandler(NewStore(database), fakeVerifier{identity: identity.Identity{UID: "user", Email: "allowed@example.com"}}, "allowed@example.com")
	request := httptest.NewRequest(http.MethodGet, "/v1/categories", nil)
	request.Header.Set("Authorization", "Bearer token")
	response := httptest.NewRecorder()

	mux.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if body := response.Body.String(); body != "[]\n" {
		t.Fatalf("empty category list body = %q, want %q", body, "[]\n")
	}
}

func integrationDatabase(t *testing.T) *sql.DB {
	t.Helper()
	name := "daily-news-go-category-test"
	docker(t, "run", "--rm", "--detach", "--name", name, "-e", "POSTGRES_USER=daily_news", "-e", "POSTGRES_PASSWORD=daily_news", "-e", "POSTGRES_DB=daily_news", "-p", "127.0.0.1::5432", "postgres:16-alpine")
	t.Cleanup(func() { exec.Command("docker", "rm", "-f", name).Run() })
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		if exec.Command("docker", "exec", name, "pg_isready", "-U", "daily_news", "-d", "daily_news").Run() == nil {
			break
		}
		time.Sleep(time.Second)
	}
	port := strings.TrimSpace(docker(t, "port", name, "5432/tcp"))
	port = port[strings.LastIndex(port, ":")+1:]
	url := "postgres://daily_news:daily_news@127.0.0.1:" + port + "/daily_news?sslmode=disable"
	waitForCategoryPostgreSQLTCP(t, url)
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

func waitForCategoryPostgreSQLTCP(t *testing.T, databaseURL string) {
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

func docker(t *testing.T, arguments ...string) string {
	t.Helper()
	output, err := exec.Command("docker", arguments...).CombinedOutput()
	if err != nil {
		t.Fatalf("docker %s: %v\n%s", strings.Join(arguments, " "), err, output)
	}
	return string(output)
}
