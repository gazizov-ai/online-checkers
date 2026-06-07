//go:build integration

package testutil

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

const defaultDatabaseURL = "postgres://postgres:postgres@localhost:5432/checkers?sslmode=disable"

func NewPostgresDB(t *testing.T, migrationDirs ...string) *sqlx.DB {
	t.Helper()

	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		databaseURL = defaultDatabaseURL
	}

	admin, err := sqlx.Open("postgres", databaseURL)
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	t.Cleanup(func() { _ = admin.Close() })

	ctx := context.Background()
	if err := admin.PingContext(ctx); err != nil {
		t.Skipf("postgres integration database is unavailable: %v", err)
	}

	schema := "test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err := admin.ExecContext(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatalf("create test schema: %v", err)
	}
	t.Cleanup(func() {
		_, _ = admin.ExecContext(context.Background(), "DROP SCHEMA IF EXISTS "+schema+" CASCADE")
	})

	testURL, err := withSearchPath(databaseURL, schema)
	if err != nil {
		t.Fatalf("configure test database URL: %v", err)
	}

	db, err := sqlx.Open("postgres", testURL)
	if err != nil {
		t.Fatalf("open schema database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	root := repositoryRoot(t)
	for _, dir := range migrationDirs {
		files, err := filepath.Glob(filepath.Join(root, dir, "*.sql"))
		if err != nil {
			t.Fatalf("list migrations in %s: %v", dir, err)
		}
		if len(files) == 0 {
			t.Fatalf("no migrations found in %s", dir)
		}

		for _, file := range files {
			contents, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("read migration %s: %v", file, err)
			}

			up := strings.SplitN(string(contents), "-- +goose Down", 2)[0]
			up = strings.Replace(up, "-- +goose Up", "", 1)
			if _, err := db.ExecContext(ctx, up); err != nil {
				t.Fatalf("apply migration %s: %v", file, err)
			}
		}
	}

	return db
}

func withSearchPath(databaseURL, schema string) (string, error) {
	parsed, err := url.Parse(databaseURL)
	if err != nil {
		return "", err
	}

	query := parsed.Query()
	option := "-csearch_path=" + schema
	if existing := query.Get("options"); existing != "" {
		option = existing + " " + option
	}
	query.Set("options", option)
	parsed.RawQuery = query.Encode()

	return parsed.String(), nil
}

func repositoryRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal(fmt.Errorf("go.mod not found from %s", dir))
		}
		dir = parent
	}
}
