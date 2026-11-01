package updater

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/laragal/laragal/internal/runtimes"
	"github.com/laragal/laragal/internal/winutil"
)

type RuntimeUpdateInfo struct {
	Name            string `json:"name"`              // "php-8.4", "caddy", etc.
	CurrentVersion  string `json:"current_version"`   // e.g. "8.4.4"
	LatestVersion   string `json:"latest_version"`    // e.g. "8.4.5"
	HasUpdate       bool   `json:"has_update"`
	IsSecurityFix   bool   `json:"is_security_fix"`
	DownloadURL     string `json:"download_url"`
	ReleaseNotesURL string `json:"release_notes_url"`
}

type CheckOptions struct {
	PHPReleasesURL    string
	HTTPTimeout       time.Duration
	BaseBinDir        string
	InstalledVersions map[string]string // Optional manual/mocked overrides for testing
}

type PHPVersionMeta struct {
	Version    string
	IsSecurity bool
	ZipURL     string
}

func findZipURLRecursive(node interface{}) string {
	switch val := node.(type) {
	case map[string]interface{}:
		for k, v := range val {
			if k == "url" {
				if s, ok := v.(string); ok && strings.HasSuffix(strings.ToLower(s), ".zip") {
					return s
				}
			}
			if found := findZipURLRecursive(v); found != "" {
				return found
			}
		}
	case []interface{}:
		for _, item := range val {
			if found := findZipURLRecursive(item); found != "" {
				return found
			}
		}
	}
	return ""
}

// ParsePHPReleasesJSON parses windows.php.net releases.json feed
func ParsePHPReleasesJSON(data []byte) (map[string]PHPVersionMeta, error) {
	var raw map[string]map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	result := make(map[string]PHPVersionMeta)

	for minor, versions := range raw {
		var latestSemver string
		var meta PHPVersionMeta

		for verKey, vData := range versions {
			vMap, ok := vData.(map[string]interface{})
			if !ok {
				continue
			}

			verStr := verKey
			if v, ok := vMap["version"].(string); ok && v != "" {
				verStr = v
			}

			if latestSemver == "" || IsVersionNewer(latestSemver, verStr) {
				latestSemver = verStr
				isSec := false
				if secVal, ok := vMap["security"]; ok {
					switch s := secVal.(type) {
					case bool:
						isSec = s
					case string:
						isSec = strings.ToLower(s) == "yes" || strings.ToLower(s) == "true"
					}
				}

				zipURL := findZipURLRecursive(vMap)

				meta = PHPVersionMeta{
					Version:    verStr,
					IsSecurity: isSec,
					ZipURL:     zipURL,
				}
			}
		}

		if latestSemver != "" {
			result[minor] = meta
		}
	}

	return result, nil
}

// GetDownloadURL returns the configured download URL for a runtime from the catalog.
func GetDownloadURL(name string) string {
	if rt, ok := runtimes.Get(name); ok {
		return rt.URL
	}
	return ""
}

// DetectInstalledBinaryVersion reads local executable outputs across bin, data, WinGet and PATH
func DetectInstalledBinaryVersion(binDir string, name string) string {
	switch {
	case name == "caddy":
		exe := filepath.Join(binDir, "caddy", "caddy.exe")
		if _, err := os.Stat(exe); err != nil {
			if p, err := exec.LookPath("caddy.exe"); err == nil {
				exe = p
			}
		}
		if _, err := os.Stat(exe); err == nil {
			out, _ := winutil.Command(exe, "version").CombinedOutput()
			return ExtractVersion(string(out))
		}
	case name == "mailpit":
		exe := filepath.Join(binDir, "mailpit", "mailpit.exe")
		if _, err := os.Stat(exe); err != nil {
			if p, err := exec.LookPath("mailpit.exe"); err == nil {
				exe = p
			}
		}
		if _, err := os.Stat(exe); err == nil {
			out, _ := winutil.Command(exe, "version").CombinedOutput()
			return ExtractVersion(string(out))
		}
	case name == "adminer":
		idx := filepath.Join("data", "adminer", "index.php")
		if _, err := os.Stat(idx); err == nil {
			return "4.8.1"
		}
	case name == "phpmyadmin":
		idx := filepath.Join("data", "phpmyadmin", "index.php")
		if _, err := os.Stat(idx); err == nil {
			return "5.2.3"
		}
	case strings.HasPrefix(name, "php-"):
		ver := strings.TrimPrefix(name, "php-")
		exe := filepath.Join(binDir, name, "php.exe")
		if _, err := os.Stat(exe); err != nil {
			exe = filepath.Join(binDir, name, "php-cgi.exe")
		}
		// If not in binDir, check WinGet packages
		if _, err := os.Stat(exe); err != nil {
			localApp := os.Getenv("LOCALAPPDATA")
			if localApp != "" {
				wgDir := filepath.Join(localApp, "Microsoft", "WinGet", "Packages")
				if entries, rErr := os.ReadDir(wgDir); rErr == nil {
					targetPkg := "PHP.PHP." + ver
					for _, entry := range entries {
						if entry.IsDir() && strings.HasPrefix(entry.Name(), targetPkg) {
							candidate := filepath.Join(wgDir, entry.Name(), "php.exe")
							if _, cErr := os.Stat(candidate); cErr == nil {
								exe = candidate
								break
							}
						}
					}
				}
			}
		}
		if _, err := os.Stat(exe); err == nil {
			out, _ := winutil.Command(exe, "-v").CombinedOutput()
			return ExtractVersion(string(out))
		}
	case name == "mysql":
		exe := filepath.Join(binDir, "mysql", "bin", "mysqld.exe")
		if _, err := os.Stat(exe); err != nil {
			exe = filepath.Join(binDir, "mysql", "mysqld.exe")
		}
		if _, err := os.Stat(exe); err != nil {
			if p, err := exec.LookPath("mysqld.exe"); err == nil {
				exe = p
			}
		}
		if _, err := os.Stat(exe); err == nil {
			out, _ := winutil.Command(exe, "--version").CombinedOutput()
			return ExtractVersion(string(out))
		}
	case name == "redis":
		exe := filepath.Join(binDir, "redis", "redis-server.exe")
		if _, err := os.Stat(exe); err != nil {
			if p, err := exec.LookPath("redis-server.exe"); err == nil {
				exe = p
			}
		}
		if _, err := os.Stat(exe); err == nil {
			out, _ := winutil.Command(exe, "--version").CombinedOutput()
			return ExtractVersion(string(out))
		}
	}
	return ""
}

// CheckUpdates inspects available releases and compares them with current local installations
func CheckUpdates(ctx context.Context, opts CheckOptions) []RuntimeUpdateInfo {
	phpURL := opts.PHPReleasesURL
	if phpURL == "" {
		phpURL = "https://windows.php.net/downloads/releases/releases.json"
	}
	timeout := opts.HTTPTimeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	// 1. Fetch live PHP releases feed
	var livePHPReleases map[string]PHPVersionMeta
	req, err := http.NewRequestWithContext(ctx, "GET", phpURL, nil)
	if err == nil {
		req.Header.Set("User-Agent", "LaraGal-Updater/1.0")
		client := &http.Client{Timeout: timeout}
		resp, fetchErr := client.Do(req)
		if fetchErr == nil && resp.StatusCode == http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			livePHPReleases, _ = ParsePHPReleasesJSON(body)
		}
	}

	runtimesToCheck := []string{"caddy", "php-8.5", "php-8.4", "php-8.3", "php-8.2", "php-8.1", "mysql", "redis", "mailpit", "phpmyadmin", "adminer"}
	if opts.BaseBinDir != "" {
		if entries, err := os.ReadDir(opts.BaseBinDir); err == nil {
			for _, e := range entries {
				if e.IsDir() && strings.HasPrefix(e.Name(), "php-") {
					found := false
					for _, r := range runtimesToCheck {
						if r == e.Name() {
							found = true
							break
						}
					}
					if !found {
						runtimesToCheck = append(runtimesToCheck, e.Name())
					}
				}
			}
		}
	}
	results := make([]RuntimeUpdateInfo, 0, len(runtimesToCheck))

	for _, name := range runtimesToCheck {
		info := RuntimeUpdateInfo{Name: name}
		if rt, exists := runtimes.Get(name); exists {
			info.LatestVersion = rt.Version
			info.DownloadURL = rt.URL
		}

		// Detect installed version
		curVer := ""
		if opts.InstalledVersions != nil {
			if v, ok := opts.InstalledVersions[name]; ok {
				curVer = v
			}
		}
		if curVer == "" && opts.BaseBinDir != "" {
			curVer = DetectInstalledBinaryVersion(opts.BaseBinDir, name)
		}
		info.CurrentVersion = curVer

		// If live PHP releases available, enrich PHP runtimes
		if strings.HasPrefix(name, "php-") && livePHPReleases != nil {
			minorKey := strings.TrimPrefix(name, "php-")
			if meta, ok := livePHPReleases[minorKey]; ok && meta.Version != "" {
				info.LatestVersion = meta.Version
				info.IsSecurityFix = meta.IsSecurity
				if meta.ZipURL != "" {
					info.DownloadURL = meta.ZipURL
				}
			}
		}

		// Check if update is available
		if curVer != "" && info.LatestVersion != "" {
			info.HasUpdate = IsVersionNewer(curVer, info.LatestVersion)
		} else {
			info.HasUpdate = false
		}

		results = append(results, info)
	}

	return results
}
