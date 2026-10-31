package main

import (
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/laragal/laragal/internal/config"
	"github.com/laragal/laragal/internal/process"
	"github.com/laragal/laragal/internal/services/adminer"
	"github.com/laragal/laragal/internal/services/caddy"
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

func TestEnsureAdminerInstalled(t *testing.T) {
	tempDir := t.TempDir()
	path, err := adminer.EnsureInstalled(tempDir)
	if err != nil {
		t.Fatalf("unexpected error ensuring adminer installed: %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Errorf("adminer index.php not created: %v", err)
	}
}

func TestGetActivePHPPort(t *testing.T) {
	cfg := config.NewDefaultConfig()
	cfg.DefaultPHP = "8.3"
	app := &App{cfg: cfg}
	port := app.getActivePHPPort()
	if port != 9083 {
		t.Errorf("expected port 9083 for default php 8.3, got %d", port)
	}
}

func TestGetSites_AutoScansWhenEmpty(t *testing.T) {
	tempDir := t.TempDir()
	pDir := filepath.Join(tempDir, "sample-site", "public")
	_ = os.MkdirAll(pDir, 0755)
	_ = os.WriteFile(filepath.Join(pDir, "index.php"), []byte("<?php echo 'ok';"), 0644)

	cfg := config.NewDefaultConfig()
	cfg.ProjectsPath = tempDir

	app := &App{
		cfg:   cfg,
		sites: []caddy.SiteDef{}, // intentionally empty sites slice
	}

	// Calling GetSites() should auto-scan and return the site rather than returning empty!
	sites := app.GetSites()
	if len(sites) == 0 {
		t.Fatalf("expected GetSites() to auto-scan projects when empty, but got 0 sites")
	}
	if sites[0].Name != "sample-site" {
		t.Errorf("expected scanned site name sample-site, got %s", sites[0].Name)
	}
}

func TestToggleService_PHP85(t *testing.T) {
	if os.Getenv("LARAGAL_INTEGRATION") == "" {
		t.Skip("set LARAGAL_INTEGRATION=1 to run real php-cgi integration test")
	}
	app := NewApp()
	t.Cleanup(func() { _ = app.supervisor.StopAll() })
	svcs := app.GetServices()
	var foundPHP85 bool
	for _, s := range svcs {
		if s.Name == "php-8.5" {
			foundPHP85 = true
			break
		}
	}
	if !foundPHP85 {
		t.Fatalf("expected GetServices() to include php-8.5")
	}

	running, err := app.ToggleService("php-8.5")
	t.Logf("ToggleService(php-8.5) start: running=%v, err=%v", running, err)
	if err != nil {
		t.Fatalf("unexpected error starting php-8.5: %v", err)
	}
	if !running {
		t.Fatalf("expected php-8.5 to be running after toggle")
	}

	stopped, err := app.ToggleService("php-8.5")
	t.Logf("ToggleService(php-8.5) stop: running=%v, err=%v", stopped, err)
	if err != nil {
		t.Fatalf("unexpected error stopping php-8.5: %v", err)
	}
	if stopped {
		t.Fatalf("expected php-8.5 to be stopped after second toggle")
	}
}

func TestWaitForPort(t *testing.T) {
	app := &App{}

	// 1. Waiting for an unused port to listen with short timeout should return false
	unusedPort := 59123
	if app.waitForPort(unusedPort, true, 200*time.Millisecond) {
		t.Errorf("expected waitForPort(%d, true) to return false for unused port", unusedPort)
	}

	// 2. Waiting for an unused port to be closed should return true immediately
	if !app.waitForPort(unusedPort, false, 200*time.Millisecond) {
		t.Errorf("expected waitForPort(%d, false) to return true for already-closed port", unusedPort)
	}

	// 3. Start an actual TCP listener on a free port and test wait logic
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen on tcp: %v", err)
	}
	defer ln.Close()
	realPort := ln.Addr().(*net.TCPAddr).Port

	// Waiting for realPort to be listening should return true
	if !app.waitForPort(realPort, true, 500*time.Millisecond) {
		t.Errorf("expected waitForPort(%d, true) to return true for active listener", realPort)
	}

	// Close the listener
	_ = ln.Close()

	// Waiting for realPort to be closed should return true
	if !app.waitForPort(realPort, false, 1*time.Second) {
		t.Errorf("expected waitForPort(%d, false) to return true after closing listener", realPort)
	}
}

func TestEnsurePHPServiceRunning_UserStopped(t *testing.T) {
	app := NewApp()
	t.Cleanup(func() { _ = app.supervisor.StopAll() })
	app.markUserStopped("php-8.5")

	if !app.isUserStopped("php-8.5") {
		t.Fatalf("expected php-8.5 to be marked user stopped")
	}

	// Calling ensurePHPServiceRunning for 8.5 should be a no-op (return nil without starting)
	err := app.ensurePHPServiceRunning("8.5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	st, _ := app.supervisor.GetServiceStatus("php-8.5")
	if st.Running {
		t.Fatalf("expected php-8.5 to NOT be running when user stopped")
	}

	// Clearing userStopped allows starting
	app.clearUserStopped("php-8.5")
	if app.isUserStopped("php-8.5") {
		t.Fatalf("expected php-8.5 user stopped to be cleared")
	}
}

func TestReloadCaddyAtomicWrite(t *testing.T) {
	tempDir := t.TempDir()
	t.Chdir(tempDir)

	cfg := config.NewDefaultConfig()
	cfg.ProjectsPath = tempDir

	app := &App{
		cfg:        cfg,
		supervisor: process.NewSupervisor(),
		sites: []caddy.SiteDef{
			{
				Name:        "testsite",
				Domain:      "testsite.test",
				Path:        tempDir,
				RootPath:    tempDir,
				PHPPort:     9083,
				PHPVersion:  "8.3",
				ProjectType: "laravel",
				IsWeb:       true,
			},
		},
	}

	err := app.ReloadCaddy()
	if err != nil {
		t.Fatalf("ReloadCaddy failed: %v", err)
	}

	caddyfilePath := filepath.Join("data", "caddy", "Caddyfile")
	content, err := os.ReadFile(caddyfilePath)
	if err != nil {
		t.Fatalf("expected Caddyfile at %s: %v", caddyfilePath, err)
	}

	if !strings.Contains(string(content), "testsite.test") {
		t.Errorf("expected Caddyfile to contain testsite.test, got:\n%s", string(content))
	}

	tmpPath := filepath.Join("data", "caddy", "Caddyfile.tmp")
	if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
		t.Errorf("temporary file %s still exists", tmpPath)
	}
}

func TestReloadCaddy_IncludesPhpMyAdminWhenInstalled(t *testing.T) {
	tempDir := t.TempDir()
	t.Chdir(tempDir)

	// Create fake data/phpmyadmin/index.php
	pmaDir := filepath.Join(tempDir, "data", "phpmyadmin")
	if err := os.MkdirAll(pmaDir, 0755); err != nil {
		t.Fatalf("failed to create fake phpmyadmin dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pmaDir, "index.php"), []byte("<?php // phpmyadmin"), 0644); err != nil {
		t.Fatalf("failed to create fake phpmyadmin index.php: %v", err)
	}

	cfg := config.NewDefaultConfig()
	cfg.ProjectsPath = tempDir

	app := &App{
		cfg:        cfg,
		supervisor: process.NewSupervisor(),
		sites:      []caddy.SiteDef{},
	}

	if err := app.ReloadCaddy(); err != nil {
		t.Fatalf("ReloadCaddy failed: %v", err)
	}

	caddyfilePath := filepath.Join("data", "caddy", "Caddyfile")
	content, err := os.ReadFile(caddyfilePath)
	if err != nil {
		t.Fatalf("expected Caddyfile at %s: %v", caddyfilePath, err)
	}

	caddyStr := string(content)
	if !strings.Contains(caddyStr, "phpmyadmin.test") {
		t.Errorf("expected Caddyfile to contain phpmyadmin.test when installed, got:\n%s", caddyStr)
	}
	if !strings.Contains(caddyStr, "3308") {
		t.Errorf("expected Caddyfile to contain port 3308 for phpmyadmin, got:\n%s", caddyStr)
	}
}

func TestOpenPhpMyAdmin_NotInstalled(t *testing.T) {
	tempDir := t.TempDir()
	t.Chdir(tempDir)

	app := &App{
		cfg: config.NewDefaultConfig(),
	}

	err := app.OpenPhpMyAdmin("")
	if err == nil {
		t.Fatalf("expected error when phpMyAdmin is not installed, got nil")
	}
}
