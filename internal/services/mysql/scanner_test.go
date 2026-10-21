package mysql

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtractTableNamesFromSQL(t *testing.T) {
	sqlSample := `
		CREATE TABLE IF NOT EXISTS users (
			id INT PRIMARY KEY,
			name VARCHAR(255)
		);
		CREATE TABLE ` + "`audit_tasks`" + ` (
			id INT PRIMARY KEY,
			status VARCHAR(50)
		);
		create table "sessions" (
			id VARCHAR(128) PRIMARY KEY
		);
		CREATE TABLE products(id INT);
	`
	tables := ExtractTableNamesFromSQL(sqlSample)
	expected := []string{"users", "audit_tasks", "sessions", "products"}

	if len(tables) != len(expected) {
		t.Fatalf("expected %d tables, got %d: %v", len(expected), len(tables), tables)
	}

	for i, name := range expected {
		if tables[i] != name {
			t.Errorf("at index %d, expected table %q, got %q", i, name, tables[i])
		}
	}
}

func TestNormalizeTargetDB(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"audit-platform", "audit_platform"},
		{"My-Super_Project.v2", "my_super_project_v2"},
		{"123-app", "app_123_app"},
		{"normal_db", "normal_db"},
	}

	for _, c := range cases {
		got := NormalizeTargetDB(c.input)
		if got != c.expected {
			t.Errorf("NormalizeTargetDB(%q) = %q; expected %q", c.input, got, c.expected)
		}
	}
}

func TestScanProjectForSchemas(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "laragal-scanner-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create a nested dummy project structure
	dbSubDir := filepath.Join(tempDir, "src", "Infrastructure", "Database")
	if err := os.MkdirAll(dbSubDir, 0755); err != nil {
		t.Fatal(err)
	}

	schemaContent := `
		CREATE TABLE IF NOT EXISTS users (id INT PRIMARY KEY);
		CREATE TABLE IF NOT EXISTS sessions (id VARCHAR(128) PRIMARY KEY);
		CREATE TABLE IF NOT EXISTS audit_tasks (id INT PRIMARY KEY);
	`
	schemaPath := filepath.Join(dbSubDir, "schema.sql")
	if err := os.WriteFile(schemaPath, []byte(schemaContent), 0644); err != nil {
		t.Fatal(err)
	}

	results, err := ScanProjectForSchemas(tempDir)
	if err != nil {
		t.Fatalf("unexpected error scanning project: %v", err)
	}

	if len(results) == 0 {
		t.Fatalf("expected at least 1 discovered schema, got 0")
	}

	found := false
	for _, r := range results {
		if r.FileName == "schema.sql" {
			found = true
			if len(r.FoundTables) != 3 {
				t.Errorf("expected 3 found tables, got %d: %v", len(r.FoundTables), r.FoundTables)
			}
			if r.Type != "schema_file" {
				t.Errorf("expected type 'schema_file', got %q", r.Type)
			}
		}
	}

	if !found {
		t.Errorf("did not find schema.sql in results: %v", results)
	}
}

func TestScanHostForLegacyDatabasesInPaths(t *testing.T) {
	tempHostDir, err := os.MkdirTemp("", "laragal-host-db-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempHostDir)

	mysqlDataDir := filepath.Join(tempHostDir, "data", "mysql")
	// Create user db and system dbs
	userDbDir := filepath.Join(mysqlDataDir, "audit_platform")
	sysDbDir := filepath.Join(mysqlDataDir, "mysql")
	perfDbDir := filepath.Join(mysqlDataDir, "performance_schema")

	for _, d := range []string{userDbDir, sysDbDir, perfDbDir} {
		if err := os.MkdirAll(d, 0755); err != nil {
			t.Fatal(err)
		}
		// Write a dummy table file
		_ = os.WriteFile(filepath.Join(d, "users.ibd"), []byte("dummy"), 0644)
	}

	results := scanHostDirsForDatabases([]string{mysqlDataDir}, "audit_platform")
	if len(results) != 1 {
		t.Fatalf("expected 1 user database detected, got %d: %v", len(results), results)
	}

	if results[0].TargetDB != "audit_platform" {
		t.Errorf("expected TargetDB 'audit_platform', got %q", results[0].TargetDB)
	}
	if results[0].Type != "legacy_server_dir" {
		t.Errorf("expected Type 'legacy_server_dir', got %q", results[0].Type)
	}
}


func TestScanRealAuditPlatform(t *testing.T) {
	auditPath := `D:\projects\audit-platform`
	if _, err := os.Stat(auditPath); os.IsNotExist(err) {
		t.Skip("D:\\projects\\audit-platform does not exist on this machine")
	}

	results, err := ScanAllDatabases(auditPath)
	if err != nil {
		t.Fatalf("unexpected error scanning audit-platform: %v", err)
	}

	if len(results) == 0 {
		t.Fatalf("expected at least 1 discovered database for audit-platform, got 0")
	}

	t.Logf("Discovered %d candidates for audit-platform:", len(results))
	for _, r := range results {
		t.Logf("- Type: %s | File: %s | TargetDB: %s | Tables: %v", r.Type, r.FileName, r.TargetDB, r.FoundTables)
	}

	// Verify schema.sql was found
	foundSchema := false
	for _, r := range results {
		if r.FileName == "schema.sql" {
			foundSchema = true
			if len(r.FoundTables) == 0 {
				t.Errorf("expected extracted tables from schema.sql, got 0")
			}
		}
	}

	if !foundSchema {
		t.Errorf("expected to find schema.sql in audit-platform")
	}
}
