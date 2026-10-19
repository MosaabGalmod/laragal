package downloader_test

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/laragal/laragal/internal/downloader"
)

func TestGetDownloadURL(t *testing.T) {
	url, err := downloader.GetDownloadURL("caddy")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(url, "github.com/caddyserver/caddy") {
		t.Errorf("invalid caddy url: %s", url)
	}

	url, err = downloader.GetDownloadURL("mailpit")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(url, "mailpit") {
		t.Errorf("invalid mailpit url: %s", url)
	}
}

func TestExtractZip(t *testing.T) {
	tempDir := t.TempDir()
	zipPath := filepath.Join(tempDir, "test.zip")
	destDir := filepath.Join(tempDir, "output")

	// Create dummy in-memory zip
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	f, err := w.Create("test.txt")
	if err != nil {
		t.Fatal(err)
	}
	f.Write([]byte("LaraGal Binary Content"))
	w.Close()

	if err := os.WriteFile(zipPath, buf.Bytes(), 0644); err != nil {
		t.Fatal(err)
	}

	if err := downloader.ExtractZip(zipPath, destDir); err != nil {
		t.Fatalf("failed to extract zip: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(destDir, "test.txt"))
	if err != nil {
		t.Fatalf("failed to read extracted file: %v", err)
	}
	if string(content) != "LaraGal Binary Content" {
		t.Errorf("content mismatch: %s", string(content))
	}
}
