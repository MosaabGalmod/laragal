package downloader_test

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/laragal/laragal/internal/downloader"
)

func TestGetDownloadURL(t *testing.T) {
	targets := []string{
		"caddy",
		"mailpit",
		"php-8.5",
		"php-8.4",
		"php-8.3",
		"php-8.2",
		"php-8.1",
		"mysql",
		"redis",
		"php8.5",
		"php85",
		"php-8.6",
		"php-9.0",
		"mariadb",
	}

	for _, target := range targets {
		url, err := downloader.GetDownloadURL(target)
		if err != nil {
			t.Errorf("expected valid url for %s, got error: %v", target, err)
		}
		if !strings.HasPrefix(url, "https://") {
			t.Errorf("url for %s must start with https://, got: %s", target, url)
		}
	}

	const expectedPHP85URL = "https://windows.php.net/downloads/releases/php-8.5.12-1-nts-Win32-vs17-x64.zip"
	for _, target := range []string{"php-8.5", "php85"} {
		url, err := downloader.GetDownloadURL(target)
		if err != nil {
			t.Errorf("GetDownloadURL(%q) unexpected error: %v", target, err)
		}
		if url != expectedPHP85URL {
			t.Errorf("GetDownloadURL(%q) = %q, want %q", target, url, expectedPHP85URL)
		}
		if strings.Contains(url, "8.4") {
			t.Errorf("GetDownloadURL(%q) url %q must not contain '8.4'", target, url)
		}
	}

	_, err := downloader.GetDownloadURL("unknown-runtime-xyz")
	if err == nil {
		t.Errorf("expected error for unknown runtime, got nil")
	}
}

func TestArtifactSHA256Configured(t *testing.T) {
	requiredArtifacts := []string{
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

	for _, name := range requiredArtifacts {
		art, err := downloader.GetDownloadArtifact(name)
		if err != nil {
			t.Errorf("GetDownloadArtifact(%q) returned error: %v", name, err)
			continue
		}
		if art.URL == "" {
			t.Errorf("artifact %s has empty URL", name)
		}
		if len(art.SHA256) != 64 {
			t.Errorf("artifact %s has invalid or missing SHA-256 hash: %q", name, art.SHA256)
		}
		expected := downloader.GetExpectedSHA256(name)
		if expected != art.SHA256 {
			t.Errorf("GetExpectedSHA256(%q) = %q, want %q", name, expected, art.SHA256)
		}
	}
}

func TestDynamicURLWarnPath(t *testing.T) {
	art, err := downloader.GetDownloadArtifact("php-8.6")
	if err != nil {
		t.Fatalf("unexpected error for dynamic php-8.6: %v", err)
	}
	if art.SHA256 != "" {
		t.Errorf("dynamic version should have empty SHA256, got: %s", art.SHA256)
	}
	expected := downloader.GetExpectedSHA256("php-8.6")
	if expected != "" {
		t.Errorf("dynamic version should return empty expected hash, got: %s", expected)
	}

	payload := []byte("Dynamic unhashed binary content")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(payload)))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(payload)
	}))
	defer server.Close()

	tempDir := t.TempDir()
	destFile := filepath.Join(tempDir, "dynamic_php.zip")

	// Dynamic URLs have no known hash; code must not refuse but succeed and leave file in place
	err = downloader.DownloadFile(context.Background(), server.URL, destFile, nil)
	if err != nil {
		t.Fatalf("dynamic download failed unexpectedly: %v", err)
	}
	if _, err := os.Stat(destFile); err != nil {
		t.Fatalf("dynamic download file missing: %v", err)
	}
}

func TestDownloadFileWithProgress(t *testing.T) {
	payload := []byte("Hello LaraGal Downloader Payload with sufficient length to test chunking and byte counts")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(payload)))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(payload)
	}))
	defer server.Close()

	tempDir := t.TempDir()
	destFile := filepath.Join(tempDir, "downloaded.bin")

	var progressReports int
	var lastPercent int

	callback := func(downloaded int64, total int64, percent int) {
		progressReports++
		lastPercent = percent
	}

	err := downloader.DownloadFile(context.Background(), server.URL, destFile, callback)
	if err != nil {
		t.Fatalf("DownloadFile failed: %v", err)
	}

	if progressReports == 0 {
		t.Errorf("expected at least one progress callback, got 0")
	}
	if lastPercent != 100 {
		t.Errorf("expected final percentage to be 100, got %d", lastPercent)
	}

	content, err := os.ReadFile(destFile)
	if err != nil {
		t.Fatalf("failed to read downloaded file: %v", err)
	}
	if string(content) != string(payload) {
		t.Errorf("content mismatch, got %q, want %q", string(content), string(payload))
	}
}

func TestDownloadFileIntegrity_MatchingHash(t *testing.T) {
	payload := []byte("Secure artifact payload content for SHA-256 integrity check")
	hasher := sha256.New()
	hasher.Write(payload)
	expectedHash := hex.EncodeToString(hasher.Sum(nil))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(payload)))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(payload)
	}))
	defer server.Close()

	tempDir := t.TempDir()
	destFile := filepath.Join(tempDir, "verified_artifact.zip")

	err := downloader.DownloadFile(context.Background(), server.URL, destFile, nil, expectedHash)
	if err != nil {
		t.Fatalf("DownloadFile with matching hash failed unexpectedly: %v", err)
	}

	content, err := os.ReadFile(destFile)
	if err != nil {
		t.Fatalf("failed to read successfully downloaded file: %v", err)
	}
	if !bytes.Equal(content, payload) {
		t.Errorf("downloaded content mismatch")
	}
}

func TestDownloadFileIntegrity_MismatchingHash(t *testing.T) {
	payload := []byte("Compromised or mismatched content payload")
	wrongHash := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855" // hash of empty string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(payload)))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(payload)
	}))
	defer server.Close()

	tempDir := t.TempDir()
	destFile := filepath.Join(tempDir, "compromised_artifact.zip")

	err := downloader.DownloadFile(context.Background(), server.URL, destFile, nil, wrongHash)
	if err == nil {
		t.Fatalf("expected error on hash mismatch, got nil")
	}

	if !strings.Contains(err.Error(), "mismatch") {
		t.Errorf("expected error message to mention 'mismatch', got: %v", err)
	}

	// Verify destFile was NOT created
	if _, err := os.Stat(destFile); !os.IsNotExist(err) {
		t.Errorf("destination file %s should NOT exist after hash mismatch", destFile)
	}

	// Verify no temporary files were left behind in tempDir
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("failed to read temp dir: %v", err)
	}
	if len(entries) > 0 {
		names := make([]string, len(entries))
		for i, e := range entries {
			names[i] = e.Name()
		}
		t.Errorf("temporary files left behind after mismatch: %v", names)
	}
}

func TestExtractZipWithSingleRootDir(t *testing.T) {
	tempDir := t.TempDir()
	zipPath := filepath.Join(tempDir, "nested.zip")
	destDir := filepath.Join(tempDir, "output")

	// Create dummy in-memory zip where all files are under "mariadb-root-10.11/"
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	f1, err := w.Create("mariadb-root-10.11/bin/mysqld.exe")
	if err != nil {
		t.Fatal(err)
	}
	f1.Write([]byte("fake-mysqld-binary"))

	f2, err := w.Create("mariadb-root-10.11/README.txt")
	if err != nil {
		t.Fatal(err)
	}
	f2.Write([]byte("readme info"))

	w.Close()

	if err := os.WriteFile(zipPath, buf.Bytes(), 0644); err != nil {
		t.Fatal(err)
	}

	if err := downloader.ExtractZip(zipPath, destDir); err != nil {
		t.Fatalf("failed to extract zip: %v", err)
	}

	// Should be extracted directly into destDir/bin/mysqld.exe (stripping mariadb-root-10.11)
	expectedBin := filepath.Join(destDir, "bin", "mysqld.exe")
	if _, err := os.Stat(expectedBin); err != nil {
		t.Errorf("expected %s to exist after root stripping, got: %v", expectedBin, err)
	}

	expectedReadme := filepath.Join(destDir, "README.txt")
	if _, err := os.Stat(expectedReadme); err != nil {
		t.Errorf("expected %s to exist after root stripping, got: %v", expectedReadme, err)
	}
}
