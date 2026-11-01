package updater_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/laragal/laragal/internal/updater"
)

func TestParsePHPReleasesJSON(t *testing.T) {
	sampleJSON := `{
		"8.4": {
			"8.4.5": {
				"version": "8.4.5",
				"security": "yes",
				"nts": {
					"vs17": {
						"x64": {
							"zip": {
								"url": "https://windows.php.net/downloads/releases/php-8.4.5-nts-Win32-vs17-x64.zip"
							}
						}
					}
				}
			}
		},
		"8.3": {
			"8.3.17": {
				"version": "8.3.17",
				"security": "no"
			}
		}
	}`

	releases, err := updater.ParsePHPReleasesJSON([]byte(sampleJSON))
	if err != nil {
		t.Fatalf("ParsePHPReleasesJSON failed: %v", err)
	}

	if ver, ok := releases["8.4"]; !ok || ver.Version != "8.4.5" {
		t.Errorf("expected 8.4.5, got %v", ver)
	}
	if !releases["8.4"].IsSecurity {
		t.Errorf("expected 8.4.5 to be marked as security release")
	}
	if releases["8.4"].ZipURL != "https://windows.php.net/downloads/releases/php-8.4.5-nts-Win32-vs17-x64.zip" {
		t.Errorf("expected valid zip url, got: %s", releases["8.4"].ZipURL)
	}

	if ver, ok := releases["8.3"]; !ok || ver.Version != "8.3.17" {
		t.Errorf("expected 8.3.17, got %v", ver)
	}
	if releases["8.3"].IsSecurity {
		t.Errorf("expected 8.3.17 to not be marked as security release")
	}
}

func TestParsePHPReleasesJSON_SelectsNTSx64BuildExplicitly(t *testing.T) {
	complexJSON := `{
		"8.4": {
			"8.4.6": {
				"version": "8.4.6",
				"security": "yes",
				"ts-vs17-x64": {
					"zip": {
						"url": "https://windows.php.net/downloads/releases/php-8.4.6-Win32-vs17-x64.zip",
						"sha256": "1111111111111111111111111111111111111111111111111111111111111111"
					}
				},
				"nts-vs17-x86": {
					"zip": {
						"url": "https://windows.php.net/downloads/releases/php-8.4.6-nts-Win32-vs17-x86.zip",
						"sha256": "2222222222222222222222222222222222222222222222222222222222222222"
					}
				},
				"debug_pack": {
					"zip": {
						"url": "https://windows.php.net/downloads/releases/php-debug-pack-8.4.6-nts-Win32-vs17-x64.zip",
						"sha256": "3333333333333333333333333333333333333333333333333333333333333333"
					}
				},
				"nts-vs17-x64": {
					"zip": {
						"url": "https://windows.php.net/downloads/releases/php-8.4.6-nts-Win32-vs17-x64.zip",
						"sha256": "4444444444444444444444444444444444444444444444444444444444444444"
					}
				}
			}
		}
	}`

	releases, err := updater.ParsePHPReleasesJSON([]byte(complexJSON))
	if err != nil {
		t.Fatalf("ParsePHPReleasesJSON failed: %v", err)
	}

	meta, ok := releases["8.4"]
	if !ok {
		t.Fatalf("expected 8.4 in releases")
	}

	expectedURL := "https://windows.php.net/downloads/releases/php-8.4.6-nts-Win32-vs17-x64.zip"
	if meta.ZipURL != expectedURL {
		t.Errorf("expected NTS x64 zip URL %q, got %q", expectedURL, meta.ZipURL)
	}

	expectedSHA := "4444444444444444444444444444444444444444444444444444444444444444"
	if meta.SHA256 != expectedSHA {
		t.Errorf("expected SHA-256 %q, got %q", expectedSHA, meta.SHA256)
	}
}

func TestCheckUpdates_WithMockServer(t *testing.T) {
	mockJSON := `{
		"8.4": {
			"8.4.5": {
				"version": "8.4.5",
				"security": "yes",
				"announcement": "Security fix for CVE-2025-XXXX"
			}
		}
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(mockJSON))
	}))
	defer server.Close()

	tempBin := t.TempDir()
	// Create simulated php-8.4 folder
	phpDir := filepath.Join(tempBin, "php-8.4")
	_ = os.MkdirAll(phpDir, 0755)

	opts := updater.CheckOptions{
		PHPReleasesURL: server.URL,
		HTTPTimeout:    2 * time.Second,
		BaseBinDir:     tempBin,
		InstalledVersions: map[string]string{
			"php-8.4": "8.4.4", // Current installed is older
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
		t.Errorf("expected HasUpdate to be true for php-8.4 (8.4.4 vs 8.4.5)")
	}
	if !phpUpdate.IsSecurityFix {
		t.Errorf("expected IsSecurityFix to be true")
	}
	if phpUpdate.LatestVersion != "8.4.5" {
		t.Errorf("expected latest version 8.4.5, got %s", phpUpdate.LatestVersion)
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
