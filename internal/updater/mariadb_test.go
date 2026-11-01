package updater_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/laragal/laragal/internal/updater"
)

func TestMariaDBSource_PickSameSeriesAndNewerLTS(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			// Root major releases listing
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{
				"major_releases": [
					{
						"release_id": "10.11",
						"release_status": "Stable",
						"release_support_type": "Long Term Support"
					},
					{
						"release_id": "11.4",
						"release_status": "Stable",
						"release_support_type": "Long Term Support"
					},
					{
						"release_id": "11.5",
						"release_status": "Stable",
						"release_support_type": "Short Term Support"
					},
					{
						"release_id": "12.0",
						"release_status": "RC",
						"release_support_type": "Preview"
					},
					{
						"release_id": "12.1",
						"release_status": "Rolling",
						"release_support_type": "Rolling"
					}
				]
			}`))
		case "/10.11/latest/":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{
				"releases": {
					"10.11.19": {
						"release_id": "10.11.19",
						"files": [
							{
								"file_name": "mariadb-10.11.19-winx64.msi",
								"file_download_url": "http://downloads.mariadb.org/rest-api/mariadb/10.11.19/mariadb-10.11.19-winx64.msi",
								"checksum": {"sha256sum": "msi-hash"}
							},
							{
								"file_name": "mariadb-10.11.19-winx64.zip",
								"file_download_url": "http://downloads.mariadb.org/rest-api/mariadb/10.11.19/mariadb-10.11.19-winx64.zip",
								"checksum": {"sha256sum": "398ea30e1234567890abcdef1234567890abcdef1234567890abcdef12345678"}
							}
						]
					}
				}
			}`))
		case "/11.4/latest/":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{
				"releases": {
					"11.4.5": {
						"release_id": "11.4.5",
						"files": [
							{
								"file_name": "mariadb-11.4.5-winx64.zip",
								"file_download_url": "http://downloads.mariadb.org/rest-api/mariadb/11.4.5/mariadb-11.4.5-winx64.zip",
								"checksum": {"sha256sum": "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"}
							}
						]
					}
				}
			}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	src := updater.NewMariaDBSource(ts.URL)
	src.BaseSeries = "10.11"

	releases, err := src.Latest(context.Background(), "mysql")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(releases) != 2 {
		t.Fatalf("expected 2 releases (same-series 10.11 patch + newer LTS 11.4), got %d: %+v", len(releases), releases)
	}

	// 1. Same-series patch
	patch := releases[0]
	if patch.Name != "mysql" {
		t.Errorf("expected patch name 'mysql', got '%s'", patch.Name)
	}
	if patch.Version != "10.11.19" {
		t.Errorf("expected version 10.11.19, got '%s'", patch.Version)
	}
	if patch.MajorChange {
		t.Errorf("expected MajorChange=false for same series patch")
	}
	if !strings.HasPrefix(patch.URL, "https://") {
		t.Errorf("expected http URL to be rewritten to https, got '%s'", patch.URL)
	}
	if patch.Checksum != "398ea30e1234567890abcdef1234567890abcdef1234567890abcdef12345678" {
		t.Errorf("expected sha256 checksum to match, got '%s'", patch.Checksum)
	}

	// 2. Newer LTS series
	lts := releases[1]
	if lts.Name != "mariadb-11.4" {
		t.Errorf("expected LTS name 'mariadb-11.4', got '%s'", lts.Name)
	}
	if lts.Version != "11.4.5" {
		t.Errorf("expected version 11.4.5, got '%s'", lts.Version)
	}
	if !lts.MajorChange {
		t.Errorf("expected MajorChange=true for newer LTS series")
	}
	if !strings.HasPrefix(lts.URL, "https://") {
		t.Errorf("expected http URL to be rewritten to https, got '%s'", lts.URL)
	}

	// Verify RC/Preview (12.0) and Short Term Support (11.5) and Rolling (12.1) are NOT present
	for _, rel := range releases {
		if strings.HasPrefix(rel.Version, "12.") {
			t.Errorf("RC/Preview/Rolling release 12.x must NOT be offered: %+v", rel)
		}
		if strings.HasPrefix(rel.Version, "11.5") {
			t.Errorf("Short Term Support release 11.5 must NOT be offered: %+v", rel)
		}
	}
}
