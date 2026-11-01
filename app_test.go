package main

import (
	"regexp"
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/laragal/laragal/internal/config"
	"github.com/laragal/laragal/internal/process"
	"github.com/laragal/laragal/internal/runtimes"
	"github.com/laragal/laragal/internal/services/adminer"
	"github.com/laragal/laragal/internal/services/caddy"
	"github.com/laragal/laragal/internal/updater"
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

func TestGetRuntimeCatalog(t *testing.T) {
	app := &App{}
	catalog := app.GetRuntimeCatalog()
	if len(catalog) == 0 {
		t.Fatal("expected non-empty runtime catalog")
	}

	phpPorts := map[string]int{
		"php-8.5": 9085,
		"php-8.4": 9084,
		"php-8.3": 9083,
		"php-8.2": 9082,
		"php-8.1": 9081,
	}

	foundPHP := make(map[string]runtimes.Artifact)
	for _, item := range catalog {
		if strings.HasPrefix(item.Name, "php-") {
			foundPHP[item.Name] = item
		}
	}

	for name, expectedPort := range phpPorts {
		item, exists := foundPHP[name]
		if !exists {
			t.Errorf("expected catalog to contain %s", name)
			continue
		}
		if item.PHPPort != expectedPort {
			t.Errorf("expected %s port to be %d, got %d", name, expectedPort, item.PHPPort)
		}
	}

	php81, ok := foundPHP["php-8.1"]
	if !ok {
		t.Fatal("php-8.1 not found in catalog")
	}
	if !php81.EOL {
		t.Errorf("expected php-8.1 EOL to be true, got false")
	}
	if php81.EOLDate != "2025-12-31" {
		t.Errorf("expected php-8.1 EOLDate to be 2025-12-31, got %s", php81.EOLDate)
	}

	php82, ok := foundPHP["php-8.2"]
	if !ok {
		t.Fatal("php-8.2 not found in catalog")
	}
	if php82.EOL {
		t.Errorf("expected php-8.2 EOL to be false, got true")
	}
	if php82.EOLDate != "2026-12-31" {
		t.Errorf("expected php-8.2 EOLDate to be 2026-12-31, got %s", php82.EOLDate)
	}
}

func createTestZip(t *testing.T, innerPath string, content []byte) ([]byte, string) {
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	w, err := zw.Create(innerPath)
	if err != nil {
		t.Fatalf("failed to create zip entry %s: %v", innerPath, err)
	}
	if _, err := w.Write(content); err != nil {
		t.Fatalf("failed to write zip content: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("failed to close zip: %v", err)
	}
	zipBytes := buf.Bytes()
	hasher := sha256.New()
	hasher.Write(zipBytes)
	hash := hex.EncodeToString(hasher.Sum(nil))
	return zipBytes, hash
}

func TestUpgradeRuntime_PicksFeedVersionAndNTSZip(t *testing.T) {
	feedBinContent := []byte("PHP 8.4.27-1 binary from live releases feed")
	feedZipBytes, feedZipHash := createTestZip(t, "php.exe", feedBinContent)

	rawFeed, err := os.ReadFile(filepath.Join("internal", "updater", "testdata", "releases.json"))
	if err != nil {
		t.Fatalf("failed to read testdata/releases.json: %v", err)
	}
	// Inject the test zip's hash into the real 8.4 release entry
	feedJSON := strings.Replace(
		string(rawFeed),
		"dbae154a0a89791f36c3d48e68bae797878333e7d6cb5721e74617de59b63ab7",
		feedZipHash,
		1,
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/releases.json":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(feedJSON))
		case "/downloads/releases/php-8.4.27-1-nts-Win32-vs17-x64.zip":
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(feedZipBytes)))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(feedZipBytes)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	tempDir := t.TempDir()
	binDir := filepath.Join(tempDir, "bin")
	tmpDir := filepath.Join(tempDir, "tmp")

	app := &App{
		supervisor:     process.NewSupervisor(),
		phpReleasesURL: server.URL + "/releases.json",
		baseBinDir:     binDir,
		tempDir:        tmpDir,
	}

	msg, err := app.UpgradeRuntime("php-8.4")
	if err != nil {
		t.Fatalf("UpgradeRuntime failed: %v", err)
	}
	if msg.Params["name"] != "php-8.4" || msg.Key != "runtime_install_success_path" {
		t.Errorf("unexpected success message: %+v", msg)
	}

	installedExe := filepath.Join(binDir, "php-8.4", "php.exe")
	content, err := os.ReadFile(installedExe)
	if err != nil {
		t.Fatalf("failed to read upgraded binary: %v", err)
	}
	if !bytes.Equal(content, feedBinContent) {
		t.Errorf("expected upgraded binary %q, got %q", string(feedBinContent), string(content))
	}
}

func TestUpgradeRuntime_HashMismatchLeavesBinUntouched(t *testing.T) {
	tamperedContent := []byte("Tampered or corrupted payload")
	tamperedZipBytes, _ := createTestZip(t, "php.exe", tamperedContent)

	rawFeed, err := os.ReadFile(filepath.Join("internal", "updater", "testdata", "releases.json"))
	if err != nil {
		t.Fatalf("failed to read testdata/releases.json: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/releases.json":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(rawFeed) // Serves real hash, but server returns tamperedZipBytes
		case "/downloads/releases/php-8.4.27-1-nts-Win32-vs17-x64.zip":
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(tamperedZipBytes)))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(tamperedZipBytes)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	tempDir := t.TempDir()
	binDir := filepath.Join(tempDir, "bin")
	tmpDir := filepath.Join(tempDir, "tmp")

	// Pre-populate existing installed php.exe
	phpDir := filepath.Join(binDir, "php-8.4")
	_ = os.MkdirAll(phpDir, 0755)
	existingContent := []byte("Original pristine installed php binary")
	installedExe := filepath.Join(phpDir, "php.exe")
	if err := os.WriteFile(installedExe, existingContent, 0755); err != nil {
		t.Fatalf("failed to write original binary: %v", err)
	}

	app := &App{
		supervisor:     process.NewSupervisor(),
		phpReleasesURL: server.URL + "/releases.json",
		baseBinDir:     binDir,
		tempDir:        tmpDir,
	}

	_, err = app.UpgradeRuntime("php-8.4")
	if err == nil {
		t.Fatalf("expected hash mismatch error, got nil")
	}
	if !strings.Contains(err.Error(), "mismatch") {
		t.Errorf("expected error to mention 'mismatch', got: %v", err)
	}

	// Verify installed binary is untouched!
	content, err := os.ReadFile(installedExe)
	if err != nil {
		t.Fatalf("installed binary was deleted or inaccessible: %v", err)
	}
	if !bytes.Equal(content, existingContent) {
		t.Errorf("installed binary was modified! got %q, want %q", string(content), string(existingContent))
	}
}

func TestUpgradeRuntime_FeedDownFallsBackToPinnedCatalog(t *testing.T) {
	pinnedContent := []byte("Pinned catalog PHP binary content")
	pinnedZipBytes, pinnedZipHash := createTestZip(t, "php.exe", pinnedContent)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/releases.json":
			w.WriteHeader(http.StatusInternalServerError)
		case "/pinned-php84.zip":
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(pinnedZipBytes)))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(pinnedZipBytes)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Override catalog for php-8.4 to use our mock server pinned URL
	restore := runtimes.SetForTest("php-8.4", runtimes.Runtime{
		Name:    "php-8.4",
		Version: "8.4.27-1",
		URL:     server.URL + "/pinned-php84.zip",
		SHA256:  pinnedZipHash,
	})
	defer restore()

	tempDir := t.TempDir()
	binDir := filepath.Join(tempDir, "bin")
	tmpDir := filepath.Join(tempDir, "tmp")

	app := &App{
		supervisor:     process.NewSupervisor(),
		phpReleasesURL: server.URL + "/releases.json", // Returns 500
		baseBinDir:     binDir,
		tempDir:        tmpDir,
	}

	msg, err := app.UpgradeRuntime("php-8.4")
	if err != nil {
		t.Fatalf("UpgradeRuntime failed on fallback to catalog: %v", err)
	}
	if msg.Params["name"] != "php-8.4" || msg.Key != "runtime_install_success_path" {
		t.Errorf("unexpected success message: %+v", msg)
	}

	installedExe := filepath.Join(binDir, "php-8.4", "php.exe")
	content, err := os.ReadFile(installedExe)
	if err != nil {
		t.Fatalf("failed to read installed binary: %v", err)
	}
	if !bytes.Equal(content, pinnedContent) {
		t.Errorf("expected binary from catalog %q, got %q", string(pinnedContent), string(content))
	}
}

func TestUpgradeRuntime_NoHashErrors(t *testing.T) {
	app := &App{
		supervisor: process.NewSupervisor(),
	}
	_, err := app.UpgradeRuntime("unknown-runtime-xyz")
	if err == nil {
		t.Fatalf("expected error for unhashed/unknown runtime, got nil")
	}
	if !strings.Contains(err.Error(), "no verified artifact") {
		t.Errorf("expected clear error message, got: %v", err)
	}
}

func TestLiveRelease_PHP86_CheckAndInstall(t *testing.T) {
	feedBinContent := []byte("PHP 8.6.0-1 binary payload from live releases feed")
	// Make a zip with php-cgi.exe and php.exe
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	w1, _ := zw.Create("php.exe")
	_, _ = w1.Write(feedBinContent)
	w2, _ := zw.Create("php-cgi.exe")
	_, _ = w2.Write(feedBinContent)
	_ = zw.Close()
	feedZipBytes := buf.Bytes()

	hasher := sha256.New()
	hasher.Write(feedZipBytes)
	feedZipHash := hex.EncodeToString(hasher.Sum(nil))

	rawFeed, err := os.ReadFile(filepath.Join("internal", "updater", "testdata", "releases.json"))
	if err != nil {
		t.Fatalf("failed to read testdata/releases.json: %v", err)
	}

	fake86Block := `
 "8.6": {
  "version": "8.6.0-1",
  "nts-vs17-x64": {
   "zip": {
    "path": "php-8.6.0-1-nts-Win32-vs17-x64.zip",
    "sha256": "` + feedZipHash + `"
   }
  }
 },`
	feedJSON := strings.Replace(string(rawFeed), "{", "{"+fake86Block, 1)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/releases.json":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(feedJSON))
		case "/downloads/releases/php-8.6.0-1-nts-Win32-vs17-x64.zip":
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(feedZipBytes)))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(feedZipBytes)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	tempDir := t.TempDir()
	binDir := filepath.Join(tempDir, "bin")
	tmpDir := filepath.Join(tempDir, "tmp")

	app := &App{
		supervisor:     process.NewSupervisor(),
		phpReleasesURL: server.URL + "/releases.json",
		baseBinDir:     binDir,
		tempDir:        tmpDir,
	}

	// 1. CheckUpdates returns php-8.6 with IsNew=true and feed sha256
	updates := app.CheckRuntimeUpdates()
	var update86 *updater.RuntimeUpdateInfo
	for i := range updates {
		if updates[i].Name == "php-8.6" {
			update86 = &updates[i]
			break
		}
	}
	if update86 == nil {
		t.Fatalf("expected CheckRuntimeUpdates to return php-8.6")
	}
	if !update86.IsNew {
		t.Errorf("expected php-8.6 IsNew=true, got false")
	}
	if update86.SHA256 != feedZipHash {
		t.Errorf("expected SHA256 %s, got %s", feedZipHash, update86.SHA256)
	}

	// 2. Install of php-8.6 via httptest
	msg, err := app.DownloadBinary("php-8.6")
	if err != nil {
		t.Fatalf("DownloadBinary(php-8.6) failed: %v", err)
	}
	if msg.Params["name"] != "php-8.6" {
		t.Errorf("unexpected message params: %+v", msg)
	}

	// 3. Verify installed executable
	installedExe := filepath.Join(binDir, "php-8.6", "php-cgi.exe")
	content, err := os.ReadFile(installedExe)
	if err != nil {
		t.Fatalf("failed to read installed php-cgi.exe: %v", err)
	}
	if !bytes.Equal(content, feedBinContent) {
		t.Errorf("installed binary content mismatch")
	}

	// 4. Verify port 9086
	port := runtimes.PHPPortForVersion("8.6")
	if port != 9086 {
		t.Errorf("PHPPortForVersion(8.6) = %d, want 9086", port)
	}

	// 5. Verify ini 8.6 has no active zend_extension=opcache
	iniPath := filepath.Join(binDir, "php-8.6", "php.ini")
	iniContent, err := os.ReadFile(iniPath)
	if err != nil {
		t.Fatalf("failed to read php.ini for 8.6: %v", err)
	}
	for _, line := range strings.Split(string(iniContent), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "zend_extension") && strings.Contains(trimmed, "opcache") {
			t.Errorf("expected no active zend_extension=opcache line in php.ini, found: %s", line)
		}
	}
	if !strings.Contains(string(iniContent), ";zend_extension=opcache") {
		t.Errorf("expected ;zend_extension=opcache commented out in 8.6 php.ini")
	}

	// 6. Verify GetServices includes php-8.6 discovered on disk
	services := app.GetServices()
	foundService := false
	for _, s := range services {
		if s.Name == "php-8.6" {
			foundService = true
			break
		}
	}
	if !foundService {
		t.Errorf("expected GetServices to include on-disk php-8.6")
	}
}

func TestLiveRelease_PHP86_HashMismatchFails(t *testing.T) {
	tamperedContent := []byte("Tampered 8.6 binary")
	tamperedZipBytes, _ := createTestZip(t, "php.exe", tamperedContent)

	rawFeed, err := os.ReadFile(filepath.Join("internal", "updater", "testdata", "releases.json"))
	if err != nil {
		t.Fatalf("failed to read testdata/releases.json: %v", err)
	}

	const realHash = "e860000000000000000000000000000000000000000000000000000000000086"
	fake86Block := `
 "8.6": {
  "version": "8.6.0-1",
  "nts-vs17-x64": {
   "zip": {
    "path": "php-8.6.0-1-nts-Win32-vs17-x64.zip",
    "sha256": "` + realHash + `"
   }
  }
 },`
	feedJSON := strings.Replace(string(rawFeed), "{", "{"+fake86Block, 1)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/releases.json":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(feedJSON))
		case "/downloads/releases/php-8.6.0-1-nts-Win32-vs17-x64.zip":
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(tamperedZipBytes)))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(tamperedZipBytes)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	tempDir := t.TempDir()
	binDir := filepath.Join(tempDir, "bin")
	tmpDir := filepath.Join(tempDir, "tmp")

	app := &App{
		supervisor:     process.NewSupervisor(),
		phpReleasesURL: server.URL + "/releases.json",
		baseBinDir:     binDir,
		tempDir:        tmpDir,
	}

	_, err = app.DownloadBinary("php-8.6")
	if err == nil {
		t.Fatalf("expected hash mismatch error on DownloadBinary(php-8.6), got nil")
	}
	if !strings.Contains(err.Error(), "mismatch") {
		t.Errorf("expected error message to mention 'mismatch', got: %v", err)
	}

	// Verify bin/php-8.6 directory does not exist or has no executable
	destDir := filepath.Join(binDir, "php-8.6")
	if _, statErr := os.Stat(filepath.Join(destDir, "php.exe")); !os.IsNotExist(statErr) {
		t.Errorf("expected php.exe not to exist after hash mismatch")
	}
}

func TestLiveRelease_PHP86_FeedDownErrors(t *testing.T) {
	tempDir := t.TempDir()
	binDir := filepath.Join(tempDir, "bin")
	tmpDir := filepath.Join(tempDir, "tmp")

	app := &App{
		supervisor:     process.NewSupervisor(),
		phpReleasesURL: "http://127.0.0.1:59998/unreachable",
		baseBinDir:     binDir,
		tempDir:        tmpDir,
	}

	// php-8.6 is not in catalog and feed is unreachable -> install must fail with clear error
	_, err := app.DownloadBinary("php-8.6")
	if err == nil {
		t.Fatalf("expected error when downloading unpinned runtime with feed down, got nil")
	}
	if !strings.Contains(err.Error(), "unknown") && !strings.Contains(err.Error(), "no verified artifact") {
		t.Errorf("expected clear error message for uncataloged runtime when feed down, got: %v", err)
	}
}

func TestLiveRelease_OlderFeedMinor_Rejected(t *testing.T) {
	tempDir := t.TempDir()
	binDir := filepath.Join(tempDir, "bin")
	tmpDir := filepath.Join(tempDir, "tmp")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rawFeed, _ := os.ReadFile(filepath.Join("internal", "updater", "testdata", "releases.json"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(rawFeed)
	}))
	defer server.Close()

	app := &App{
		supervisor:     process.NewSupervisor(),
		phpReleasesURL: server.URL,
		baseBinDir:     binDir,
		tempDir:        tmpDir,
	}

	_, err := app.DownloadBinary("php-8.0")
	if err == nil {
		t.Fatalf("expected error when attempting to download uncataloged older feed minor php-8.0, got nil")
	}
	if !strings.Contains(err.Error(), "unknown") && !strings.Contains(err.Error(), "no verified artifact") {
		t.Errorf("expected error to mention unknown runtime or no verified artifact, got: %v", err)
	}
}

func TestUIMessage_CheckPortConflict(t *testing.T) {
	app := &App{
		supervisor: process.NewSupervisor(),
	}
	msg := app.CheckPortConflict(39999)
	if msg.Key != "" {
		t.Errorf("expected empty key for unoccupied port, got: %v", msg)
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen on ephemeral port: %v", err)
	}
	defer ln.Close()
	port := ln.Addr().(*net.TCPAddr).Port

	msgOpen := app.CheckPortConflict(port)
	if msgOpen.Key != "port_conflict_warning" {
		t.Errorf("expected port_conflict_warning key, got: %s", msgOpen.Key)
	}
	if msgOpen.Params["port"] != fmt.Sprintf("%d", port) {
		t.Errorf("expected port param %d, got: %s", port, msgOpen.Params["port"])
	}
}

func TestUIMessage_SyncHosts(t *testing.T) {
	app := &App{
		cfg:   config.NewDefaultConfig(),
		sites: []caddy.SiteDef{},
	}
	msg, err := app.SyncHosts()
	if err != nil {
		t.Fatalf("SyncHosts failed: %v", err)
	}
	if msg.Key != "hosts_synced_count" && msg.Key != "hosts_clean_no_sites" {
		t.Errorf("expected hosts_synced_count or hosts_clean_no_sites key, got: %s", msg.Key)
	}
}

func TestUIMessage_DiagnoseProject(t *testing.T) {
	tempDir := t.TempDir()
	app := &App{
		cfg: config.NewDefaultConfig(),
	}
	diag := app.DiagnoseProject(tempDir)
	if len(diag.Issues) == 0 {
		t.Errorf("expected diagnostic issues for empty folder, got 0")
	}
	hasKey := false
	for _, issue := range diag.Issues {
		if issue.Key != "" {
			hasKey = true
			break
		}
	}
	if !hasKey {
		t.Errorf("expected structured UIMessage with key in issues, got %+v", diag.Issues)
	}
}

func TestNoArabicInBackendGoFiles(t *testing.T) {
	arabicRegex := regexp.MustCompile(`[\x{0600}-\x{06FF}]`)

	filesToScan := []string{"app.go"}
	err := filepath.Walk("internal", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		normPath := filepath.ToSlash(path)
		if !strings.HasSuffix(normPath, ".go") || strings.HasSuffix(normPath, "_test.go") {
			return nil
		}
		if strings.HasPrefix(normPath, "internal/services/caddy") {
			return nil
		}
		if strings.HasSuffix(normPath, ".php") {
			return nil
		}
		filesToScan = append(filesToScan, normPath)
		return nil
	})
	if err != nil {
		t.Fatalf("failed to walk internal directory: %v", err)
	}

	for _, file := range filesToScan {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("failed to read file %s: %v", file, err)
		}
		normalized := strings.ReplaceAll(string(data), "\r\n", "\n")
		lines := strings.Split(normalized, "\n")
		for idx, line := range lines {
			if arabicRegex.MatchString(line) {
				t.Errorf("file %s line %d contains Arabic characters: %s", file, idx+1, strings.TrimSpace(line))
			}
		}
	}
}

func TestCheckFrameable(t *testing.T) {
	tests := []struct {
		name          string
		headerKey     string
		headerVal     string
		statusCode    int
		methodAllowed string
		wantFrameable bool
		wantReason    string
	}{
		{
			name:          "DENY header",
			headerKey:     "X-Frame-Options",
			headerVal:     "DENY",
			statusCode:    http.StatusOK,
			wantFrameable: false,
			wantReason:    "x_frame_options",
		},
		{
			name:          "SAMEORIGIN header",
			headerKey:     "X-Frame-Options",
			headerVal:     "SAMEORIGIN",
			statusCode:    http.StatusOK,
			wantFrameable: false,
			wantReason:    "x_frame_options",
		},
		{
			name:          "CSP frame-ancestors self",
			headerKey:     "Content-Security-Policy",
			headerVal:     "default-src 'self'; frame-ancestors 'self'",
			statusCode:    http.StatusOK,
			wantFrameable: false,
			wantReason:    "csp_frame_ancestors",
		},
		{
			name:          "CSP frame-ancestors wildcard",
			headerKey:     "Content-Security-Policy",
			headerVal:     "default-src 'self'; frame-ancestors *",
			statusCode:    http.StatusOK,
			wantFrameable: true,
			wantReason:    "",
		},
		{
			name:          "no framing headers",
			headerKey:     "",
			headerVal:     "",
			statusCode:    http.StatusOK,
			wantFrameable: true,
			wantReason:    "",
		},
		{
			name:          "HEAD 405 fallback to GET with no headers",
			headerKey:     "",
			headerVal:     "",
			statusCode:    http.StatusOK,
			methodAllowed: "GET_ONLY",
			wantFrameable: true,
			wantReason:    "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tc.methodAllowed == "GET_ONLY" && r.Method == http.MethodHead {
					w.WriteHeader(http.StatusMethodNotAllowed)
					return
				}
				if tc.headerKey != "" {
					w.Header().Set(tc.headerKey, tc.headerVal)
				}
				w.WriteHeader(tc.statusCode)
			}))
			defer ts.Close()

			domain := strings.TrimPrefix(ts.URL, "https://")

			app := &App{
				sites: []caddy.SiteDef{
					{Domain: domain},
				},
				checkClient: ts.Client(),
			}

			res, err := app.CheckFrameable(domain)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if res.Frameable != tc.wantFrameable {
				t.Errorf("expected frameable=%v, got %v", tc.wantFrameable, res.Frameable)
			}
			if res.Reason != tc.wantReason {
				t.Errorf("expected reason=%q, got %q", tc.wantReason, res.Reason)
			}
			if res.Status != tc.statusCode {
				t.Errorf("expected status=%d, got %d", tc.statusCode, res.Status)
			}
		})
	}

	t.Run("unknown domain returns error", func(t *testing.T) {
		app := &App{
			sites: []caddy.SiteDef{
				{Domain: "known.test"},
			},
		}
		_, err := app.CheckFrameable("unknown.test")
		if err == nil {
			t.Fatal("expected error for unknown domain, got nil")
		}
	})

	t.Run("unreachable host returns unreachable reason", func(t *testing.T) {
		// Pick an address that is not listening
		app := &App{
			sites: []caddy.SiteDef{
				{Domain: "127.0.0.1:59999"},
			},
		}
		res, err := app.CheckFrameable("127.0.0.1:59999")
		if err != nil {
			t.Fatalf("expected nil error for unreachable host, got %v", err)
		}
		if res.Frameable {
			t.Errorf("expected frameable=false for unreachable, got true")
		}
		if res.Reason != "unreachable" {
			t.Errorf("expected reason=unreachable, got %q", res.Reason)
		}
	})

	t.Run("TLS cert loaded from disk root cert path without checkClient", func(t *testing.T) {
		ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Frame-Options", "DENY")
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		domain := strings.TrimPrefix(ts.URL, "https://")

		// Export ts.Certificate() to a temp PEM file
		certPEM := pem.EncodeToMemory(&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: ts.Certificate().Raw,
		})
		tempCertFile := filepath.Join(t.TempDir(), "root.crt")
		if err := os.WriteFile(tempCertFile, certPEM, 0644); err != nil {
			t.Fatalf("failed to write temp cert: %v", err)
		}

		app := &App{
			sites: []caddy.SiteDef{
				{Domain: domain},
			},
			caddyRootCertPath: tempCertFile,
		}

		res, err := app.CheckFrameable(domain)
		if err != nil {
			t.Fatalf("unexpected error when loading cert from root.crt: %v", err)
		}
		if res.Frameable {
			t.Errorf("expected frameable=false, got true")
		}
		if res.Reason != "x_frame_options" {
			t.Errorf("expected reason=x_frame_options, got %q", res.Reason)
		}
	})
}

func TestGitHub_DigestMismatchFailsInstall(t *testing.T) {
	fakeContent := []byte("Corrupted Caddy Binary")
	zipBytes, _ := createTestZip(t, "caddy.exe", fakeContent)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/release":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"tag_name": "v2.12.0",
				"assets": [
					{
						"name": "caddy_2.12.0_windows_amd64.zip",
						"browser_download_url": "` + r.Header.Get("X-Server-URL") + `/caddy.zip",
						"digest": "sha256:0000000000000000000000000000000000000000000000000000000000000000"
					}
				]
			}`))
		case "/caddy.zip":
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(zipBytes)))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(zipBytes)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	tempDir := t.TempDir()
	binDir := filepath.Join(tempDir, "bin")
	tmpDir := filepath.Join(tempDir, "tmp")

	// Pre-populate existing installed caddy.exe
	caddyDir := filepath.Join(binDir, "caddy")
	_ = os.MkdirAll(caddyDir, 0755)
	existingContent := []byte("Pristine existing caddy.exe")
	if err := os.WriteFile(filepath.Join(caddyDir, "caddy.exe"), existingContent, 0755); err != nil {
		t.Fatalf("failed to write original binary: %v", err)
	}

	caddySrc := updater.NewGitHubSource("caddyserver/caddy", regexp.MustCompile(`^caddy_.*_windows_amd64\.zip$`), "caddy")
	caddySrc.BaseURL = server.URL + "/release"
	caddySrc.CacheTTL = -1
	caddySrc.HTTPClient = &http.Client{
		Transport: &testHeaderTransport{header: "X-Server-URL", value: server.URL},
	}

	app := &App{
		supervisor: process.NewSupervisor(),
		baseBinDir: binDir,
		tempDir:    tmpDir,
		registry: updater.Registry{
			"caddy": caddySrc,
		},
	}

	_, err := app.UpgradeRuntimeConfirmed("caddy")
	if err == nil {
		t.Fatalf("expected digest mismatch error, got nil")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "mismatch") {
		t.Errorf("expected error to mention mismatch, got: %v", err)
	}

	// Verify original file untouched
	installedContent, err := os.ReadFile(filepath.Join(caddyDir, "caddy.exe"))
	if err != nil {
		t.Fatalf("failed to read caddy.exe: %v", err)
	}
	if !bytes.Equal(installedContent, existingContent) {
		t.Errorf("expected caddy.exe untouched, got %q", string(installedContent))
	}
}

type testHeaderTransport struct {
	header string
	value  string
}

func (t *testHeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set(t.header, t.value)
	return http.DefaultTransport.RoundTrip(req)
}

func TestGitHub_RateLimitFallbackToCatalog(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"message": "API rate limit exceeded"}`))
	}))
	defer server.Close()

	caddySrc := updater.NewGitHubSource("caddyserver/caddy", regexp.MustCompile(`^caddy_.*_windows_amd64\.zip$`), "caddy")
	caddySrc.BaseURL = server.URL
	caddySrc.CacheTTL = -1

	app := &App{
		supervisor: process.NewSupervisor(),
		registry: updater.Registry{
			"caddy": caddySrc,
		},
	}

	updates := app.CheckRuntimeUpdates()
	var caddyUpdate *updater.RuntimeUpdateInfo
	for i := range updates {
		if updates[i].Name == "caddy" {
			caddyUpdate = &updates[i]
			break
		}
	}

	if caddyUpdate == nil {
		t.Fatalf("expected update entry for caddy")
	}
	// Falls back to catalog version 2.11.7
	if caddyUpdate.LatestVersion != "2.11.7" {
		t.Errorf("expected fallback to catalog version 2.11.7, got %s", caddyUpdate.LatestVersion)
	}
}

func TestGitHub_AdminerMajorBlockedFromPlainUpgrade(t *testing.T) {
	adminerContent := []byte("<?php echo 'Adminer 6.2.0 Single File Script';")
	hasher := sha256.New()
	hasher.Write(adminerContent)
	adminerHash := hex.EncodeToString(hasher.Sum(nil))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/release":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"tag_name": "v6.2.0",
				"assets": [
					{
						"name": "adminer-6.2.0.php",
						"browser_download_url": "` + r.Header.Get("X-Server-URL") + `/adminer-plain.php",
						"digest": "sha256:1111"
					},
					{
						"name": "adminer-6.2.0-mysql.php",
						"browser_download_url": "` + r.Header.Get("X-Server-URL") + `/adminer-mysql.php",
						"digest": "sha256:` + adminerHash + `"
					}
				]
			}`))
		case "/adminer-mysql.php":
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(adminerContent)))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(adminerContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	tempDir := t.TempDir()
	binDir := filepath.Join(tempDir, "bin")
	tmpDir := filepath.Join(tempDir, "tmp")
	dataDir := filepath.Join(tempDir, "data")

	adminerSrc := updater.NewGitHubSource("vrana/adminer", regexp.MustCompile(`^adminer-.*-mysql\.php$`), "adminer")
	adminerSrc.BaseURL = server.URL + "/release"
	adminerSrc.CacheTTL = -1
	adminerSrc.HTTPClient = &http.Client{
		Transport: &testHeaderTransport{header: "X-Server-URL", value: server.URL},
	}

	app := &App{
		supervisor: process.NewSupervisor(),
		baseBinDir: binDir,
		tempDir:    tmpDir,
		dataDir:    dataDir,
		registry: updater.Registry{
			"adminer": adminerSrc,
		},
	}

	// 1. Verify CheckRuntimeUpdates marks MajorChange=true
	updates := app.CheckRuntimeUpdates()
	var adminerInfo *updater.RuntimeUpdateInfo
	for i := range updates {
		if updates[i].Name == "adminer" {
			adminerInfo = &updates[i]
			break
		}
	}
	if adminerInfo == nil {
		t.Fatalf("expected adminer in updates")
	}
	if !adminerInfo.MajorChange {
		t.Errorf("expected adminer MajorChange=true (4.8.1 -> 6.2.0), got false")
	}

	// 2. Plain UpgradeRuntime must be BLOCKED with major_change_requires_confirmation
	_, err := app.UpgradeRuntime("adminer")
	if err == nil {
		t.Fatalf("expected UpgradeRuntime('adminer') to fail with major change confirmation error, got nil")
	}
	if err.Error() != "major_change_requires_confirmation" {
		t.Errorf("expected 'major_change_requires_confirmation', got: %v", err)
	}

	// 3. UpgradeRuntimeConfirmed must SUCCEED
	msg, err := app.UpgradeRuntimeConfirmed("adminer")
	if err != nil {
		t.Fatalf("UpgradeRuntimeConfirmed failed: %v", err)
	}
	if msg.Key != "adminer_download_success" {
		t.Errorf("unexpected msg key: %+v", msg)
	}

	// Verify data/adminer/adminer-core.php exists and has downloaded content
	coreFile := filepath.Join(dataDir, "adminer", "adminer-core.php")
	content, err := os.ReadFile(coreFile)
	if err != nil {
		t.Fatalf("failed to read adminer-core.php: %v", err)
	}
	if !bytes.Equal(content, adminerContent) {
		t.Errorf("adminer-core.php content mismatch: %q", string(content))
	}
}

func TestGitHub_CaddyMinorNotMajor_PlainUpgradeSucceeds(t *testing.T) {
	caddyContent := []byte("Caddy 2.12.0 binary")
	zipBytes, zipHash := createTestZip(t, "caddy.exe", caddyContent)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/release":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"tag_name": "v2.12.0",
				"assets": [
					{
						"name": "caddy_2.12.0_windows_amd64.zip",
						"browser_download_url": "` + r.Header.Get("X-Server-URL") + `/caddy.zip",
						"digest": "sha256:` + zipHash + `"
					}
				]
			}`))
		case "/caddy.zip":
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(zipBytes)))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(zipBytes)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	tempDir := t.TempDir()
	binDir := filepath.Join(tempDir, "bin")
	tmpDir := filepath.Join(tempDir, "tmp")

	caddySrc := updater.NewGitHubSource("caddyserver/caddy", regexp.MustCompile(`^caddy_.*_windows_amd64\.zip$`), "caddy")
	caddySrc.BaseURL = server.URL + "/release"
	caddySrc.CacheTTL = -1
	caddySrc.HTTPClient = &http.Client{
		Transport: &testHeaderTransport{header: "X-Server-URL", value: server.URL},
	}

	app := &App{
		supervisor: process.NewSupervisor(),
		baseBinDir: binDir,
		tempDir:    tmpDir,
		registry: updater.Registry{
			"caddy": caddySrc,
		},
	}

	// 1. Check updates -> MajorChange must be false
	updates := app.CheckRuntimeUpdates()
	var caddyInfo *updater.RuntimeUpdateInfo
	for i := range updates {
		if updates[i].Name == "caddy" {
			caddyInfo = &updates[i]
			break
		}
	}
	if caddyInfo == nil {
		t.Fatalf("expected caddy update")
	}
	if caddyInfo.MajorChange {
		t.Errorf("expected MajorChange=false for caddy (2.11.7 -> 2.12.0), got true")
	}

	// 2. Plain UpgradeRuntime must SUCCEED directly
	msg, err := app.UpgradeRuntime("caddy")
	if err != nil {
		t.Fatalf("expected plain UpgradeRuntime('caddy') to succeed, got: %v", err)
	}
	if msg.Params["name"] != "caddy" {
		t.Errorf("unexpected message name param: %+v", msg)
	}

	installed := filepath.Join(binDir, "caddy", "caddy.exe")
	content, err := os.ReadFile(installed)
	if err != nil {
		t.Fatalf("failed to read installed caddy.exe: %v", err)
	}
	if !bytes.Equal(content, caddyContent) {
		t.Errorf("caddy.exe content mismatch")
	}
}

func TestGitHub_RedisSuccessor_AssetSelectionAndMajorBlock(t *testing.T) {
	redisContent := []byte("Redis 8.10.3 binary from redis-windows msys2")
	redisConfContent := []byte("# Redis 8.10.3 test configuration\nport 6379\n")

	// Create zip with top-level folder as found in real redis-windows archive
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	wExe, _ := zw.Create("Redis-8.10.3-Windows-x64-msys2/redis-server.exe")
	_, _ = wExe.Write(redisContent)
	wConf, _ := zw.Create("Redis-8.10.3-Windows-x64-msys2/redis.conf")
	_, _ = wConf.Write(redisConfContent)
	_ = zw.Close()
	zipBytes := buf.Bytes()

	hasher := sha256.New()
	hasher.Write(zipBytes)
	zipHash := hex.EncodeToString(hasher.Sum(nil))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/release":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"tag_name": "8.10.3",
				"assets": [
					{
						"name": "Redis-8.10.3-Windows-x64-cygwin-with-Service.zip",
						"browser_download_url": "` + r.Header.Get("X-Server-URL") + `/cygwin-service.zip",
						"digest": "sha256:1111"
					},
					{
						"name": "Redis-8.10.3-Windows-x64-cygwin.zip",
						"browser_download_url": "` + r.Header.Get("X-Server-URL") + `/cygwin.zip",
						"digest": "sha256:2222"
					},
					{
						"name": "Redis-8.10.3-Windows-x64-msys2-with-Service.zip",
						"browser_download_url": "` + r.Header.Get("X-Server-URL") + `/msys2-service.zip",
						"digest": "sha256:3333"
					},
					{
						"name": "Redis-8.10.3-Windows-x64-msys2.zip",
						"browser_download_url": "` + r.Header.Get("X-Server-URL") + `/redis-msys2.zip",
						"digest": "sha256:` + zipHash + `"
					}
				]
			}`))
		case "/redis-msys2.zip":
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(zipBytes)))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(zipBytes)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	tempDir := t.TempDir()
	binDir := filepath.Join(tempDir, "bin")
	tmpDir := filepath.Join(tempDir, "tmp")

	redisSrc := updater.NewGitHubSource("redis-windows/redis-windows", regexp.MustCompile(`^Redis-.*-Windows-x64-msys2\.zip$`), "redis")
	redisSrc.BaseURL = server.URL + "/release"
	redisSrc.CacheTTL = -1
	redisSrc.HTTPClient = &http.Client{
		Transport: &testHeaderTransport{header: "X-Server-URL", value: server.URL},
	}

	app := &App{
		supervisor: process.NewSupervisor(),
		baseBinDir: binDir,
		tempDir:    tmpDir,
		registry: updater.Registry{
			"redis": redisSrc,
		},
	}

	// 1. Check updates -> MajorChange must be true (tporadowski 5.0.14.1 -> redis-windows 8.10.3)
	updates := app.CheckRuntimeUpdates()
	var redisInfo *updater.RuntimeUpdateInfo
	for i := range updates {
		if updates[i].Name == "redis" {
			redisInfo = &updates[i]
			break
		}
	}
	if redisInfo == nil {
		t.Fatalf("expected redis in updates")
	}
	if !redisInfo.MajorChange {
		t.Errorf("expected redis MajorChange=true, got false")
	}

	// 2. Plain UpgradeRuntime must be BLOCKED
	_, err := app.UpgradeRuntime("redis")
	if err == nil {
		t.Fatalf("expected UpgradeRuntime('redis') to fail with major change confirmation error, got nil")
	}
	if err.Error() != "major_change_requires_confirmation" {
		t.Errorf("expected 'major_change_requires_confirmation', got: %v", err)
	}

	// 3. UpgradeRuntimeConfirmed must SUCCEED
	msg, err := app.UpgradeRuntimeConfirmed("redis")
	if err != nil {
		t.Fatalf("UpgradeRuntimeConfirmed failed: %v", err)
	}
	if msg.Params["name"] != "redis" {
		t.Errorf("unexpected message params: %+v", msg)
	}

	// Verify installed binary and config
	installed := filepath.Join(binDir, "redis", "redis-server.exe")
	if _, err := os.Stat(installed); err != nil {
		// Also check subfolder if not stripped
		matches, _ := filepath.Glob(filepath.Join(binDir, "redis", "*", "redis-server.exe"))
		if len(matches) > 0 {
			installed = matches[0]
		}
	}
	content, err := os.ReadFile(installed)
	if err != nil {
		t.Fatalf("failed to find installed redis-server.exe: %v", err)
	}
	if !bytes.Equal(content, redisContent) {
		t.Errorf("redis-server.exe content mismatch")
	}

	// Verify findRedisExecutable finds it
	found := app.findRedisExecutable()
	if found == "" {
		t.Errorf("findRedisExecutable failed to locate the newly installed redis executable")
	}
}

