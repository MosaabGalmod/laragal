package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/laragal/laragal/internal/config"
	"github.com/laragal/laragal/internal/process"
	"github.com/laragal/laragal/internal/services/caddy"
)

func setupTestApp(t *testing.T) (*App, string, string, string) {
	tempDir := t.TempDir()
	binDir := filepath.Join(tempDir, "bin")
	dataDir := filepath.Join(tempDir, "data")
	cfgPath := filepath.Join(tempDir, "laragal.json")

	// Isolate LOCALAPPDATA so unit tests don't pick up host winget packages by default
	t.Setenv("LOCALAPPDATA", filepath.Join(tempDir, "localappdata"))

	_ = os.MkdirAll(binDir, 0755)
	_ = os.MkdirAll(dataDir, 0755)

	cfg := config.NewDefaultConfig()
	cfg.ProjectsPath = filepath.Join(tempDir, "sites")
	cfg.DefaultPHP = "8.3"
	_ = cfg.Save(cfgPath)

	app := &App{
		cfg:        cfg,
		cfgPath:    cfgPath,
		baseBinDir: binDir,
		dataDir:    dataDir,
		supervisor: process.NewSupervisor(),
	}

	return app, tempDir, binDir, dataDir
}

func createManagedPHP(t *testing.T, binDir string, version string) string {
	folder := filepath.Join(binDir, "php-"+version)
	_ = os.MkdirAll(folder, 0755)
	exe := filepath.Join(folder, "php-cgi.exe")
	if err := os.WriteFile(exe, []byte("fake php binary"), 0755); err != nil {
		t.Fatalf("failed to create fake php executable: %v", err)
	}
	return folder
}

func TestUninstallRuntime_RefusesNonPHP(t *testing.T) {
	app, _, _, _ := setupTestApp(t)

	for _, invalid := range []string{"mysql", "mariadb", "caddy", "redis", "mailpit"} {
		_, err := app.UninstallRuntime(invalid)
		if err == nil {
			t.Fatalf("expected error for non-PHP runtime %q, got nil", invalid)
		}
	}
}

func TestUninstallRuntime_RefusesPathTraversal(t *testing.T) {
	app, _, _, _ := setupTestApp(t)

	traversals := []string{
		"php-8.1/../mysql",
		"..\\data",
		"../bin/php-8.1",
		"php-8.1/../../etc",
	}

	for _, path := range traversals {
		_, err := app.UninstallRuntime(path)
		if err == nil {
			t.Fatalf("expected error for path traversal %q, got nil", path)
		}
	}
}

func TestUninstallRuntime_RefusesExternal(t *testing.T) {
	app, _, _, _ := setupTestApp(t)

	// php-8.3 is not created in binDir, so it is external (e.g. WinGet)
	msg, err := app.UninstallRuntime("php-8.3")
	if err == nil {
		t.Fatalf("expected error for external PHP, got nil")
	}

	if msg.Key != "runtime_uninstall_external" {
		t.Errorf("expected Key runtime_uninstall_external, got %q", msg.Key)
	}

	wantCmd := "winget uninstall --id PHP.PHP.8.3"
	if msg.Params["command"] != wantCmd {
		t.Errorf("expected command %q, got %q", wantCmd, msg.Params["command"])
	}

	ue, ok := err.(*UninstallError)
	if !ok || ue.Key != "runtime_uninstall_external" || ue.Params["command"] != wantCmd {
		t.Errorf("expected UninstallError with correct key and params, got: %v", err)
	}
}

func TestUninstallRuntime_RefusesInUse(t *testing.T) {
	app, _, binDir, _ := setupTestApp(t)

	// Create managed php-8.1 and php-8.5 so it is not the last PHP
	createManagedPHP(t, binDir, "8.1")
	createManagedPHP(t, binDir, "8.5")

	// Set sites using 8.1
	app.sitesMu.Lock()
	app.sites = []caddy.SiteDef{
		{Name: "my-app", Domain: "my-app.test", PHPVersion: "8.1"},
		{Name: "blog-app", Domain: "blog.test", PHPVersion: "php-8.1"},
		{Name: "other-app", Domain: "other.test", PHPVersion: "8.5"},
	}
	app.sitesMu.Unlock()

	// Check GetUninstallBlockers
	blockers := app.GetUninstallBlockers("php-8.1")
	if len(blockers) != 2 || blockers[0] != "my-app.test" || blockers[1] != "blog.test" {
		t.Fatalf("unexpected blockers: %v", blockers)
	}

	// Uninstall should refuse with runtime_uninstall_in_use
	msg, err := app.UninstallRuntime("php-8.1")
	if err == nil {
		t.Fatalf("expected refusal when runtime is in use by projects")
	}

	if msg.Key != "runtime_uninstall_in_use" {
		t.Errorf("expected Key runtime_uninstall_in_use, got %q", msg.Key)
	}

	if !strings.Contains(msg.Params["domains"], "my-app.test") || !strings.Contains(msg.Params["domains"], "blog.test") {
		t.Errorf("expected domains in params, got %q", msg.Params["domains"])
	}
}

func TestUninstallRuntime_RefusesInUseFromSitesCache(t *testing.T) {
	app, _, binDir, dataDir := setupTestApp(t)

	createManagedPHP(t, binDir, "8.1")
	createManagedPHP(t, binDir, "8.5")

	// Cache file with sites in app's dataDir
	cachedSites := []caddy.SiteDef{
		{Name: "cached-proj", Domain: "cached.test", PHPVersion: "8.1"},
	}
	cachedData, _ := json.Marshal(cachedSites)
	_ = os.WriteFile(filepath.Join(dataDir, "sites_cache.json"), cachedData, 0644)

	app.sitesMu.Lock()
	app.sites = nil
	app.sitesMu.Unlock()

	blockers := app.GetUninstallBlockers("php-8.1")
	if len(blockers) != 1 || blockers[0] != "cached.test" {
		t.Fatalf("expected cached.test blocker, got %v", blockers)
	}

	msg, err := app.UninstallRuntime("php-8.1")
	if err == nil || msg.Key != "runtime_uninstall_in_use" {
		t.Fatalf("expected refusal for in-use via cache file, got msg=%v, err=%v", msg, err)
	}
}

func TestUninstallRuntime_RefusesLastOrDefault(t *testing.T) {
	app, _, binDir, _ := setupTestApp(t)

	// Case 1: Refuses default PHP in config (config has DefaultPHP: "8.3")
	createManagedPHP(t, binDir, "8.3")
	createManagedPHP(t, binDir, "8.5")
	msg, err := app.UninstallRuntime("php-8.3")
	if err == nil || msg.Key != "runtime_uninstall_last_or_default" {
		t.Fatalf("expected refusal for default PHP, got msg=%v, err=%v", msg, err)
	}

	// Case 2: Refuses last installed PHP
	// Remove 8.3 and leave only 8.5
	_ = os.RemoveAll(filepath.Join(binDir, "php-8.3"))
	app.cfg.DefaultPHP = "8.1" // Change default so 8.5 is not default

	msg, err = app.UninstallRuntime("php-8.5")
	if err == nil || msg.Key != "runtime_uninstall_last_or_default" {
		t.Fatalf("expected refusal for last installed PHP version, got msg=%v, err=%v", msg, err)
	}
}

func TestUninstallRuntime_DeletesManagedAndPreservesSiblingsAndData(t *testing.T) {
	app, _, binDir, dataDir := setupTestApp(t)

	// Create php-8.1 and php-8.5, and some data in dataDir
	dir81 := createManagedPHP(t, binDir, "8.1")
	dir85 := createManagedPHP(t, binDir, "8.5")

	testDataFile := filepath.Join(dataDir, "test.txt")
	_ = os.WriteFile(testDataFile, []byte("data content"), 0644)

	// Set config default to 8.5 so 8.1 is not default
	app.cfg.DefaultPHP = "8.5"

	// Mock supervisor service for 8.1 to verify service stopped
	app.supervisor.SetServiceStatus("php-8.1", process.ServiceStatus{
		Name:    "php-8.1",
		Running: true,
		Port:    9081,
	})
	app.markUserStopped("php-8.1")

	// Uninstall php-8.1
	msg, err := app.UninstallRuntime("php-8.1")
	if err != nil {
		t.Fatalf("unexpected error uninstalling managed php-8.1: %v", err)
	}
	if msg.Key != "runtime_uninstall_success" {
		t.Errorf("expected Key runtime_uninstall_success, got %q", msg.Key)
	}

	// 1. Target directory must be deleted
	if _, err := os.Stat(dir81); !os.IsNotExist(err) {
		t.Errorf("expected %s to be deleted", dir81)
	}

	// 2. Sibling php-8.5 must be untouched
	if _, err := os.Stat(dir85); err != nil {
		t.Errorf("expected sibling %s to remain intact: %v", dir85, err)
	}

	// 3. Data directory must be untouched
	if data, err := os.ReadFile(testDataFile); err != nil || string(data) != "data content" {
		t.Errorf("expected data file to remain untouched: %v", err)
	}

	// 4. Service must be stopped and removed from supervisor
	if st, ok := app.supervisor.GetServiceStatus("php-8.1"); ok && st.Running {
		t.Errorf("expected php-8.1 to be stopped in supervisor")
	}
	if app.isUserStopped("php-8.1") {
		t.Errorf("expected userStopped to be cleared for php-8.1")
	}
}

func TestHiddenRuntimes_GetAndSet(t *testing.T) {
	app, _, _, _ := setupTestApp(t)

	// Initially empty
	hidden := app.GetHiddenRuntimes()
	if len(hidden) != 0 {
		t.Fatalf("expected empty hidden runtimes, got %v", hidden)
	}

	// Hide php-8.2
	if err := app.SetRuntimeHidden("php-8.2", true); err != nil {
		t.Fatalf("failed to hide php-8.2: %v", err)
	}

	hidden = app.GetHiddenRuntimes()
	if len(hidden) != 1 || hidden[0] != "php-8.2" {
		t.Fatalf("expected [php-8.2], got %v", hidden)
	}

	// Hide mariadb_series
	if err := app.SetRuntimeHidden("mariadb_series", true); err != nil {
		t.Fatalf("failed to hide mariadb_series: %v", err)
	}
	hidden = app.GetHiddenRuntimes()
	if len(hidden) != 2 {
		t.Fatalf("expected 2 hidden runtimes, got %v", hidden)
	}

	// Reload config from disk to verify persistence
	loadedCfg, err := config.LoadConfig(app.cfgPath)
	if err != nil {
		t.Fatalf("failed to load saved config: %v", err)
	}
	if len(loadedCfg.HiddenRuntimes) != 2 {
		t.Fatalf("expected 2 hidden runtimes in persisted config, got %v", loadedCfg.HiddenRuntimes)
	}

	// Unhide php-8.2
	if err := app.SetRuntimeHidden("php-8.2", false); err != nil {
		t.Fatalf("failed to unhide php-8.2: %v", err)
	}
	hidden = app.GetHiddenRuntimes()
	if len(hidden) != 1 || hidden[0] != "mariadb_series" {
		t.Fatalf("expected [mariadb_series], got %v", hidden)
	}
}
