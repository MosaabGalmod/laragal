package mysql

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSyncProjectDatabaseEnv_WhenEnvMissing(t *testing.T) {
	tempProject, err := os.MkdirTemp("", "laragal-env-sync-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempProject)

	// Create an .env.example
	exampleContent := `
APP_NAME=Laravel
APP_ENV=local
APP_KEY=base64:12345
DB_CONNECTION=mysql
DB_HOST=127.0.0.1
DB_PORT=3306
DB_DATABASE=old_placeholder
DB_USERNAME=root
DB_PASSWORD=secret
`
	if err := os.WriteFile(filepath.Join(tempProject, ".env.example"), []byte(exampleContent), 0644); err != nil {
		t.Fatal(err)
	}

	err = SyncProjectDatabaseEnv(tempProject, "audit_platform")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	envBytes, err := os.ReadFile(filepath.Join(tempProject, ".env"))
	if err != nil {
		t.Fatalf("expected .env to be created, but read failed: %v", err)
	}

	envStr := string(envBytes)
	if !strings.Contains(envStr, "DB_DATABASE=audit_platform") {
		t.Errorf("expected DB_DATABASE=audit_platform in .env, got:\n%s", envStr)
	}
	if !strings.Contains(envStr, "APP_KEY=base64:12345") {
		t.Errorf("expected existing keys like APP_KEY to be preserved, got:\n%s", envStr)
	}
}

func TestSyncProjectDatabaseEnv_WhenEnvExists(t *testing.T) {
	tempProject, err := os.MkdirTemp("", "laragal-env-existing-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempProject)

	existingEnv := `
APP_NAME=MyPlatform
APP_KEY=base64:existing
DB_DATABASE=previous_db
`
	if err := os.WriteFile(filepath.Join(tempProject, ".env"), []byte(existingEnv), 0644); err != nil {
		t.Fatal(err)
	}

	err = SyncProjectDatabaseEnv(tempProject, "audit_platform")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	envBytes, err := os.ReadFile(filepath.Join(tempProject, ".env"))
	if err != nil {
		t.Fatalf("failed to read .env: %v", err)
	}

	envStr := string(envBytes)
	if !strings.Contains(envStr, "DB_DATABASE=audit_platform") {
		t.Errorf("expected DB_DATABASE=audit_platform in .env, got:\n%s", envStr)
	}
	if !strings.Contains(envStr, "APP_KEY=base64:existing") {
		t.Errorf("expected APP_KEY to be preserved, got:\n%s", envStr)
	}
}
