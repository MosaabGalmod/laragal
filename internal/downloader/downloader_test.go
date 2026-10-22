package downloader_test

import (
	"archive/zip"
	"bytes"
	"context"
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
		"php-8.4",
		"php-8.3",
		"php-8.2",
		"php-8.1",
		"mysql",
		"redis",
		"php8.4",
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

	_, err := downloader.GetDownloadURL("unknown-runtime-xyz")
	if err == nil {
		t.Errorf("expected error for unknown runtime, got nil")
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
