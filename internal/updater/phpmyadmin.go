package updater

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/laragal/laragal/internal/runtimes"
)

type PhpMyAdminSource struct {
	BaseURL          string        // e.g. "https://www.phpmyadmin.net/home_page/version.json" or httptest URL
	FilesBaseURL     string        // e.g. "https://files.phpmyadmin.net" or httptest URL
	ActivePHPVersion string        // e.g. "8.4" (if empty, defaults to "8.4")
	HTTPClient       *http.Client  // Optional HTTP client
	HTTPTimeout      time.Duration // Default 10s
	CacheTTL         time.Duration // Default 1h (if < 0, caching disabled)

	mu       sync.Mutex
	cached   []LiveRelease
	cachedAt time.Time
}

func NewPhpMyAdminSource(baseURL ...string) *PhpMyAdminSource {
	u := ""
	if len(baseURL) > 0 {
		u = baseURL[0]
	}
	return &PhpMyAdminSource{
		BaseURL:          u,
		FilesBaseURL:     "https://files.phpmyadmin.net",
		ActivePHPVersion: "8.4",
	}
}

type pmaVersionResponse struct {
	Version  string           `json:"version"`
	Releases []pmaReleaseItem `json:"releases"`
}

type pmaReleaseItem struct {
	Version     string `json:"version"`
	PHPVersions string `json:"php_versions"`
	Date        string `json:"date"`
}

func (s *PhpMyAdminSource) Latest(ctx context.Context, name string) ([]LiveRelease, error) {
	s.mu.Lock()
	cacheTTL := s.CacheTTL
	if cacheTTL == 0 {
		cacheTTL = 1 * time.Hour
	}
	if cacheTTL > 0 && len(s.cached) > 0 && time.Since(s.cachedAt) < cacheTTL {
		res := make([]LiveRelease, len(s.cached))
		copy(res, s.cached)
		s.mu.Unlock()
		return res, nil
	}
	s.mu.Unlock()

	baseURL := s.BaseURL
	if baseURL == "" {
		baseURL = "https://www.phpmyadmin.net/home_page/version.json"
	}

	timeout := s.HTTPTimeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	client := s.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: timeout}
	}

	req, err := http.NewRequestWithContext(ctx, "GET", baseURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "LaraGal-Updater/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("phpmyadmin version API returned HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var pmaResp pmaVersionResponse
	if err := json.Unmarshal(body, &pmaResp); err != nil {
		return nil, fmt.Errorf("failed to parse phpmyadmin version JSON: %w", err)
	}

	ver := strings.TrimSpace(pmaResp.Version)
	if ver == "" && len(pmaResp.Releases) > 0 {
		ver = strings.TrimSpace(pmaResp.Releases[0].Version)
	}
	if ver == "" {
		return nil, fmt.Errorf("phpmyadmin version JSON contained no valid version")
	}

	phpVersionsConstraint := ""
	for _, rel := range pmaResp.Releases {
		if rel.Version == ver {
			phpVersionsConstraint = strings.TrimSpace(rel.PHPVersions)
			break
		}
	}
	if phpVersionsConstraint == "" && len(pmaResp.Releases) > 0 {
		phpVersionsConstraint = strings.TrimSpace(pmaResp.Releases[0].PHPVersions)
	}

	// Build download and checksum URLs
	zipURL, shaURL := s.BuildURLs(ver)

	// Fetch .sha256 checksum file
	shaReq, err := http.NewRequestWithContext(ctx, "GET", shaURL, nil)
	if err != nil {
		return nil, err
	}
	shaReq.Header.Set("User-Agent", "LaraGal-Updater/1.0")

	shaResp, err := client.Do(shaReq)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch phpmyadmin sha256 checksum: %w", err)
	}
	defer shaResp.Body.Close()

	if shaResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("phpmyadmin sha256 returned HTTP %d: %s", shaResp.StatusCode, shaResp.Status)
	}

	shaBytes, err := io.ReadAll(shaResp.Body)
	if err != nil {
		return nil, err
	}

	checksum, err := ParseSha256ChecksumFile(shaBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse phpmyadmin sha256 checksum: %w", err)
	}

	// Check major change against installed or catalog version
	curVer := "5.2.3"
	if catRt, exists := runtimes.Get("phpmyadmin"); exists && catRt.Version != "" {
		curVer = catRt.Version
	}
	majorChange := IsMajorChange(curVer, ver)

	// Evaluate PHP compatibility
	activePHP := s.ActivePHPVersion
	if activePHP == "" {
		activePHP = "8.4"
	}
	compatNote := ""
	if phpVersionsConstraint != "" {
		if !CheckPHPCompatibility(activePHP, phpVersionsConstraint) {
			compatNote = fmt.Sprintf("PHP %s is outside supported range (%s)", activePHP, phpVersionsConstraint)
		}
	}

	relName := name
	if relName == "" {
		relName = "phpmyadmin"
	}

	liveRel := LiveRelease{
		Name:         relName,
		Version:      ver,
		URL:          zipURL,
		Checksum:     checksum,
		ChecksumAlgo: "sha256",
		MajorChange:  majorChange,
		Source:       "phpmyadmin.net",
		CompatNote:   compatNote,
	}

	if cacheTTL > 0 {
		s.mu.Lock()
		s.cached = []LiveRelease{liveRel}
		s.cachedAt = time.Now()
		s.mu.Unlock()
	}

	return []LiveRelease{liveRel}, nil
}

// BuildURLs constructs the official zip download URL and .sha256 checksum URL for a phpMyAdmin version
func (s *PhpMyAdminSource) BuildURLs(ver string) (string, string) {
	filesBase := s.FilesBaseURL
	if filesBase == "" {
		filesBase = "https://files.phpmyadmin.net"
	}
	filesBase = strings.TrimRight(filesBase, "/")

	zipFileName := fmt.Sprintf("phpMyAdmin-%s-all-languages.zip", ver)
	zipURL := fmt.Sprintf("%s/phpMyAdmin/%s/%s", filesBase, ver, zipFileName)
	shaURL := zipURL + ".sha256"
	return zipURL, shaURL
}

// ParseSha256ChecksumFile parses standard "<hex>  <filename>" or "<hex>" format
func ParseSha256ChecksumFile(data []byte) (string, error) {
	fields := strings.Fields(string(data))
	if len(fields) == 0 {
		return "", fmt.Errorf("empty sha256 checksum file")
	}
	hexPart := strings.ToLower(strings.TrimSpace(fields[0]))
	if len(hexPart) != 64 {
		return "", fmt.Errorf("invalid sha256 checksum length (%d): %s", len(hexPart), hexPart)
	}
	for _, c := range hexPart {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return "", fmt.Errorf("invalid sha256 checksum: contains non-hex character '%c'", c)
		}
	}
	return hexPart, nil
}

// CheckPHPCompatibility checks if activePHP satisfies a version constraint like ">=7.2,<8.4".
func CheckPHPCompatibility(activePHP string, constraint string) bool {
	cleanPHP := CleanVersionString(activePHP)
	if cleanPHP == "" {
		return true
	}

	clauses := strings.Split(constraint, ",")
	for _, clause := range clauses {
		c := strings.TrimSpace(clause)
		if c == "" {
			continue
		}

		switch {
		case strings.HasPrefix(c, ">="):
			target := strings.TrimSpace(c[2:])
			if CompareVersionParts(cleanPHP, target) < 0 {
				return false
			}
		case strings.HasPrefix(c, "<="):
			target := strings.TrimSpace(c[2:])
			if CompareVersionParts(cleanPHP, target) > 0 {
				return false
			}
		case strings.HasPrefix(c, ">"):
			target := strings.TrimSpace(c[1:])
			if CompareVersionParts(cleanPHP, target) <= 0 {
				return false
			}
		case strings.HasPrefix(c, "<"):
			target := strings.TrimSpace(c[1:])
			if CompareVersionParts(cleanPHP, target) >= 0 {
				return false
			}
		case strings.HasPrefix(c, "=="):
			target := strings.TrimSpace(c[2:])
			if CompareVersionParts(cleanPHP, target) != 0 {
				return false
			}
		case strings.HasPrefix(c, "="):
			target := strings.TrimSpace(c[1:])
			if CompareVersionParts(cleanPHP, target) != 0 {
				return false
			}
		}
	}

	return true
}

// CompareVersionParts compares two version strings numerically part-by-part.
// Returns -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2.
func CompareVersionParts(v1, v2 string) int {
	p1, b1, _ := parseVersionComponents(v1)
	p2, b2, _ := parseVersionComponents(v2)

	maxLen := len(p1)
	if len(p2) > maxLen {
		maxLen = len(p2)
	}

	for i := 0; i < maxLen; i++ {
		val1 := 0
		if i < len(p1) {
			val1 = p1[i]
		}
		val2 := 0
		if i < len(p2) {
			val2 = p2[i]
		}
		if val1 < val2 {
			return -1
		}
		if val1 > val2 {
			return 1
		}
	}

	if b1 < b2 {
		return -1
	}
	if b1 > b2 {
		return 1
	}

	return 0
}
