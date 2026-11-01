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

func TestCheckUpdates_MariaDBInstalledSeriesFiltering(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/":
			w.Write([]byte(`{
				"major_releases": [
					{"release_id": "10.11", "release_status": "Stable", "release_support_type": "Long Term Support"},
					{"release_id": "11.4", "release_status": "Stable", "release_support_type": "Long Term Support"},
					{"release_id": "11.8", "release_status": "Stable", "release_support_type": "Long Term Support"},
					{"release_id": "12.3", "release_status": "Stable", "release_support_type": "Long Term Support"}
				]
			}`))
		case "/10.11/latest/":
			w.Write([]byte(`{"releases": {"10.11.19": {"release_id": "10.11.19", "files": [{"file_name": "mariadb-10.11.19-winx64.zip", "file_download_url": "https://downloads.mariadb.org/10.11.19.zip", "checksum": {"sha256sum": "aaa"}}]}}}`))
		case "/11.4/latest/":
			w.Write([]byte(`{"releases": {"11.4.5": {"release_id": "11.4.5", "files": [{"file_name": "mariadb-11.4.5-winx64.zip", "file_download_url": "https://downloads.mariadb.org/11.4.5.zip", "checksum": {"sha256sum": "bbb"}}]}}}`))
		case "/11.8/latest/":
			w.Write([]byte(`{"releases": {"11.8.10": {"release_id": "11.8.10", "files": [{"file_name": "mariadb-11.8.10-winx64.zip", "file_download_url": "https://downloads.mariadb.org/11.8.10.zip", "checksum": {"sha256sum": "ccc"}}]}}}`))
		case "/12.3/latest/":
			w.Write([]byte(`{"releases": {"12.3.1": {"release_id": "12.3.1", "files": [{"file_name": "mariadb-12.3.1-winx64.zip", "file_download_url": "https://downloads.mariadb.org/12.3.1.zip", "checksum": {"sha256sum": "ddd"}}]}}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	// 1. Installed 11.8.9:
	// - Patch release 11.8.10 is newer -> has_update=true for mysql
	// - Only mariadb-12.3 is offered (11.4 and 11.8 disappear)
	mariaSrc1 := updater.NewMariaDBSource(ts.URL)
	mariaSrc1.CacheTTL = -1
	opts1 := updater.CheckOptions{
		Registry: updater.Registry{
			"mysql": mariaSrc1,
		},
		InstalledVersions: map[string]string{
			"mysql": "11.8.9",
		},
	}
	updates1 := updater.CheckUpdates(context.Background(), opts1)

	var mysql1 *updater.RuntimeUpdateInfo
	var offeredSeries1 []string
	for _, u := range updates1 {
		if u.Name == "mysql" {
			mysqlCopy := u
			mysql1 = &mysqlCopy
		} else if strings.HasPrefix(u.Name, "mariadb-") {
			offeredSeries1 = append(offeredSeries1, u.Name)
		}
	}

	if mysql1 == nil {
		t.Fatalf("expected mysql update info in results")
	}
	if mysql1.CurrentVersion != "11.8.9" {
		t.Errorf("expected CurrentVersion=11.8.9, got %s", mysql1.CurrentVersion)
	}
	if mysql1.LatestVersion != "11.8.10" {
		t.Errorf("expected LatestVersion=11.8.10, got %s", mysql1.LatestVersion)
	}
	if !mysql1.HasUpdate {
		t.Errorf("expected same-series patch detection to flag HasUpdate=true for 11.8.9 -> 11.8.10")
	}
	if len(offeredSeries1) != 1 || offeredSeries1[0] != "mariadb-12.3" {
		t.Errorf("installed 11.8.9: expected only [mariadb-12.3] offered, got %+v", offeredSeries1)
	}

	// 2. Installed 10.11.19:
	// - Patch release 10.11.19 is up-to-date -> has_update=false for mysql
	// - All newer LTS series (11.4, 11.8, 12.3) are offered
	mariaSrc2 := updater.NewMariaDBSource(ts.URL)
	mariaSrc2.CacheTTL = -1
	opts2 := updater.CheckOptions{
		Registry: updater.Registry{
			"mysql": mariaSrc2,
		},
		InstalledVersions: map[string]string{
			"mysql": "10.11.19",
		},
	}
	updates2 := updater.CheckUpdates(context.Background(), opts2)

	var mysql2 *updater.RuntimeUpdateInfo
	var offeredSeries2 []string
	for _, u := range updates2 {
		if u.Name == "mysql" {
			mysqlCopy := u
			mysql2 = &mysqlCopy
		} else if strings.HasPrefix(u.Name, "mariadb-") {
			offeredSeries2 = append(offeredSeries2, u.Name)
		}
	}

	if mysql2 == nil {
		t.Fatalf("expected mysql update info in results")
	}
	if mysql2.CurrentVersion != "10.11.19" {
		t.Errorf("expected CurrentVersion=10.11.19, got %s", mysql2.CurrentVersion)
	}
	if mysql2.LatestVersion != "10.11.19" {
		t.Errorf("expected LatestVersion=10.11.19, got %s", mysql2.LatestVersion)
	}
	if mysql2.HasUpdate {
		t.Errorf("expected HasUpdate=false when installed 10.11.19 matches latest patch")
	}
	expectedSeries2 := []string{"mariadb-11.4", "mariadb-11.8", "mariadb-12.3"}
	if len(offeredSeries2) != 3 {
		t.Fatalf("installed 10.11.19: expected 3 newer series offered, got %+v", offeredSeries2)
	}
	for i, exp := range expectedSeries2 {
		if offeredSeries2[i] != exp {
			t.Errorf("expected offered series [%d] == %s, got %s", i, exp, offeredSeries2[i])
		}
	}
}
