package updater

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/laragal/laragal/internal/runtimes"
)

type GitHubSource struct {
	Repo         string
	AssetPattern *regexp.Regexp
	Family       string
	BaseURL      string        // Optional custom URL (e.g. for testing with httptest)
	HTTPClient   *http.Client  // Optional HTTP client
	HTTPTimeout  time.Duration // Default 10s
	CacheTTL     time.Duration // Default 1h (if < 0, caching is disabled)

	mu       sync.Mutex
	cached   []LiveRelease
	cachedAt time.Time
}

func NewGitHubSource(repo string, pattern *regexp.Regexp, family string) *GitHubSource {
	return &GitHubSource{
		Repo:         repo,
		AssetPattern: pattern,
		Family:       family,
	}
}

type gitHubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Digest             string `json:"digest"`
}

type gitHubReleaseResponse struct {
	TagName string        `json:"tag_name"`
	Name    string        `json:"name"`
	Assets  []gitHubAsset `json:"assets"`
	Message string        `json:"message,omitempty"`
}

func (s *GitHubSource) Latest(ctx context.Context, name string) ([]LiveRelease, error) {
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

	url := s.BaseURL
	if url == "" {
		url = fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", s.Repo)
	}

	timeout := s.HTTPTimeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "LaraGal-Updater/1.0")

	client := s.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: timeout}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusTooManyRequests {
		resetHeader := resp.Header.Get("X-RateLimit-Reset")
		if resetHeader != "" {
			if sec, err := strconv.ParseInt(resetHeader, 10, 64); err == nil {
				resetTime := time.Unix(sec, 0).Local()
				return nil, fmt.Errorf("github rate limit, resets %s", resetTime.Format("15:04"))
			}
		}
		return nil, fmt.Errorf("github api rate limit exceeded (HTTP %d): %s", resp.StatusCode, resp.Status)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github api returned HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var ghRel gitHubReleaseResponse
	if err := json.Unmarshal(body, &ghRel); err != nil {
		return nil, fmt.Errorf("failed to parse github release JSON: %w", err)
	}

	verStr := strings.TrimSpace(ghRel.TagName)
	if verStr == "" {
		verStr = strings.TrimSpace(ghRel.Name)
	}
	verStr = strings.TrimPrefix(verStr, "v")
	verStr = strings.TrimPrefix(verStr, "V")
	verStr = strings.TrimSpace(verStr)

	var matchedAsset *gitHubAsset
	for i := range ghRel.Assets {
		asset := &ghRel.Assets[i]
		if s.AssetPattern != nil && s.AssetPattern.MatchString(asset.Name) {
			matchedAsset = asset
			break
		}
	}

	if matchedAsset == nil {
		return nil, fmt.Errorf("no matching asset found for pattern %s in repo %s", s.AssetPattern, s.Repo)
	}

	algo := "sha256"
	checksum := strings.TrimSpace(matchedAsset.Digest)
	if strings.HasPrefix(strings.ToLower(checksum), "sha256:") {
		algo = "sha256"
		checksum = strings.TrimPrefix(checksum, "sha256:")
		checksum = strings.TrimPrefix(checksum, "SHA256:")
	} else if strings.HasPrefix(strings.ToLower(checksum), "sha512:") {
		algo = "sha512"
		checksum = strings.TrimPrefix(checksum, "sha512:")
		checksum = strings.TrimPrefix(checksum, "SHA512:")
	}
	checksum = strings.ToLower(strings.TrimSpace(checksum))

	relName := name
	if relName == "" {
		relName = s.Family
	}
	if relName == "" {
		relName = s.Repo
	}

	majorChange := false
	// Provider change check (e.g. redis switching from tporadowski to redis-windows)
	if (s.Family == "redis" || relName == "redis") && !strings.Contains(s.Repo, "tporadowski") {
		majorChange = true
	}

	if catRt, exists := runtimes.Get(relName); exists {
		if IsMajorChange(catRt.Version, verStr) {
			majorChange = true
		}
	}

	liveRel := LiveRelease{
		Name:         relName,
		Version:      verStr,
		URL:          matchedAsset.BrowserDownloadURL,
		Checksum:     checksum,
		ChecksumAlgo: algo,
		MajorChange:  majorChange,
		Source:       "github.com/" + s.Repo,
	}

	if cacheTTL > 0 {
		s.mu.Lock()
		s.cached = []LiveRelease{liveRel}
		s.cachedAt = time.Now()
		s.mu.Unlock()
	}

	return []LiveRelease{liveRel}, nil
}
