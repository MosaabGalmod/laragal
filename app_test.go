package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/laragal/laragal/internal/config"
)

func TestProjectsDirectoryLogic(t *testing.T) {
	tempDir := t.TempDir()
	cfgPath := filepath.Join(tempDir, "laragal.json")

	cfg := config.NewDefaultConfig()
	cfg.ProjectsPath = "C:\\laragal\\sites"
	if err := cfg.Save(cfgPath); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	app := &App{
		cfg:     cfg,
		cfgPath: cfgPath,
	}

	// Verify initial get
	if got := app.GetProjectsDirectory(); got != "C:\\laragal\\sites" {
		t.Errorf("expected C:\\laragal\\sites, got %s", got)
	}

	// Update directory
	newDir := filepath.Join(tempDir, "my-custom-projects")
	_ = os.MkdirAll(newDir, 0755)
	app.cfg.ProjectsPath = newDir
	if err := app.cfg.Save(app.cfgPath); err != nil {
		t.Fatalf("failed to save updated config: %v", err)
	}

	// Reload config from disk
	loadedCfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("failed to reload config: %v", err)
	}

	if loadedCfg.ProjectsPath != newDir {
		t.Errorf("expected loaded path %s, got %s", newDir, loadedCfg.ProjectsPath)
	}
}

func TestScanProjects(t *testing.T) {
	tempDir := t.TempDir()
	cfgPath := filepath.Join(tempDir, "laragal.json")

	// Create test project folders:
	// 1. laravel-app with public/ directory
	// 2. simple-php with direct index.php
	// 3. docs-folder (non-web directory)
	laravelDir := filepath.Join(tempDir, "laravel-app", "public")
	_ = os.MkdirAll(laravelDir, 0755)
	_ = os.WriteFile(filepath.Join(laravelDir, "index.php"), []byte("<?php echo 'laravel';"), 0644)

	simpleDir := filepath.Join(tempDir, "simple-php")
	_ = os.MkdirAll(simpleDir, 0755)
	_ = os.WriteFile(filepath.Join(simpleDir, "index.php"), []byte("<?php echo 'simple';"), 0644)

	docsDir := filepath.Join(tempDir, "docs-folder")
	_ = os.MkdirAll(docsDir, 0755)
	_ = os.WriteFile(filepath.Join(docsDir, "README.md"), []byte("# Docs"), 0644)

	cfg := config.NewDefaultConfig()
	cfg.ProjectsPath = tempDir

	app := &App{
		cfg:     cfg,
		cfgPath: cfgPath,
	}

	sites := app.ScanProjects()
	if len(sites) != 3 {
		t.Fatalf("expected 3 directories scanned, got %d", len(sites))
	}

	foundLaravel := false
	foundSimple := false
	foundDocsNonWeb := false
	for _, s := range sites {
		if s.Domain == "laravel-app.test" {
			foundLaravel = true
			if !s.IsWeb {
				t.Errorf("expected laravel-app to be web site")
			}
			if !filepath.IsAbs(filepath.FromSlash(s.RootPath)) {
				t.Errorf("expected absolute root path, got %s", s.RootPath)
			}
		}
		if s.Domain == "simple-php.test" {
			foundSimple = true
			if !s.IsWeb {
				t.Errorf("expected simple-php to be web site")
			}
		}
		if s.Name == "docs-folder" {
			foundDocsNonWeb = true
			if s.IsWeb {
				t.Errorf("expected docs-folder to NOT be a web site")
			}
			if s.Domain != "" {
				t.Errorf("expected non-web site domain to be empty, got %s", s.Domain)
			}
		}
	}

	if !foundLaravel || !foundSimple || !foundDocsNonWeb {
		t.Errorf("did not find expected sites: laravel=%v, simple=%v, docs=%v", foundLaravel, foundSimple, foundDocsNonWeb)
	}
}

func TestRecentDirectories(t *testing.T) {
	tempDir := t.TempDir()
	cfgPath := filepath.Join(tempDir, "laragal.json")

	cfg := config.NewDefaultConfig()
	cfg.RecentPaths = []string{}
	cfg.ProjectsPath = filepath.Join(tempDir, "dir1")

	app := &App{
		cfg:     cfg,
		cfgPath: cfgPath,
	}

	app.addRecentDirectory(filepath.Join(tempDir, "dir1"))
	app.addRecentDirectory(filepath.Join(tempDir, "dir2"))
	app.addRecentDirectory(filepath.Join(tempDir, "dir3"))

	recent := app.GetRecentDirectories()
	if len(recent) != 3 {
		t.Fatalf("expected 3 recent dirs, got %d", len(recent))
	}
	if recent[0] != filepath.Join(tempDir, "dir3") {
		t.Errorf("expected most recent to be dir3, got %s", recent[0])
	}

	// Switch directory
	newDir := filepath.Join(tempDir, "dir4")
	_ = os.MkdirAll(newDir, 0755)
	sites, err := app.SetProjectsDirectory(newDir)
	if err != nil {
		t.Fatalf("SetProjectsDirectory failed: %v", err)
	}
	if app.GetProjectsDirectory() != newDir {
		t.Errorf("expected current dir %s, got %s", newDir, app.GetProjectsDirectory())
	}
	if len(sites) != 0 {
		t.Errorf("expected 0 sites in empty dir, got %d", len(sites))
	}
}


func TestScanProjectDatabasesIntegration(t *testing.T) {
	tempProject := t.TempDir()
	sqlDir := filepath.Join(tempProject, "database")
	_ = os.MkdirAll(sqlDir, 0755)

	sqlFile := filepath.Join(sqlDir, "schema.sql")
	_ = os.WriteFile(sqlFile, []byte("CREATE TABLE users (id INT PRIMARY KEY);"), 0644)

	app := &App{}
	results, err := app.ScanProjectDatabases(tempProject)
	if err != nil {
		t.Fatalf("unexpected error from ScanProjectDatabases: %v", err)
	}

	if len(results) == 0 {
		t.Fatalf("expected at least 1 discovered database, got 0")
	}

	ok, err := app.ImportProjectDatabase(tempProject, results[0])
	if err != nil {
		t.Fatalf("unexpected error from ImportProjectDatabase: %v", err)
	}
	if !ok {
		t.Errorf("expected import to return true, got false")
	}

	// Verify .env was generated
	if _, err := os.Stat(filepath.Join(tempProject, ".env")); err != nil {
		t.Errorf("expected .env to be created, but error: %v", err)
	}
}

func TestCheckRuntimeUpdates(t *testing.T) {
	app := &App{}
	updates := app.CheckRuntimeUpdates()
	if len(updates) == 0 {
		t.Errorf("expected non-empty updates list")
	}
}
