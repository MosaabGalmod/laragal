package updater

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
)

var mariadbFileRegex = regexp.MustCompile(`^mariadb-[\d.]+-winx64\.zip$`)

type MariaDBSource struct {
	BaseURL          string        // e.g. "https://downloads.mariadb.org/rest-api/mariadb/" or httptest URL
	BaseSeries       string        // Default: "10.11"
	InstalledVersion string        // Detected or specified installed version (e.g. "11.8.9")
	HTTPClient       *http.Client  // Optional HTTP client
	HTTPTimeout      time.Duration // Default 10s
	CacheTTL         time.Duration // Default 1h (if < 0, caching disabled)

	mu           sync.Mutex
	cached       []LiveRelease
	cachedSeries string
	cachedAt     time.Time
}

func NewMariaDBSource(baseURL ...string) *MariaDBSource {
	u := ""
	if len(baseURL) > 0 {
		u = baseURL[0]
	}
	return &MariaDBSource{BaseURL: u}
}

type mariaDBMajorReleasesResponse struct {
	MajorReleases []mariaDBMajorRelease `json:"major_releases"`
}

type mariaDBMajorRelease struct {
	ReleaseID          string `json:"release_id"`
	ReleaseName        string `json:"release_name"`
	ReleaseStatus      string `json:"release_status"`
	ReleaseSupportType string `json:"release_support_type"`
	ReleaseEolDate     string `json:"release_eol_date"`
}

type mariaDBSpecificReleaseResponse struct {
	Releases map[string]mariaDBReleaseDetail `json:"releases"`
}

type mariaDBReleaseDetail struct {
	ReleaseID string             `json:"release_id"`
	Files     []mariaDBFileEntry `json:"files"`
}

type mariaDBFileEntry struct {
	FileName        string          `json:"file_name"`
	FileDownloadURL string          `json:"file_download_url"`
	Checksum        mariaDBChecksum `json:"checksum"`
}

type mariaDBChecksum struct {
	SHA256Sum string `json:"sha256sum"`
}

func (s *MariaDBSource) Latest(ctx context.Context, name string) ([]LiveRelease, error) {
	curSeries := s.BaseSeries
	if s.InstalledVersion != "" {
		curSeries = ExtractSeries(s.InstalledVersion)
	}
	if curSeries == "" {
		curSeries = "10.11"
	}

	s.mu.Lock()
	cacheTTL := s.CacheTTL
	if cacheTTL == 0 {
		cacheTTL = 1 * time.Hour
	}
	if cacheTTL > 0 && len(s.cached) > 0 && s.cachedSeries == curSeries && time.Since(s.cachedAt) < cacheTTL {
		res := make([]LiveRelease, len(s.cached))
		copy(res, s.cached)
		s.mu.Unlock()
		return res, nil
	}
	s.mu.Unlock()

	baseURL := s.BaseURL
	if baseURL == "" {
		baseURL = "https://downloads.mariadb.org/rest-api/mariadb/"
	}
	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}

	timeout := s.HTTPTimeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	client := s.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: timeout}
	}

	// Discover major releases & EOL dates from root API
	majorReleases, majorErr := s.fetchMajorReleases(ctx, client, baseURL)
	eolMap := make(map[string]string)
	if majorErr == nil {
		for _, mr := range majorReleases {
			eolMap[mr.ReleaseID] = mr.ReleaseEolDate
		}
	}

	// If a specific series is requested (e.g. "mariadb-11.4" or "mysql-11.4")
	if strings.HasPrefix(name, "mariadb-") || strings.HasPrefix(name, "mysql-") {
		reqSeries := strings.TrimPrefix(name, "mariadb-")
		reqSeries = strings.TrimPrefix(reqSeries, "mysql-")
		isMajor := IsVersionNewer(curSeries, reqSeries)
		rel, err := s.fetchSeriesRelease(ctx, client, baseURL, reqSeries, name, isMajor)
		if err != nil {
			return nil, err
		}
		if eol, ok := eolMap[reqSeries]; ok {
			rel.EOLDate = eol
		}
		return []LiveRelease{rel}, nil
	}

	var results []LiveRelease

	// 1. Fetch current/installed series patch release
	patchRel, err := s.fetchSeriesRelease(ctx, client, baseURL, curSeries, "mysql", false)
	if err == nil {
		if eol, ok := eolMap[curSeries]; ok {
			patchRel.EOLDate = eol
		}
		results = append(results, patchRel)
	}

	// 2. Discover newer Stable LTS series
	if majorErr == nil {
		for _, mr := range majorReleases {
			// Policy: Only Stable LTS. Never offer Preview/RC/Rolling/Short Term Support.
			statusLower := strings.ToLower(strings.TrimSpace(mr.ReleaseStatus))
			supportLower := strings.ToLower(strings.TrimSpace(mr.ReleaseSupportType))

			if statusLower != "stable" {
				continue
			}
			if strings.Contains(statusLower, "rc") || strings.Contains(statusLower, "preview") || strings.Contains(statusLower, "rolling") {
				continue
			}
			if !strings.Contains(supportLower, "long term") && !strings.Contains(supportLower, "lts") {
				continue
			}
			if strings.Contains(supportLower, "short term") {
				continue
			}

			// Must be strictly newer series than current installed series
			if IsVersionNewer(curSeries, mr.ReleaseID) {
				newerRelName := "mariadb-" + mr.ReleaseID
				newerRel, err := s.fetchSeriesRelease(ctx, client, baseURL, mr.ReleaseID, newerRelName, true)
				if err == nil {
					newerRel.EOLDate = mr.ReleaseEolDate
					results = append(results, newerRel)
				}
			}
		}
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no valid MariaDB releases found")
	}

	if cacheTTL > 0 {
		s.mu.Lock()
		s.cached = make([]LiveRelease, len(results))
		copy(s.cached, results)
		s.cachedSeries = curSeries
		s.cachedAt = time.Now()
		s.mu.Unlock()
	}

	return results, nil
}

func (s *MariaDBSource) fetchMajorReleases(ctx context.Context, client *http.Client, baseURL string) ([]mariaDBMajorRelease, error) {
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
		return nil, fmt.Errorf("mariadb major releases api returned HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var respObj mariaDBMajorReleasesResponse
	if err := json.Unmarshal(body, &respObj); err != nil {
		return nil, fmt.Errorf("failed to parse mariadb major releases JSON: %w", err)
	}

	return respObj.MajorReleases, nil
}

func (s *MariaDBSource) fetchSeriesRelease(ctx context.Context, client *http.Client, baseURL, series, relName string, majorChange bool) (LiveRelease, error) {
	seriesURL := fmt.Sprintf("%s%s/latest/", baseURL, series)
	req, err := http.NewRequestWithContext(ctx, "GET", seriesURL, nil)
	if err != nil {
		return LiveRelease{}, err
	}
	req.Header.Set("User-Agent", "LaraGal-Updater/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return LiveRelease{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return LiveRelease{}, fmt.Errorf("mariadb series %s api returned HTTP %d: %s", series, resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return LiveRelease{}, err
	}

	var respObj mariaDBSpecificReleaseResponse
	if err := json.Unmarshal(body, &respObj); err != nil {
		return LiveRelease{}, fmt.Errorf("failed to parse mariadb series release JSON: %w", err)
	}

	for verKey, relDetail := range respObj.Releases {
		ver := relDetail.ReleaseID
		if ver == "" {
			ver = verKey
		}
		for _, file := range relDetail.Files {
			if mariadbFileRegex.MatchString(file.FileName) {
				dlURL := file.FileDownloadURL
				// Ensure HTTPS: rewrite http:// to https:// (keep plain http only for local test loopback)
				if strings.HasPrefix(strings.ToLower(dlURL), "http://") {
					if !strings.Contains(dlURL, "127.0.0.1") && !strings.Contains(dlURL, "localhost") {
						dlURL = "https://" + dlURL[7:]
					}
				}

				sha := strings.ToLower(strings.TrimSpace(file.Checksum.SHA256Sum))
				if sha == "" {
					continue
				}

				return LiveRelease{
					Name:         relName,
					Version:      ver,
					URL:          dlURL,
					Checksum:     sha,
					ChecksumAlgo: "sha256",
					MajorChange:  majorChange,
					Source:       "downloads.mariadb.org",
				}, nil
			}
		}
	}

	return LiveRelease{}, fmt.Errorf("no matching winx64 zip found for mariadb series %s", series)
}

// RecommendedMariaDBSeries finds the release among the given MariaDB series that has the latest EOLDate.
// If multiple series have the same EOLDate or invalid dates, ties are broken by version.
func RecommendedMariaDBSeries(releases []LiveRelease) *LiveRelease {
	if len(releases) == 0 {
		return nil
	}
	var best *LiveRelease
	var bestTime time.Time
	var bestHasDate bool

	for i := range releases {
		rel := &releases[i]
		t, err := time.Parse("2006-01-02", strings.TrimSpace(rel.EOLDate))
		hasDate := (err == nil && !t.IsZero())

		if best == nil {
			best = rel
			bestTime = t
			bestHasDate = hasDate
			continue
		}

		if hasDate && bestHasDate {
			if t.After(bestTime) {
				best = rel
				bestTime = t
			} else if t.Equal(bestTime) {
				if IsVersionNewer(best.Version, rel.Version) {
					best = rel
					bestTime = t
				}
			}
		} else if hasDate && !bestHasDate {
			best = rel
			bestTime = t
			bestHasDate = true
		} else if !hasDate && !bestHasDate {
			if IsVersionNewer(best.Version, rel.Version) {
				best = rel
			}
		}
	}
	return best
}

// RecommendedMariaDBRelease is an alias for RecommendedMariaDBSeries.
func RecommendedMariaDBRelease(releases []LiveRelease) *LiveRelease {
	return RecommendedMariaDBSeries(releases)
}

// RecommendedMariaDBUpdate finds the update among the given MariaDB updates that has the latest EOLDate.
func RecommendedMariaDBUpdate(updates []RuntimeUpdateInfo) *RuntimeUpdateInfo {
	if len(updates) == 0 {
		return nil
	}
	var best *RuntimeUpdateInfo
	var bestTime time.Time
	var bestHasDate bool

	for i := range updates {
		u := &updates[i]
		t, err := time.Parse("2006-01-02", strings.TrimSpace(u.EOLDate))
		hasDate := (err == nil && !t.IsZero())

		if best == nil {
			best = u
			bestTime = t
			bestHasDate = hasDate
			continue
		}

		if hasDate && bestHasDate {
			if t.After(bestTime) {
				best = u
				bestTime = t
			} else if t.Equal(bestTime) {
				if IsVersionNewer(best.LatestVersion, u.LatestVersion) {
					best = u
					bestTime = t
				}
			}
		} else if hasDate && !bestHasDate {
			best = u
			bestTime = t
			bestHasDate = true
		} else if !hasDate && !bestHasDate {
			if IsVersionNewer(best.LatestVersion, u.LatestVersion) {
				best = u
			}
		}
	}
	return best
}
