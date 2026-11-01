package updater_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/laragal/laragal/internal/updater"
)

func TestGitHubSource_DigestParsed(t *testing.T) {
	const expectedDigest = "0a1edc0b799512051c57071ce0e798d3f2cf67dc98171366f1d2b326072e3b06"
	fakeJSON := fmt.Sprintf(`{
		"tag_name": "v2.11.7",
		"name": "v2.11.7",
		"assets": [
			{
				"name": "caddy_2.11.7_windows_amd64.zip",
				"browser_download_url": "https://example.com/caddy.zip",
				"digest": "sha256:%s"
			}
		]
	}`, expectedDigest)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(fakeJSON))
	}))
	defer server.Close()

	src := updater.NewGitHubSource("caddyserver/caddy", regexp.MustCompile(`^caddy_.*_windows_amd64\.zip$`), "caddy")
	src.BaseURL = server.URL
	src.CacheTTL = -1

	releases, err := src.Latest(context.Background(), "caddy")
	if err != nil {
		t.Fatalf("Latest failed: %v", err)
	}
	if len(releases) != 1 {
		t.Fatalf("expected 1 release, got %d", len(releases))
	}
	rel := releases[0]
	if rel.Version != "2.11.7" {
		t.Errorf("expected version 2.11.7, got %s", rel.Version)
	}
	if rel.ChecksumAlgo != "sha256" {
		t.Errorf("expected sha256 checksum algo, got %s", rel.ChecksumAlgo)
	}
	if rel.Checksum != expectedDigest {
		t.Errorf("expected checksum %s, got %s", expectedDigest, rel.Checksum)
	}
	if rel.MajorChange {
		t.Errorf("expected MajorChange=false for caddy 2.11.7, got true")
	}
}

func TestGitHubSource_AssetSelection(t *testing.T) {
	// Test Redis selection: picks plain msys2, ignores cygwin and with-Service
	redisJSON := `{
		"tag_name": "8.10.3",
		"assets": [
			{
				"name": "Redis-8.10.3-Windows-x64-cygwin-with-Service.zip",
				"browser_download_url": "https://example.com/cygwin-service.zip",
				"digest": "sha256:1111"
			},
			{
				"name": "Redis-8.10.3-Windows-x64-cygwin.zip",
				"browser_download_url": "https://example.com/cygwin.zip",
				"digest": "sha256:2222"
			},
			{
				"name": "Redis-8.10.3-Windows-x64-msys2-with-Service.zip",
				"browser_download_url": "https://example.com/msys2-service.zip",
				"digest": "sha256:3333"
			},
			{
				"name": "Redis-8.10.3-Windows-x64-msys2.zip",
				"browser_download_url": "https://example.com/redis-msys2.zip",
				"digest": "sha256:4444"
			}
		]
	}`

	redisServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(redisJSON))
	}))
	defer redisServer.Close()

	redisSrc := updater.NewGitHubSource("redis-windows/redis-windows", regexp.MustCompile(`^Redis-.*-Windows-x64-msys2\.zip$`), "redis")
	redisSrc.BaseURL = redisServer.URL
	redisSrc.CacheTTL = -1

	releases, err := redisSrc.Latest(context.Background(), "redis")
	if err != nil {
		t.Fatalf("redis Latest failed: %v", err)
	}
	if len(releases) != 1 {
		t.Fatalf("expected 1 release, got %d", len(releases))
	}
	if releases[0].URL != "https://example.com/redis-msys2.zip" {
		t.Errorf("expected plain msys2 zip URL, got %s", releases[0].URL)
	}
	if releases[0].Checksum != "4444" {
		t.Errorf("expected checksum 4444, got %s", releases[0].Checksum)
	}
	// Provider changed from tporadowski and major changed from 5 to 8 -> MajorChange must be true
	if !releases[0].MajorChange {
		t.Errorf("expected MajorChange=true for redis-windows 8.10.3, got false")
	}

	// Test Adminer selection: picks -mysql.php, ignores plain .php and -en.php
	adminerJSON := `{
		"tag_name": "v6.2.0",
		"assets": [
			{
				"name": "adminer-6.2.0.php",
				"browser_download_url": "https://example.com/adminer.php",
				"digest": "sha256:aaaa"
			},
			{
				"name": "adminer-6.2.0-en.php",
				"browser_download_url": "https://example.com/adminer-en.php",
				"digest": "sha256:bbbb"
			},
			{
				"name": "adminer-6.2.0-mysql.php",
				"browser_download_url": "https://example.com/adminer-mysql.php",
				"digest": "sha256:cccc"
			}
		]
	}`

	adminerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(adminerJSON))
	}))
	defer adminerServer.Close()

	adminerSrc := updater.NewGitHubSource("vrana/adminer", regexp.MustCompile(`^adminer-.*-mysql\.php$`), "adminer")
	adminerSrc.BaseURL = adminerServer.URL
	adminerSrc.CacheTTL = -1

	aReleases, err := adminerSrc.Latest(context.Background(), "adminer")
	if err != nil {
		t.Fatalf("adminer Latest failed: %v", err)
	}
	if len(aReleases) != 1 {
		t.Fatalf("expected 1 release, got %d", len(aReleases))
	}
	if aReleases[0].URL != "https://example.com/adminer-mysql.php" {
		t.Errorf("expected -mysql.php URL, got %s", aReleases[0].URL)
	}
	// Major changed from 4 to 6 -> MajorChange must be true
	if !aReleases[0].MajorChange {
		t.Errorf("expected MajorChange=true for adminer 6.2.0, got false")
	}
}

func TestGitHubSource_RateLimitError(t *testing.T) {
	for _, code := range []int{http.StatusForbidden, http.StatusTooManyRequests} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(code)
			_, _ = w.Write([]byte(`{"message": "rate limit exceeded"}`))
		}))

		src := updater.NewGitHubSource("caddyserver/caddy", regexp.MustCompile(`.*`), "caddy")
		src.BaseURL = server.URL
		src.CacheTTL = -1

		_, err := src.Latest(context.Background(), "caddy")
		if err == nil {
			t.Errorf("expected error for HTTP %d, got nil", code)
		}
		server.Close()
	}
}

func TestGitHubSource_Caching(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"tag_name": "v1.33.0",
			"assets": [
				{
					"name": "mailpit-windows-amd64.zip",
					"browser_download_url": "https://example.com/mailpit.zip",
					"digest": "sha256:feedbeef"
				}
			]
		}`))
	}))
	defer server.Close()

	src := updater.NewGitHubSource("axllent/mailpit", regexp.MustCompile(`^mailpit-windows-amd64\.zip$`), "mailpit")
	src.BaseURL = server.URL
	src.CacheTTL = 1 * time.Hour

	// Call 1: hits network
	res1, err := src.Latest(context.Background(), "mailpit")
	if err != nil {
		t.Fatalf("call 1 failed: %v", err)
	}
	if calls != 1 {
		t.Errorf("expected 1 network call, got %d", calls)
	}

	// Call 2: served from cache within 1 hour
	res2, err := src.Latest(context.Background(), "mailpit")
	if err != nil {
		t.Fatalf("call 2 failed: %v", err)
	}
	if calls != 1 {
		t.Errorf("expected still 1 network call due to cache, got %d", calls)
	}
	if res1[0].Checksum != res2[0].Checksum {
		t.Errorf("cached result mismatch")
	}
}

func TestMajorChange_VersionComparison(t *testing.T) {
	// Adminer 4.8.1 -> 6.2.0: Major change
	if !updater.IsMajorChange("4.8.1", "6.2.0") {
		t.Errorf("expected 4.8.1 -> 6.2.0 to be major change")
	}

	// Caddy 2.11.7 -> 2.12.0: Not major change
	if updater.IsMajorChange("2.11.7", "2.12.0") {
		t.Errorf("expected 2.11.7 -> 2.12.0 NOT to be major change")
	}

	// Redis 5.0.14.1 -> 8.10.3: Major change
	if !updater.IsMajorChange("5.0.14.1", "8.10.3") {
		t.Errorf("expected 5.0.14.1 -> 8.10.3 to be major change")
	}
}
