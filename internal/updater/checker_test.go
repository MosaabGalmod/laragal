package updater_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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
