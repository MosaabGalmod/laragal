package runtimes_test

import (
	"context"
	"encoding/hex"
	"strings"
	"testing"
	"time"

	"github.com/laragal/laragal/internal/downloader"
	"github.com/laragal/laragal/internal/runtimes"
	"github.com/laragal/laragal/internal/updater"
)

func TestCatalogIntegrity(t *testing.T) {
	all := runtimes.All()
	if len(all) == 0 {
		t.Fatal("expected runtime catalog to not be empty")
	}

	requiredRuntimes := []string{
		"caddy",
		"mailpit",
		"php-8.5",
		"php-8.4",
		"php-8.3",
		"php-8.2",
		"php-8.1",
		"mysql",
		"redis",
		"adminer",
		"phpmyadmin",
	}

	foundMap := make(map[string]runtimes.Runtime)
	for _, rt := range all {
		foundMap[rt.Name] = rt

		if rt.Name == "" {
			t.Errorf("runtime entry has empty name: %+v", rt)
		}
		if rt.Version == "" {
			t.Errorf("runtime %s has empty version", rt.Name)
		}
		if !strings.HasPrefix(rt.URL, "https://") {
			t.Errorf("runtime %s URL must start with https://, got %s", rt.Name, rt.URL)
		}
		if len(rt.SHA256) != 64 {
			t.Errorf("runtime %s SHA256 must be 64 characters long, got %d (%s)", rt.Name, len(rt.SHA256), rt.SHA256)
		}
		if _, err := hex.DecodeString(rt.SHA256); err != nil {
			t.Errorf("runtime %s SHA256 must be valid hex: %v", rt.Name, err)
		}
	}

	for _, req := range requiredRuntimes {
		if _, ok := foundMap[req]; !ok {
			t.Errorf("required runtime %q not found in catalog", req)
		}
	}
}

func TestPHPPortsUnique(t *testing.T) {
	all := runtimes.All()
	seenPorts := make(map[int]string)
	for _, rt := range all {
		if strings.HasPrefix(rt.Name, "php-") {
			if rt.PHPPort <= 0 {
				t.Errorf("PHP runtime %s must have positive PHPPort, got %d", rt.Name, rt.PHPPort)
				continue
			}
			if prev, exists := seenPorts[rt.PHPPort]; exists {
				t.Errorf("duplicate PHPPort %d for %s and %s", rt.PHPPort, prev, rt.Name)
			}
			seenPorts[rt.PHPPort] = rt.Name
		} else {
			if rt.PHPPort != 0 {
				t.Errorf("non-PHP runtime %s must have PHPPort 0, got %d", rt.Name, rt.PHPPort)
			}
		}
	}
}

func TestDownloaderMatchesCheckerURL(t *testing.T) {
	all := runtimes.All()
	for _, rt := range all {
		downloaderURL, err := downloader.GetDownloadURL(rt.Name)
		if err != nil {
			t.Fatalf("downloader error for %s: %v", rt.Name, err)
		}
		checkerURL := updater.GetDownloadURL(rt.Name)
		if checkerURL == "" {
			t.Fatalf("checker returned empty URL for %s", rt.Name)
		}
		if downloaderURL != checkerURL {
			t.Errorf("URL mismatch for %s: downloader=%s, checker=%s", rt.Name, downloaderURL, checkerURL)
		}
	}
}

func TestDownloaderMatchesCheckUpdatesFallbackURL(t *testing.T) {
	opts := updater.CheckOptions{
		HTTPTimeout:    10 * time.Millisecond,
		PHPReleasesURL: "http://127.0.0.1:59999/unreachable",
	}
	updates := updater.CheckUpdates(context.Background(), opts)
	updateMap := make(map[string]updater.RuntimeUpdateInfo)
	for _, u := range updates {
		updateMap[u.Name] = u
	}

	for _, rt := range runtimes.All() {
		downloaderURL, err := downloader.GetDownloadURL(rt.Name)
		if err != nil {
			t.Fatalf("downloader error for %s: %v", rt.Name, err)
		}
		u, ok := updateMap[rt.Name]
		if !ok {
			t.Errorf("CheckUpdates missing runtime %s", rt.Name)
			continue
		}
		if downloaderURL != u.DownloadURL {
			t.Errorf("CheckUpdates URL mismatch for %s: downloader=%s, checker=%s", rt.Name, downloaderURL, u.DownloadURL)
		}
	}
}

func TestPortMapping(t *testing.T) {
	mapping := runtimes.PHPPortMapping()

	expectedPorts := map[string]int{
		"8.1": 9081,
		"8.2": 9082,
		"8.3": 9083,
		"8.4": 9084,
		"8.5": 9085,
	}

	for ver, wantPort := range expectedPorts {
		gotPort, ok := mapping[ver]
		if !ok {
			t.Errorf("mapping missing version %s", ver)
		} else if gotPort != wantPort {
			t.Errorf("mapping[%q] = %d; want %d", ver, gotPort, wantPort)
		}

		if p := runtimes.PHPPortForVersion(ver); p != wantPort {
			t.Errorf("PHPPortForVersion(%q) = %d; want %d", ver, p, wantPort)
		}
		if p := runtimes.PHPPortForVersion("php-" + ver); p != wantPort {
			t.Errorf("PHPPortForVersion(%q) = %d; want %d", "php-"+ver, p, wantPort)
		}
	}

	// Dynamic calculation for higher versions
	if p := runtimes.PHPPortForVersion("8.6"); p != 9086 {
		t.Errorf("PHPPortForVersion(8.6) = %d; want 9086", p)
	}
}

func TestEOLFlags(t *testing.T) {
	php81, ok := runtimes.Get("php-8.1")
	if !ok {
		t.Fatal("php-8.1 not found in catalog")
	}
	if !php81.EOL {
		t.Errorf("php-8.1 must have EOL=true, got false")
	}
	if php81.EOLDate != "2025-12-31" {
		t.Errorf("php-8.1 EOLDate expected 2025-12-31, got %s", php81.EOLDate)
	}

	php82, ok := runtimes.Get("php-8.2")
	if !ok {
		t.Fatal("php-8.2 not found in catalog")
	}
	if php82.EOL {
		t.Errorf("php-8.2 must have EOL=false (security-only until 2026-12-31), got true")
	}
	if php82.EOLDate != "2026-12-31" {
		t.Errorf("php-8.2 EOLDate expected 2026-12-31, got %s", php82.EOLDate)
	}

	for _, v := range []string{"php-8.3", "php-8.4", "php-8.5"} {
		rt, ok := runtimes.Get(v)
		if !ok {
			t.Errorf("%s not found in catalog", v)
			continue
		}
		if rt.EOL {
			t.Errorf("%s must not be EOL", v)
		}
	}

	redis, ok := runtimes.Get("redis")
	if !ok {
		t.Fatal("redis not found in catalog")
	}
	if !redis.EOL {
		t.Errorf("redis must have EOL=true, got false")
	}
	if redis.EOLDate != "2022-02-17" {
		t.Errorf("redis EOLDate expected 2022-02-17, got %s", redis.EOLDate)
	}
}
