package platform

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestApplyMigrationsMatchesDailyNewsSchema(t *testing.T) {
	name := "daily-news-go-schema-test"
	runDocker(t, "run", "--rm", "--detach", "--name", name, "-e", "POSTGRES_USER=daily_news", "-e", "POSTGRES_PASSWORD=daily_news", "-e", "POSTGRES_DB=daily_news", "-p", "127.0.0.1::5432", "postgres:16-alpine")
	t.Cleanup(func() { exec.Command("docker", "rm", "-f", name).Run() })

	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		if exec.Command("docker", "exec", name, "pg_isready", "-U", "daily_news", "-d", "daily_news").Run() == nil {
			break
		}
		time.Sleep(time.Second)
	}
	port := strings.TrimSpace(runDocker(t, "port", name, "5432/tcp"))
	port = port[strings.LastIndex(port, ":")+1:]
	databaseURL := "postgres://daily_news:daily_news@127.0.0.1:" + port + "/daily_news?sslmode=disable"
	waitForPostgreSQLTCP(t, databaseURL)
	migrationsPath, err := filepath.Abs("../../migrations")
	if err != nil {
		t.Fatal(err)
	}
	if err := ApplyMigrations("file://"+migrationsPath, databaseURL); err != nil {
		t.Fatalf("ApplyMigrations() error = %v", err)
	}
	database, err := OpenDB(context.Background(), databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	var tables string
	if err := database.QueryRowContext(context.Background(), `SELECT string_agg(tablename, ',' ORDER BY tablename) FROM pg_tables WHERE schemaname='public'`).Scan(&tables); err != nil {
		t.Fatal(err)
	}
	if want := "articles,categories,category_articles,ingestion_attempts,ingestion_runs,schema_migrations,source_settings"; tables != want {
		t.Fatalf("tables = %q, want %q", tables, want)
	}
	var indexes string
	if err := database.QueryRowContext(context.Background(), `SELECT string_agg(indexname, ',' ORDER BY indexname) FROM pg_indexes WHERE schemaname='public' AND indexname IN ('ix_articles_canonical_url_hash','ix_articles_normalized_title_hash','ix_category_articles_category_article','ix_category_articles_feed_cursor','ix_ingestion_runs_idempotency_key')`).Scan(&indexes); err != nil {
		t.Fatal(err)
	}
	if want := "ix_articles_canonical_url_hash,ix_articles_normalized_title_hash,ix_category_articles_category_article,ix_category_articles_feed_cursor,ix_ingestion_runs_idempotency_key"; indexes != want {
		t.Fatalf("indexes = %q, want %q", indexes, want)
	}
}

func waitForPostgreSQLTCP(t *testing.T, databaseURL string) {
	t.Helper()
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		context, cancel := context.WithTimeout(context.Background(), time.Second)
		database, err := OpenDB(context, databaseURL)
		cancel()
		if err == nil {
			_ = database.Close()
			return
		}
		time.Sleep(250 * time.Millisecond)
	}
	t.Fatalf("PostgreSQL did not accept TCP connections within 30 seconds")
}

func runDocker(t *testing.T, arguments ...string) string {
	t.Helper()
	command := exec.Command("docker", arguments...)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("docker %s: %v\n%s", strings.Join(arguments, " "), err, output)
	}
	return fmt.Sprint(string(output))
}
