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
	"regexp"
	"strings"
	"time"

	"github.com/laragal/laragal/internal/runtimes"
	"github.com/laragal/laragal/internal/winutil"
)

type LiveRelease struct {
	Name         string `json:"name"`
	Version      string `json:"version"`
	URL          string `json:"url"`
	Checksum     string `json:"checksum"`
	ChecksumAlgo string `json:"checksum_algo"` /* "sha256"|"sha512" */
	MajorChange  bool   `json:"major_change"`
	Source       string `json:"source"`
	CompatNote   string `json:"compat_note,omitempty"`
}

type ReleaseSource interface {
	Latest(ctx context.Context, name string) ([]LiveRelease, error)
} // may return several entries (e.g. one per PHP minor)

type Registry map[string]ReleaseSource

// RuntimeFamily returns the family identifier for a runtime name (e.g. "php-8.4" -> "php", "mariadb" -> "mysql").
func RuntimeFamily(name string) string {
	n := strings.ToLower(strings.TrimSpace(name))
	if n == "phpmyadmin" || n == "pma" {
		return "phpmyadmin"
	}
	if strings.HasPrefix(n, "php") {
		return "php"
	}
	if strings.HasPrefix(n, "mariadb") || strings.HasPrefix(n, "mysql") {
		return "mysql"
	}
	return n
}

type PHPSource struct {
	BaseURL     string
	HTTPTimeout time.Duration
}

func NewPHPSource(baseURL ...string) *PHPSource {
	u := ""
	if len(baseURL) > 0 {
		u = baseURL[0]
	}
	return &PHPSource{BaseURL: u}
}

func (s *PHPSource) Latest(ctx context.Context, name string) ([]LiveRelease, error) {
	url := s.BaseURL
	if url == "" {
		url = "https://windows.php.net/downloads/releases/releases.json"
	}
	timeout := s.HTTPTimeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	metas, err := FetchPHPReleases(ctx, url, timeout)
	if err != nil {
		return nil, err
	}

	var results []LiveRelease
	for minor, meta := range metas {
		if meta.Version == "" || meta.ZipURL == "" {
			continue
		}
		rtName := "php-" + minor
		if name != "" && name != "php" {
			norm := runtimes.NormalizePHPVersion(name)
			if minor != norm && rtName != name {
				continue
			}
		}
		results = append(results, LiveRelease{
			Name:         rtName,
			Version:      meta.Version,
			URL:          meta.ZipURL,
			Checksum:     meta.SHA256,
			ChecksumAlgo: "sha256",
			MajorChange:  false,
			Source:       "windows.php.net",
		})
	}
	return results, nil
}

// DefaultRegistry defines the global registry mapping runtime families to ReleaseSource.
var DefaultRegistry = Registry{
	"php":        &PHPSource{},
	"caddy":      NewGitHubSource("caddyserver/caddy", regexp.MustCompile(`^caddy_.*_windows_amd64\.zip$`), "caddy"),
	"mailpit":    NewGitHubSource("axllent/mailpit", regexp.MustCompile(`^mailpit-windows-amd64\.zip$`), "mailpit"),
	"adminer":    NewGitHubSource("vrana/adminer", regexp.MustCompile(`^adminer-.*-mysql\.php$`), "adminer"),
	"redis":      NewGitHubSource("redis-windows/redis-windows", regexp.MustCompile(`^Redis-.*-Windows-x64-msys2\.zip$`), "redis"),
	"mysql":      NewMariaDBSource(),
	"phpmyadmin": NewPhpMyAdminSource(),
}

type RuntimeUpdateInfo struct {
	Name            string `json:"name"`              // "php-8.4", "caddy", etc.
	CurrentVersion  string `json:"current_version"`   // e.g. "8.4.4"
	LatestVersion   string `json:"latest_version"`    // e.g. "8.4.5"
	HasUpdate       bool   `json:"has_update"`
	IsSecurityFix   bool   `json:"is_security_fix"`
	DownloadURL     string `json:"download_url"`
	ReleaseNotesURL string `json:"release_notes_url"`
	SHA256          string `json:"sha256,omitempty"`
	IsNew           bool   `json:"is_new,omitempty"`
	MajorChange     bool   `json:"major_change,omitempty"`
	CompatNote      string `json:"compat_note,omitempty"`
	SourceStatus    string `json:"source_status"`          // "live" | "catalog_fallback"
	SourceError     string `json:"source_error,omitempty"` // short, English, e.g. "github rate limit, resets 23:40" or "timeout"
}

type CheckOptions struct {
	PHPReleasesURL    string
	HTTPTimeout       time.Duration
	BaseBinDir        string
	DataDir           string            // Optional data directory (if empty, derived from BaseBinDir sibling "data")
	LogDir            string            // Optional logs directory (if empty, derived from DataDir/logs)
	InstalledVersions map[string]string // Optional manual/mocked overrides for testing
	Registry          Registry          // Optional custom registry overrides
}

type PHPVersionMeta struct {
	Version    string
	IsSecurity bool
	ZipURL     string
	SHA256     string
}

const defaultPHPReleasesBaseURL = "https://windows.php.net/downloads/releases/"

// ParsePHPReleasesJSON parses windows.php.net releases.json feed according to the real schema.
// Schema: {"8.4": {"version": "8.4.27-1", "nts-vs17-x64": {"zip": {"path": "...", "sha256": "..."}}, ...}}
func ParsePHPReleasesJSON(data []byte, customBaseURL ...string) (map[string]PHPVersionMeta, error) {
	var raw map[string]map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	baseURL := defaultPHPReleasesBaseURL
	if len(customBaseURL) > 0 && customBaseURL[0] != "" {
		baseURL = customBaseURL[0]
		if !strings.HasSuffix(baseURL, "/") {
			baseURL += "/"
		}
	}

	result := make(map[string]PHPVersionMeta)

	for minor, minorData := range raw {
		verVal, ok := minorData["version"]
		if !ok {
			continue
		}
		verStr, ok := verVal.(string)
		if !ok || strings.TrimSpace(verStr) == "" {
			continue
		}

		// Pick "nts-vs17-x64", else "nts-vs16-x64", else "nts-vc15-x64" or any nts-*-x64 build; skip if none
		var buildNode interface{}
		if nts17, exists := minorData["nts-vs17-x64"]; exists {
			buildNode = nts17
		} else if nts16, exists := minorData["nts-vs16-x64"]; exists {
			buildNode = nts16
		} else if nts15, exists := minorData["nts-vc15-x64"]; exists {
			buildNode = nts15
		} else {
			for k, v := range minorData {
				if strings.HasPrefix(k, "nts-") && strings.HasSuffix(k, "-x64") {
					buildNode = v
					break
				}
			}
			if buildNode == nil {
				continue
			}
		}

		buildMap, ok := buildNode.(map[string]interface{})
		if !ok {
			continue
		}

		zipNode, ok := buildMap["zip"]
		if !ok {
			continue
		}
		zipMap, ok := zipNode.(map[string]interface{})
		if !ok {
			continue
		}

		pathVal, ok := zipMap["path"]
		if !ok {
			continue
		}
		pathStr, ok := pathVal.(string)
		if !ok || strings.TrimSpace(pathStr) == "" {
			continue
		}

		shaVal, _ := zipMap["sha256"].(string)

		var zipURL string
		cleanPath := strings.TrimSpace(pathStr)
		if strings.HasPrefix(cleanPath, "http://") || strings.HasPrefix(cleanPath, "https://") {
			zipURL = cleanPath
		} else if strings.HasPrefix(cleanPath, "/") {
			zipURL = "https://windows.php.net" + cleanPath
		} else {
			zipURL = baseURL + cleanPath
		}

		result[minor] = PHPVersionMeta{
			Version:    verStr,
			IsSecurity: false,
			ZipURL:     zipURL,
			SHA256:     strings.ToLower(strings.TrimSpace(shaVal)),
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

	var customBase string
	if strings.Contains(releasesURL, "127.0.0.1") || strings.Contains(releasesURL, "localhost") {
		idx := strings.LastIndex(releasesURL, "/")
		if idx != -1 {
			customBase = releasesURL[:idx] + "/downloads/releases/"
		}
	}

	return ParsePHPReleasesJSON(body, customBase)
}

// GetDownloadURL returns the configured download URL for a runtime from the catalog.
func GetDownloadURL(name string) string {
	if rt, ok := runtimes.Get(name); ok {
		return rt.URL
	}
	return ""
}

var (
	adminerVersionVarRegex = regexp.MustCompile(`(?i)\$VERSION\s*=\s*["']([^"']+)["']`)
	adminerDocRegex        = regexp.MustCompile(`(?i)@version\s+([0-9]+(?:\.[0-9]+)+)`)
	adminerIaRegex         = regexp.MustCompile(`\$ia\s*=\s*["']([0-9]+(?:\.[0-9]+)+)["']`)
	adminerHeaderRegex     = regexp.MustCompile(`(?i)Adminer\s+([0-9]+(?:\.[0-9]+)+)`)
	pmaVersionConstRegex   = regexp.MustCompile(`(?i)(?:const\s+VERSION|\$VERSION|VERSION)\s*=\s*["']([^"']+)["']`)
)

func extractAdminerVersion(content string) string {
	if m := adminerVersionVarRegex.FindStringSubmatch(content); len(m) > 1 {
		return strings.TrimSpace(m[1])
	}
	if m := adminerDocRegex.FindStringSubmatch(content); len(m) > 1 {
		return strings.TrimSpace(m[1])
	}
	if m := adminerIaRegex.FindStringSubmatch(content); len(m) > 1 {
		return strings.TrimSpace(m[1])
	}
	if m := adminerHeaderRegex.FindStringSubmatch(content); len(m) > 1 {
		return strings.TrimSpace(m[1])
	}
	return ""
}

func deriveDataDir(binDir, customDataDir string) string {
	if customDataDir != "" {
		return customDataDir
	}
	if binDir != "" {
		parent := filepath.Dir(binDir)
		if parent != "." && parent != "" {
			return filepath.Join(parent, "data")
		}
	}
	return "data"
}

// DetectInstalledBinaryVersion reads local executable outputs across bin, data, WinGet and PATH
func DetectInstalledBinaryVersion(binDir string, name string, optionalDataDir ...string) string {
	customDataDir := ""
	if len(optionalDataDir) > 0 {
		customDataDir = optionalDataDir[0]
	}
	dataDir := deriveDataDir(binDir, customDataDir)

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
		candidateDirs := []string{filepath.Join(dataDir, "adminer")}
		if dataDir != "data" {
			candidateDirs = append(candidateDirs, filepath.Join("data", "adminer"))
		}
		for _, dir := range candidateDirs {
			for _, f := range []string{"adminer-core.php", "index.php", "adminer.php"} {
				if data, err := os.ReadFile(filepath.Join(dir, f)); err == nil {
					if v := extractAdminerVersion(string(data)); v != "" {
						return v
					}
				}
			}
		}
		return ""
	case name == "phpmyadmin":
		candidateDirs := []string{filepath.Join(dataDir, "phpmyadmin")}
		if dataDir != "data" {
			candidateDirs = append(candidateDirs, filepath.Join("data", "phpmyadmin"))
		}
		for _, pmaDir := range candidateDirs {
			if entries, err := os.ReadDir(pmaDir); err == nil {
				for _, e := range entries {
					if strings.HasPrefix(e.Name(), "RELEASE-DATE-") {
						v := strings.TrimPrefix(e.Name(), "RELEASE-DATE-")
						if v != "" {
							return v
						}
					}
				}
			}
			for _, relFile := range []string{
				filepath.Join("libraries", "classes", "Version.php"),
				filepath.Join("libraries", "Version.php"),
			} {
				vf := filepath.Join(pmaDir, relFile)
				if data, err := os.ReadFile(vf); err == nil {
					if m := pmaVersionConstRegex.FindStringSubmatch(string(data)); len(m) > 1 {
						return strings.TrimSpace(m[1])
					}
				}
			}
			relDate := filepath.Join(pmaDir, "RELEASE-DATE")
			if data, err := os.ReadFile(relDate); err == nil {
				if v := ExtractVersion(string(data)); v != "" {
					return v
				}
			}
		}
		return ""
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
			if matches, gErr := filepath.Glob(filepath.Join(binDir, "redis", "*", "redis-server.exe")); gErr == nil && len(matches) > 0 {
				exe = matches[0]
			}
		}
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

// HighestCatalogPHPMajorMinor returns the highest PHP minor version currently present in the catalog (e.g. "8.5").
func HighestCatalogPHPMajorMinor() string {
	highest := ""
	for _, rt := range runtimes.All() {
		if strings.HasPrefix(rt.Name, "php-") {
			ver := strings.TrimPrefix(rt.Name, "php-")
			if highest == "" || IsVersionNewer(highest, ver) {
				highest = ver
			}
		}
	}
	if highest == "" {
		return "8.5"
	}
	return highest
}

func familyForRuntime(name string) string {
	if strings.HasPrefix(name, "php-") || name == "php" {
		return "php"
	}
	if strings.HasPrefix(name, "mariadb-") || strings.HasPrefix(name, "mysql-") || name == "mysql" {
		return "mysql"
	}
	return name
}

func cleanSourceError(err error) string {
	if err == nil {
		return ""
	}
	msg := err.Error()
	low := strings.ToLower(msg)
	if strings.Contains(low, "deadline exceeded") || strings.Contains(low, "timeout") {
		return "timeout"
	}
	return msg
}

func logSourceFailure(logDir, family string, err error) {
	if logDir == "" || err == nil {
		return
	}
	_ = os.MkdirAll(logDir, 0755)
	logPath := filepath.Join(logDir, "updater.log")

	// Rotate by truncating when >= 1 MB
	const maxLogSize = 1024 * 1024
	if fi, sErr := os.Stat(logPath); sErr == nil {
		if fi.Size() >= maxLogSize {
			_ = os.Truncate(logPath, 0)
		}
	}

	f, oErr := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if oErr != nil {
		return
	}
	defer f.Close()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	line := fmt.Sprintf("%s [%s] source error: %s\n", timestamp, family, err.Error())
	_, _ = f.WriteString(line)
}

// CheckUpdates inspects available releases and compares them with current local installations.
// Uses the registry to enrich catalog runtimes and discover new PHP minors in the feed.
func CheckUpdates(ctx context.Context, opts CheckOptions) []RuntimeUpdateInfo {
	reg := opts.Registry
	if reg == nil {
		reg = DefaultRegistry
	}

	dataDir := deriveDataDir(opts.BaseBinDir, opts.DataDir)
	logDir := opts.LogDir
	if logDir == "" {
		logDir = filepath.Join(dataDir, "logs")
	}

	// 1. Fetch live releases from the registry across all families
	liveMap := make(map[string]LiveRelease)
	var livePHPReleases []LiveRelease
	familyErrors := make(map[string]error)

	if reg != nil {
		for family, src := range reg {
			if src == nil {
				continue
			}
			if family == "php" && opts.PHPReleasesURL != "" {
				src = &PHPSource{
					BaseURL:     opts.PHPReleasesURL,
					HTTPTimeout: opts.HTTPTimeout,
				}
			}
			releases, err := src.Latest(ctx, family)
			if err != nil {
				familyErrors[family] = err
				logSourceFailure(logDir, family, err)
				continue
			}
			for _, rel := range releases {
				liveMap[rel.Name] = rel
				if family == "php" || strings.HasPrefix(rel.Name, "php-") {
					livePHPReleases = append(livePHPReleases, rel)
				}
			}
		}
	} else if opts.PHPReleasesURL != "" {
		phpSrc := &PHPSource{
			BaseURL:     opts.PHPReleasesURL,
			HTTPTimeout: opts.HTTPTimeout,
		}
		releases, err := phpSrc.Latest(ctx, "php")
		if err != nil {
			familyErrors["php"] = err
			logSourceFailure(logDir, "php", err)
		} else {
			for _, rel := range releases {
				liveMap[rel.Name] = rel
				livePHPReleases = append(livePHPReleases, rel)
			}
		}
	}

	// 2. Base catalog runtimes
	catalogItems := runtimes.All()
	runtimesToCheck := make([]string, 0, len(catalogItems))
	for _, rt := range catalogItems {
		runtimesToCheck = append(runtimesToCheck, rt.Name)
	}

	// Include any on-disk php-X.Y folders containing php-cgi.exe or php.exe
	if opts.BaseBinDir != "" {
		if entries, err := os.ReadDir(opts.BaseBinDir); err == nil {
			for _, e := range entries {
				if e.IsDir() && strings.HasPrefix(e.Name(), "php-") {
					cgi := filepath.Join(opts.BaseBinDir, e.Name(), "php-cgi.exe")
					exe := filepath.Join(opts.BaseBinDir, e.Name(), "php.exe")
					cgiExists := false
					if _, sErr := os.Stat(cgi); sErr == nil {
						cgiExists = true
					} else if _, sErr := os.Stat(exe); sErr == nil {
						cgiExists = true
					}
					if cgiExists {
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
	}

	results := make([]RuntimeUpdateInfo, 0, len(runtimesToCheck)+len(livePHPReleases))
	processed := make(map[string]bool)

	for _, name := range runtimesToCheck {
		processed[name] = true
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
			curVer = DetectInstalledBinaryVersion(opts.BaseBinDir, name, opts.DataDir)
		}
		info.CurrentVersion = curVer

		// Determine source status and error
		fam := familyForRuntime(name)
		if _, ok := liveMap[name]; ok {
			info.SourceStatus = "live"
		} else {
			info.SourceStatus = "catalog_fallback"
			if err, hasErr := familyErrors[fam]; hasErr && err != nil {
				info.SourceError = cleanSourceError(err)
			}
		}

		// If live release available in feed, enrich
		if liveRel, ok := liveMap[name]; ok {
			if liveRel.Version != "" {
				info.LatestVersion = liveRel.Version
			}
			if liveRel.URL != "" {
				info.DownloadURL = liveRel.URL
			}
			if liveRel.Checksum != "" {
				info.SHA256 = liveRel.Checksum
			}
			if liveRel.MajorChange {
				info.MajorChange = true
			}
			if liveRel.CompatNote != "" {
				info.CompatNote = liveRel.CompatNote
			}
		}

		baseVer := curVer
		if baseVer == "" {
			if rt, exists := runtimes.Get(name); exists {
				baseVer = rt.Version
			}
		}
		if baseVer != "" && info.LatestVersion != "" {
			if IsMajorChange(baseVer, info.LatestVersion) {
				info.MajorChange = true
			}
		}
		if name == "redis" {
			if liveRel, ok := liveMap[name]; ok && !strings.Contains(liveRel.Source, "tporadowski") {
				info.MajorChange = true
			}
		}

		if curVer != "" && info.LatestVersion != "" {
			info.HasUpdate = IsVersionNewer(curVer, info.LatestVersion)
		} else {
			info.HasUpdate = false
		}

		results = append(results, info)
	}

	// 3. Any PHP minor in the feed but not in the catalog (e.g. 8.6, 9.0) becomes an entry with IsNew=true
	// Rule: a feed-only minor is offered as new ONLY if it is newer than the highest PHP minor in the catalog (today 8.5).
	// Older minors not in the catalog (e.g. 8.0, 7.4) are ignored.
	highestCatalogPHP := HighestCatalogPHPMajorMinor()

	for _, liveRel := range livePHPReleases {
		if !strings.HasPrefix(liveRel.Name, "php-") {
			continue
		}
		minorVer := strings.TrimPrefix(liveRel.Name, "php-")

		// If already in catalog, it was already handled in step 2
		if _, inCatalog := runtimes.Get(liveRel.Name); inCatalog {
			continue
		}

		// Feed-only minor must be strictly newer than the highest catalog minor
		if !IsVersionNewer(highestCatalogPHP, minorVer) {
			continue
		}

		if processed[liveRel.Name] {
			// If already processed (e.g. found on disk), mark IsNew since it is newer than highest in catalog
			for i := range results {
				if results[i].Name == liveRel.Name {
					results[i].IsNew = true
					break
				}
			}
			continue
		}

		curVer := ""
		if opts.InstalledVersions != nil {
			if v, ok := opts.InstalledVersions[liveRel.Name]; ok {
				curVer = v
			}
		}
		if curVer == "" && opts.BaseBinDir != "" {
			curVer = DetectInstalledBinaryVersion(opts.BaseBinDir, liveRel.Name, opts.DataDir)
		}

		hasUpdate := false
		if curVer != "" && liveRel.Version != "" {
			hasUpdate = IsVersionNewer(curVer, liveRel.Version)
		}

		results = append(results, RuntimeUpdateInfo{
			Name:           liveRel.Name,
			CurrentVersion: curVer,
			LatestVersion:  liveRel.Version,
			DownloadURL:    liveRel.URL,
			SHA256:         liveRel.Checksum,
			HasUpdate:      hasUpdate,
			IsNew:          true,
			MajorChange:    liveRel.MajorChange,
			CompatNote:     liveRel.CompatNote,
			SourceStatus:   "live",
		})
		processed[liveRel.Name] = true
	}

	// 4. Any other non-PHP live releases not yet in results (e.g. newer MariaDB LTS series like mariadb-11.4)
	for _, liveRel := range liveMap {
		if processed[liveRel.Name] {
			continue
		}
		if strings.HasPrefix(liveRel.Name, "mariadb-") || strings.HasPrefix(liveRel.Name, "mysql-") {
			results = append(results, RuntimeUpdateInfo{
				Name:           liveRel.Name,
				CurrentVersion: "",
				LatestVersion:  liveRel.Version,
				DownloadURL:    liveRel.URL,
				SHA256:         liveRel.Checksum,
				HasUpdate:      true,
				IsNew:          true,
				MajorChange:    liveRel.MajorChange,
				CompatNote:     liveRel.CompatNote,
				SourceStatus:   "live",
			})
			processed[liveRel.Name] = true
		}
	}

	return results
}
