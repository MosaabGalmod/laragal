package updater

import (
	"context"
	"encoding/json"
	"fmt"
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
	SHA256          string `json:"sha256,omitempty"`
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
	SHA256     string
}

func extractZipAndSHA(node interface{}) (string, string) {
	m, ok := node.(map[string]interface{})
	if !ok {
		return "", ""
	}

	var urlStr, shaStr string

	if zipNode, hasZip := m["zip"]; hasZip {
		if zipMap, ok := zipNode.(map[string]interface{}); ok {
			if u, ok := zipMap["url"].(string); ok {
				urlStr = u
			}
			if s, ok := zipMap["sha256"].(string); ok {
				shaStr = s
			}
		}
	}

	if urlStr == "" {
		if u, ok := m["url"].(string); ok && strings.HasSuffix(strings.ToLower(u), ".zip") {
			urlStr = u
		}
	}

	if shaStr == "" {
		if s, ok := m["sha256"].(string); ok {
			shaStr = s
		}
	}

	return urlStr, strings.ToLower(strings.TrimSpace(shaStr))
}

func normalizePHPURL(u string) string {
	if u == "" {
		return ""
	}
	if strings.HasPrefix(u, "http://") || strings.HasPrefix(u, "https://") {
		return u
	}
	if strings.HasPrefix(u, "/") {
		return "https://windows.php.net" + u
	}
	return "https://windows.php.net/downloads/releases/" + u
}

// extractNTSx64Build selects the 64-bit Non-Thread Safe (NTS) zip build explicitly,
// prioritizing modern vs17/vs16 toolchains and rejecting TS, debug packs, and x86 builds.
func extractNTSx64Build(vMap map[string]interface{}) (string, string) {
	// 1. Explicit flat keys with compiler prioritization (e.g. "nts-vs17-x64")
	flatKeys := []string{"nts-vs17-x64", "nts-vs16-x64", "nts-vs15-x64", "nts-x64"}
	for _, k := range flatKeys {
		if child, ok := vMap[k]; ok {
			if u, s := extractZipAndSHA(child); u != "" {
				return normalizePHPURL(u), s
			}
		}
	}

	// 2. Nested under "nts": e.g. vMap["nts"]["vs17"]["x64"]["zip"]
	if ntsNode, ok := vMap["nts"].(map[string]interface{}); ok {
		compKeys := []string{"vs17", "vs16", "vs15"}
		for _, comp := range compKeys {
			if compNode, ok := ntsNode[comp].(map[string]interface{}); ok {
				if x64Node, ok := compNode["x64"]; ok {
					if u, s := extractZipAndSHA(x64Node); u != "" {
						return normalizePHPURL(u), s
					}
				}
			}
		}
		if x64Node, ok := ntsNode["x64"]; ok {
			if u, s := extractZipAndSHA(x64Node); u != "" {
				return normalizePHPURL(u), s
			}
		}
		for _, compVal := range ntsNode {
			if compNode, ok := compVal.(map[string]interface{}); ok {
				if x64Node, ok := compNode["x64"]; ok {
					if u, s := extractZipAndSHA(x64Node); u != "" {
						return normalizePHPURL(u), s
					}
				}
			}
		}
	}

	// 3. Nested under compiler key first: e.g. vMap["vs17"]["nts"]["x64"]
	for _, comp := range []string{"vs17", "vs16", "vs15"} {
		if compNode, ok := vMap[comp].(map[string]interface{}); ok {
			if ntsNode, ok := compNode["nts"].(map[string]interface{}); ok {
				if x64Node, ok := ntsNode["x64"]; ok {
					if u, s := extractZipAndSHA(x64Node); u != "" {
						return normalizePHPURL(u), s
					}
				}
			}
		}
	}

	// 4. Any key matching nts and x64 and not debug or ts
	for k, v := range vMap {
		kLower := strings.ToLower(k)
		if strings.Contains(kLower, "nts") && strings.Contains(kLower, "x64") && !strings.Contains(kLower, "debug") && !strings.HasPrefix(kLower, "ts") {
			if u, s := extractZipAndSHA(v); u != "" {
				return normalizePHPURL(u), s
			}
		}
	}

	return "", ""
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

				zipURL, sha := extractNTSx64Build(vMap)

				meta = PHPVersionMeta{
					Version:    verStr,
					IsSecurity: isSec,
					ZipURL:     zipURL,
					SHA256:     sha,
				}
			}
		}

		if latestSemver != "" {
			result[minor] = meta
		}
	}

	return result, nil
}

// FetchPHPReleases fetches and parses windows.php.net releases.json feed
func FetchPHPReleases(ctx context.Context, releasesURL string, timeout time.Duration) (map[string]PHPVersionMeta, error) {
	if releasesURL == "" {
		releasesURL = "https://windows.php.net/downloads/releases/releases.json"
	}
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	req, err := http.NewRequestWithContext(ctx, "GET", releasesURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "LaraGal-Updater/1.0")

	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return ParsePHPReleasesJSON(body)
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
	// 1. Fetch live PHP releases feed
	livePHPReleases, _ := FetchPHPReleases(ctx, opts.PHPReleasesURL, opts.HTTPTimeout)

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
			info.SHA256 = rt.SHA256
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
				if meta.SHA256 != "" {
					info.SHA256 = meta.SHA256
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
