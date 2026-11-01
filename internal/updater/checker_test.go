package updater_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/laragal/laragal/internal/updater"
)

func TestParsePHPReleasesJSON(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "releases.json"))
	if err != nil {
		t.Fatalf("failed to read testdata/releases.json: %v", err)
	}

	releases, err := updater.ParsePHPReleasesJSON(data)
	if err != nil {
		t.Fatalf("ParsePHPReleasesJSON failed: %v", err)
	}

	meta84, ok := releases["8.4"]
	if !ok {
		t.Fatalf("expected 8.4 in releases")
	}
	if meta84.Version != "8.4.27-1" {
		t.Errorf("expected 8.4.27-1, got %s", meta84.Version)
	}
	const expected84URL = "https://windows.php.net/downloads/releases/php-8.4.27-1-nts-Win32-vs17-x64.zip"
	if meta84.ZipURL != expected84URL {
		t.Errorf("expected %s, got %s", expected84URL, meta84.ZipURL)
	}
	const expected84SHA = "dbae154a0a89791f36c3d48e68bae797878333e7d6cb5721e74617de59b63ab7"
	if meta84.SHA256 != expected84SHA {
		t.Errorf("expected %s, got %s", expected84SHA, meta84.SHA256)
	}
	if meta84.IsSecurity {
		t.Errorf("expected IsSecurity to be false")
	}

	// 8.1 uses nts-vs16-x64
	meta81, ok := releases["8.1"]
	if !ok {
		t.Fatalf("expected 8.1 in releases")
	}
	if meta81.Version != "8.1.34" {
		t.Errorf("expected 8.1.34, got %s", meta81.Version)
	}
	const expected81URL = "https://windows.php.net/downloads/releases/php-8.1.34-nts-Win32-vs16-x64.zip"
	if meta81.ZipURL != expected81URL {
		t.Errorf("expected %s, got %s", expected81URL, meta81.ZipURL)
	}
	const expected81SHA = "9cfe246cb144076c16f5913a3ef88a474c3dd7e60f0f0c8bb95faf68674016cc"
	if meta81.SHA256 != expected81SHA {
		t.Errorf("expected %s, got %s", expected81SHA, meta81.SHA256)
	}
}

func TestCheckUpdates_WithMockServer(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "releases.json"))
	if err != nil {
		t.Fatalf("failed to read testdata/releases.json: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(data)
	}))
	defer server.Close()

	tempBin := t.TempDir()
	phpDir := filepath.Join(tempBin, "php-8.4")
	_ = os.MkdirAll(phpDir, 0755)

	opts := updater.CheckOptions{
		PHPReleasesURL: server.URL,
		HTTPTimeout:    2 * time.Second,
		BaseBinDir:     tempBin,
		InstalledVersions: map[string]string{
			"php-8.4": "8.4.26", // Current installed is older than 8.4.27-1
			"caddy":   "2.8.4",
		},
	}

	updates := updater.CheckUpdates(context.Background(), opts)
	if len(updates) == 0 {
		t.Fatalf("expected updates result, got empty list")
	}

	var phpUpdate *updater.RuntimeUpdateInfo
	for i := range updates {
		if updates[i].Name == "php-8.4" {
			phpUpdate = &updates[i]
			break
		}
	}

	if phpUpdate == nil {
		t.Fatalf("expected update entry for php-8.4")
	}
	if !phpUpdate.HasUpdate {
		t.Errorf("expected HasUpdate to be true for php-8.4 (8.4.26 vs 8.4.27-1)")
	}
	if phpUpdate.LatestVersion != "8.4.27-1" {
		t.Errorf("expected latest version 8.4.27-1, got %s", phpUpdate.LatestVersion)
	}
	const expected84SHA = "dbae154a0a89791f36c3d48e68bae797878333e7d6cb5721e74617de59b63ab7"
	if phpUpdate.SHA256 != expected84SHA {
		t.Errorf("expected SHA256 %s, got %s", expected84SHA, phpUpdate.SHA256)
	}
}

func TestCheckUpdates_GracefulOfflineFallback(t *testing.T) {
	// Point to unreachable port with immediate timeout
	opts := updater.CheckOptions{
		PHPReleasesURL: "http://127.0.0.1:59999/unreachable",
		HTTPTimeout:    100 * time.Millisecond,
		InstalledVersions: map[string]string{
			"php-8.3": "8.3.16",
		},
	}

	updates := updater.CheckUpdates(context.Background(), opts)
	if len(updates) == 0 {
		t.Errorf("expected graceful fallback updates list even when offline")
	}
}

func TestCheckUpdates_WithNewMinorFromFeed(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "releases.json"))
	if err != nil {
		t.Fatalf("failed to read testdata/releases.json: %v", err)
	}

	const fake86SHA = "e860000000000000000000000000000000000000000000000000000000000086"
	fake86Block := `
 "8.6": {
  "version": "8.6.0-1",
  "nts-vs17-x64": {
   "zip": {
    "path": "php-8.6.0-1-nts-Win32-vs17-x64.zip",
    "sha256": "` + fake86SHA + `"
   }
  }
 },`

	// Insert "8.6" entry right after the opening "{"
	feedJSON := strings.Replace(string(data), "{", "{"+fake86Block, 1)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(feedJSON))
	}))
	defer server.Close()

	opts := updater.CheckOptions{
		PHPReleasesURL: server.URL,
		HTTPTimeout:    2 * time.Second,
	}

	updates := updater.CheckUpdates(context.Background(), opts)
	var found86 *updater.RuntimeUpdateInfo
	for i := range updates {
		if updates[i].Name == "php-8.6" {
			found86 = &updates[i]
			break
		}
	}

	if found86 == nil {
		t.Fatalf("expected CheckUpdates to discover feed-only php-8.6")
	}
	if !found86.IsNew {
		t.Errorf("expected php-8.6 to have IsNew=true, got false")
	}
	if found86.LatestVersion != "8.6.0-1" {
		t.Errorf("expected latest version 8.6.0-1, got %s", found86.LatestVersion)
	}
	if found86.SHA256 != fake86SHA {
		t.Errorf("expected SHA256 %s, got %s", fake86SHA, found86.SHA256)
	}

	for _, u := range updates {
		if u.Name == "php-8.0" {
			t.Errorf("expected updates NOT to contain EOL feed minor php-8.0, but found: %+v", u)
		}
		if u.Name == "php-7.4" {
			t.Errorf("expected updates NOT to contain EOL feed minor php-7.4, but found: %+v", u)
		}
	}
}

func TestPHPSource_Latest(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "releases.json"))
	if err != nil {
		t.Fatalf("failed to read testdata/releases.json: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(data)
	}))
	defer server.Close()

	var src updater.ReleaseSource = updater.NewPHPSource(server.URL)
	releases, err := src.Latest(context.Background(), "php")
	if err != nil {
		t.Fatalf("src.Latest failed: %v", err)
	}
	if len(releases) == 0 {
		t.Fatalf("expected non-empty live releases from PHPSource")
	}

	found84 := false
	for _, rel := range releases {
		if rel.Name == "php-8.4" {
			found84 = true
			if rel.Version != "8.4.27-1" {
				t.Errorf("expected version 8.4.27-1, got %s", rel.Version)
			}
			if rel.ChecksumAlgo != "sha256" {
				t.Errorf("expected ChecksumAlgo sha256, got %s", rel.ChecksumAlgo)
			}
		}
	}
	if !found84 {
		t.Errorf("expected php-8.4 in live releases")
	}
}

func TestCheckUpdates_PHPBuildSuffix_InstalledWithoutSuffix_NoUpdate(t *testing.T) {
	feedJSON := `{
  "8.5": {
    "version": "8.5.12-1",
    "nts-vs17-x64": {
      "zip": {
        "path": "php-8.5.12-1-nts-Win32-vs17-x64.zip",
        "sha256": "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
      }
    }
  }
}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(feedJSON))
	}))
	defer server.Close()

	tempBin := t.TempDir()
	_ = os.MkdirAll(filepath.Join(tempBin, "php-8.5"), 0755)

	opts := updater.CheckOptions{
		PHPReleasesURL: server.URL,
		HTTPTimeout:    2 * time.Second,
		BaseBinDir:     tempBin,
		InstalledVersions: map[string]string{
			"php-8.5": "8.5.12",
		},
	}

	updates := updater.CheckUpdates(context.Background(), opts)
	var phpUpdate *updater.RuntimeUpdateInfo
	for i := range updates {
		if updates[i].Name == "php-8.5" {
			phpUpdate = &updates[i]
			break
		}
	}

	if phpUpdate == nil {
		t.Fatalf("expected update entry for php-8.5")
	}
	if phpUpdate.CurrentVersion != "8.5.12" {
		t.Errorf("expected CurrentVersion 8.5.12, got %s", phpUpdate.CurrentVersion)
	}
	if phpUpdate.LatestVersion != "8.5.12-1" {
		t.Errorf("expected LatestVersion 8.5.12-1, got %s", phpUpdate.LatestVersion)
	}
	if phpUpdate.HasUpdate {
		t.Errorf("expected HasUpdate to be false when installed version 8.5.12 has no suffix and feed version is 8.5.12-1, got true")
	}
}

func TestCheckUpdates_SourceStatus_LiveAndFallback_WithLogging(t *testing.T) {
	resetEpoch := time.Now().Add(25 * time.Minute).Unix()
	expectedResetTime := time.Unix(resetEpoch, 0).Local().Format("15:04")

	// Server 1: Redis failing with 403 Rate Limit and reset header
	redisServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", resetEpoch))
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"message": "API rate limit exceeded"}`))
	}))
	defer redisServer.Close()

	// Server 2: Caddy succeeding with live 200 response
	caddyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"tag_name": "v2.12.0",
			"assets": [
				{
					"name": "caddy_2.12.0_windows_amd64.zip",
					"browser_download_url": "https://example.com/caddy.zip",
					"digest": "sha256:123456"
				}
			]
		}`))
	}))
	defer caddyServer.Close()

	tempRoot := t.TempDir()
	logDir := filepath.Join(tempRoot, "data", "logs")

	redisSrc := updater.NewGitHubSource("redis-windows/redis-windows", regexp.MustCompile(`^Redis-.*-Windows-x64-msys2\.zip$`), "redis")
	redisSrc.BaseURL = redisServer.URL
	redisSrc.CacheTTL = -1

	caddySrc := updater.NewGitHubSource("caddyserver/caddy", regexp.MustCompile(`^caddy_.*_windows_amd64\.zip$`), "caddy")
	caddySrc.BaseURL = caddyServer.URL
	caddySrc.CacheTTL = -1

	opts := updater.CheckOptions{
		LogDir: logDir,
		Registry: updater.Registry{
			"redis": redisSrc,
			"caddy": caddySrc,
		},
	}

	updates := updater.CheckUpdates(context.Background(), opts)
	if len(updates) == 0 {
		t.Fatalf("expected updates result")
	}

	var redisInfo, caddyInfo *updater.RuntimeUpdateInfo
	for i := range updates {
		if updates[i].Name == "redis" {
			redisInfo = &updates[i]
		}
		if updates[i].Name == "caddy" {
			caddyInfo = &updates[i]
		}
	}

	if redisInfo == nil {
		t.Fatalf("expected redis in updates")
	}
	if redisInfo.SourceStatus != "catalog_fallback" {
		t.Errorf("expected redis SourceStatus to be 'catalog_fallback', got '%s'", redisInfo.SourceStatus)
	}
	if !strings.Contains(strings.ToLower(redisInfo.SourceError), "rate limit") {
		t.Errorf("expected redis SourceError to mention rate limit, got '%s'", redisInfo.SourceError)
	}
	if !strings.Contains(redisInfo.SourceError, expectedResetTime) {
		t.Errorf("expected redis SourceError to contain reset time %s, got '%s'", expectedResetTime, redisInfo.SourceError)
	}

	if caddyInfo == nil {
		t.Fatalf("expected caddy in updates")
	}
	if caddyInfo.SourceStatus != "live" {
		t.Errorf("expected caddy SourceStatus to be 'live', got '%s'", caddyInfo.SourceStatus)
	}
	if caddyInfo.SourceError != "" {
		t.Errorf("expected caddy SourceError to be empty on success, got '%s'", caddyInfo.SourceError)
	}

	// Verify data/logs/updater.log was written
	logPath := filepath.Join(logDir, "updater.log")
	logData, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("expected updater.log to exist at %s: %v", logPath, err)
	}
	logStr := string(logData)
	if !strings.Contains(logStr, "[redis]") {
		t.Errorf("expected updater.log to contain [redis], got: %s", logStr)
	}
	if !strings.Contains(strings.ToLower(logStr), "rate limit") {
		t.Errorf("expected updater.log to contain rate limit, got: %s", logStr)
	}
}

func TestCheckUpdates_LogRotation(t *testing.T) {
	tempRoot := t.TempDir()
	logDir := filepath.Join(tempRoot, "data", "logs")
	_ = os.MkdirAll(logDir, 0755)
	logPath := filepath.Join(logDir, "updater.log")

	// Create > 1MB dummy log file
	dummy := make([]byte, 1024*1024+100)
	if err := os.WriteFile(logPath, dummy, 0644); err != nil {
		t.Fatalf("failed to write dummy log: %v", err)
	}

	// Failing server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	src := updater.NewGitHubSource("test/test", regexp.MustCompile(`.*`), "redis")
	src.BaseURL = server.URL
	src.CacheTTL = -1

	opts := updater.CheckOptions{
		LogDir: logDir,
		Registry: updater.Registry{
			"redis": src,
		},
	}

	_ = updater.CheckUpdates(context.Background(), opts)

	fi, err := os.Stat(logPath)
	if err != nil {
		t.Fatalf("failed to stat log file: %v", err)
	}
	if fi.Size() >= 1024*1024 {
		t.Errorf("expected log file to be rotated/truncated under 1MB, got %d bytes", fi.Size())
	}
	content, _ := os.ReadFile(logPath)
	if !strings.Contains(string(content), "[redis]") {
		t.Errorf("expected new log entry to be written after truncation")
	}
}

func TestDetectInstalledBinaryVersion_AdminerAndPhpMyAdmin_SiblingDataDir(t *testing.T) {
	tempRoot := t.TempDir()
	tempBin := filepath.Join(tempRoot, "bin")
	tempData := filepath.Join(tempRoot, "data")
	_ = os.MkdirAll(tempBin, 0755)

	// Adminer test 1: no files present -> should return "" (not hardcoded 4.8.1)
	if v := updater.DetectInstalledBinaryVersion(tempBin, "adminer"); v != "" {
		t.Errorf("expected empty string when adminer not installed, got '%s'", v)
	}

	// Adminer test 2: file with $VERSION = "4.8.1"
	adminerDir := filepath.Join(tempData, "adminer")
	_ = os.MkdirAll(adminerDir, 0755)
	adminerFile := filepath.Join(adminerDir, "index.php")
	_ = os.WriteFile(adminerFile, []byte("<?php $VERSION = \"4.8.1\"; ?>"), 0644)

	if v := updater.DetectInstalledBinaryVersion(tempBin, "adminer"); v != "4.8.1" {
		t.Errorf("expected adminer version '4.8.1', got '%s'", v)
	}

	// Adminer test 3: file with 6.2.0 script header
	_ = os.WriteFile(adminerFile, []byte("<?php echo 'Adminer 6.2.0 Single File Script';"), 0644)
	if v := updater.DetectInstalledBinaryVersion(tempBin, "adminer"); v != "6.2.0" {
		t.Errorf("expected adminer version '6.2.0', got '%s'", v)
	}

	// Adminer test 4: minified adminer.php with * @version 4.8.1
	_ = os.WriteFile(adminerFile, []byte("<?php /** * @version 4.8.1 */ $ia='4.8.1';"), 0644)
	if v := updater.DetectInstalledBinaryVersion(tempBin, "adminer"); v != "4.8.1" {
		t.Errorf("expected adminer version '4.8.1', got '%s'", v)
	}

	// phpMyAdmin test 1: no files present -> should return "" (not hardcoded 5.2.3)
	if v := updater.DetectInstalledBinaryVersion(tempBin, "phpmyadmin"); v != "" {
		t.Errorf("expected empty string when phpmyadmin not installed, got '%s'", v)
	}

	// phpMyAdmin test 2: RELEASE-DATE-5.2.3 file
	pmaDir := filepath.Join(tempData, "phpmyadmin")
	_ = os.MkdirAll(pmaDir, 0755)
	relDateFile := filepath.Join(pmaDir, "RELEASE-DATE-5.2.3")
	_ = os.WriteFile(relDateFile, []byte("2024-05-10"), 0644)

	if v := updater.DetectInstalledBinaryVersion(tempBin, "phpmyadmin"); v != "5.2.3" {
		t.Errorf("expected phpmyadmin version '5.2.3' from RELEASE-DATE-5.2.3, got '%s'", v)
	}
	_ = os.Remove(relDateFile)

	// phpMyAdmin test 3: libraries/classes/Version.php
	versionClassDir := filepath.Join(pmaDir, "libraries", "classes")
	_ = os.MkdirAll(versionClassDir, 0755)
	versionClassFile := filepath.Join(versionClassDir, "Version.php")
	_ = os.WriteFile(versionClassFile, []byte(`<?php
namespace PhpMyAdmin;
final class Version {
    public const VERSION = '5.2.2';
}
`), 0644)

	if v := updater.DetectInstalledBinaryVersion(tempBin, "phpmyadmin"); v != "5.2.2" {
		t.Errorf("expected phpmyadmin version '5.2.2' from Version.php, got '%s'", v)
	}

	// Test passing custom dataDir explicitly
	isolatedData := filepath.Join(tempRoot, "custom_data", "adminer")
	_ = os.MkdirAll(isolatedData, 0755)
	_ = os.WriteFile(filepath.Join(isolatedData, "index.php"), []byte("<?php $VERSION = \"4.9.0\"; ?>"), 0644)
	if v := updater.DetectInstalledBinaryVersion(tempBin, "adminer", filepath.Join(tempRoot, "custom_data")); v != "4.9.0" {
		t.Errorf("expected adminer version '4.9.0' from customDataDir, got '%s'", v)
	}
}

func TestCheckUpdates_RetryAfterFailure_SuccessLive(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"message": "temporary outage"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"tag_name": "v2.10.0",
			"assets": [
				{
					"name": "caddy_2.10.0_windows_amd64.zip",
					"browser_download_url": "https://example.com/caddy.zip",
					"digest": "sha256:fedcba"
				}
			]
		}`))
	}))
	defer server.Close()

	tempRoot := t.TempDir()
	logDir := filepath.Join(tempRoot, "data", "logs")

	src := updater.NewGitHubSource("caddyserver/caddy", regexp.MustCompile(`^caddy_.*_windows_amd64\.zip$`), "caddy")
	src.BaseURL = server.URL
	src.CacheTTL = 1 * time.Hour // cache enabled for successes

	opts := updater.CheckOptions{
		LogDir: logDir,
		Registry: updater.Registry{
			"caddy": src,
		},
	}

	// First check: fails with 503 -> catalog_fallback + SourceError
	res1 := updater.CheckUpdates(context.Background(), opts)
	var caddy1 *updater.RuntimeUpdateInfo
	for i := range res1 {
		if res1[i].Name == "caddy" {
			caddy1 = &res1[i]
			break
		}
	}
	if caddy1 == nil {
		t.Fatalf("expected caddy in first check results")
	}
	if caddy1.SourceStatus != "catalog_fallback" {
		t.Errorf("expected first check SourceStatus to be 'catalog_fallback', got '%s'", caddy1.SourceStatus)
	}
	if caddy1.SourceError == "" {
		t.Errorf("expected first check SourceError to be non-empty")
	}

	// Verify failure was logged to updater.log
	logContent, err := os.ReadFile(filepath.Join(logDir, "updater.log"))
	if err != nil || !strings.Contains(string(logContent), "[caddy]") {
		t.Errorf("expected updater.log to record [caddy] failure, err: %v, content: %s", err, string(logContent))
	}

	// Second check: server now responds with 200 -> retry immediately, success -> live
	res2 := updater.CheckUpdates(context.Background(), opts)
	var caddy2 *updater.RuntimeUpdateInfo
	for i := range res2 {
		if res2[i].Name == "caddy" {
			caddy2 = &res2[i]
			break
		}
	}
	if caddy2 == nil {
		t.Fatalf("expected caddy in second check results")
	}
	if caddy2.SourceStatus != "live" {
		t.Errorf("expected second check SourceStatus to be 'live', got '%s'", caddy2.SourceStatus)
	}
	if caddy2.SourceError != "" {
		t.Errorf("expected second check SourceError to be empty on success, got '%s'", caddy2.SourceError)
	}
	if caddy2.LatestVersion != "2.10.0" {
		t.Errorf("expected live version 2.10.0, got '%s'", caddy2.LatestVersion)
	}
	if calls != 2 {
		t.Errorf("expected 2 server calls (failure not cached, retried), got %d", calls)
	}
}
