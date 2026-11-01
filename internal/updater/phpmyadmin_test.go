package updater_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/laragal/laragal/internal/updater"
)

func TestPhpMyAdminSource_ChecksumParsedAndCompatNote(t *testing.T) {
	fakeHash := "7e8f5223e7f9185a53a921473de079d2319f6a15db47d341bb82e66d981882d4"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/home_page/version.json":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{
				"version": "5.2.3",
				"releases": [
					{
						"version": "5.2.3",
						"php_versions": ">=7.2,<8.4",
						"date": "2025-01-01"
					}
				]
			}`))
		case strings.HasSuffix(r.URL.Path, ".zip.sha256"):
			// Standard sha256 output format: "<hex>  <filename>"
			w.Write([]byte(fakeHash + "  phpMyAdmin-5.2.3-all-languages.zip\n"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	// 1. With ActivePHPVersion = "8.4" -> Should produce CompatNote because "<8.4" rejects 8.4
	src := updater.NewPhpMyAdminSource(ts.URL + "/home_page/version.json")
	src.ActivePHPVersion = "8.4"

	releases, err := src.Latest(context.Background(), "phpmyadmin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(releases) != 1 {
		t.Fatalf("expected 1 release, got %d", len(releases))
	}

	rel := releases[0]
	if rel.Name != "phpmyadmin" {
		t.Errorf("expected name 'phpmyadmin', got '%s'", rel.Name)
	}
	if rel.Version != "5.2.3" {
		t.Errorf("expected version '5.2.3', got '%s'", rel.Version)
	}
	if rel.Checksum != fakeHash {
		t.Errorf("expected checksum '%s', got '%s'", fakeHash, rel.Checksum)
	}
	if rel.MajorChange {
		t.Errorf("expected MajorChange=false for 5.2.3 when catalog is 5.2.3")
	}
	if rel.CompatNote == "" {
		t.Errorf("expected CompatNote to be non-empty for PHP 8.4 vs '<8.4'")
	}
	if !strings.Contains(rel.CompatNote, "8.4") || !strings.Contains(rel.CompatNote, "<8.4") {
		t.Errorf("expected CompatNote to mention 8.4 and <8.4, got: %s", rel.CompatNote)
	}

	// 2. With ActivePHPVersion = "8.3" -> Should have empty CompatNote because 8.3 is in ">=7.2,<8.4"
	src2 := updater.NewPhpMyAdminSource(ts.URL + "/home_page/version.json")
	src2.ActivePHPVersion = "8.3"
	src2.CacheTTL = -1 // disable cache

	releases2, err := src2.Latest(context.Background(), "phpmyadmin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if releases2[0].CompatNote != "" {
		t.Errorf("expected empty CompatNote for PHP 8.3, got '%s'", releases2[0].CompatNote)
	}
}

func TestPhpMyAdminSource_InvalidChecksumFails(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/home_page/version.json":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{
				"version": "5.2.3",
				"releases": [{"version": "5.2.3", "php_versions": ">=7.2"}]
			}`))
		case strings.HasSuffix(r.URL.Path, ".zip.sha256"):
			// Return corrupt/non-hex checksum
			w.Write([]byte("not-a-valid-sha256-hash\n"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	src := updater.NewPhpMyAdminSource(ts.URL + "/home_page/version.json")
	_, err := src.Latest(context.Background(), "phpmyadmin")
	if err == nil {
		t.Fatalf("expected error for invalid sha256 checksum, got nil")
	}
}

func TestPhpMyAdminSource_MajorChangeDetection(t *testing.T) {
	fakeHash := "1111111111111111111111111111111111111111111111111111111111111111"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/home_page/version.json":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{
				"version": "6.0.0",
				"releases": [{"version": "6.0.0", "php_versions": ">=8.2"}]
			}`))
		case strings.HasSuffix(r.URL.Path, ".zip.sha256"):
			w.Write([]byte(fakeHash + "  phpMyAdmin-6.0.0-all-languages.zip\n"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	src := updater.NewPhpMyAdminSource(ts.URL + "/home_page/version.json")
	releases, err := src.Latest(context.Background(), "phpmyadmin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !releases[0].MajorChange {
		t.Errorf("expected MajorChange=true for phpmyadmin 6.0.0 (catalog is 5.2.3)")
	}
}
