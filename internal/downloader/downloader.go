package downloader

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Artifact represents a downloadable binary runtime package with its official integrity checksum.
type Artifact struct {
	URL    string `json:"url"`
	SHA256 string `json:"sha256"`
}

var defaultArtifacts = map[string]Artifact{
	"caddy": {
		URL:    "https://github.com/caddyserver/caddy/releases/download/v2.8.4/caddy_2.8.4_windows_amd64.zip",
		SHA256: "2631fbe1c4f6fb08783ed6dc6ac238d6655b4c923abe2bce1433dbd7fd079630",
	},
	"mailpit": {
		URL:    "https://github.com/axllent/mailpit/releases/download/v1.21.8/mailpit-windows-amd64.zip",
		SHA256: "934dbd469da0f61ec7d1619e2d3c5a8c6179b7beab4a6d1900c3916ee2d547fb",
	},
	"php-8.5": {
		URL:    "https://windows.php.net/downloads/releases/php-8.5.12-1-nts-Win32-vs17-x64.zip",
		SHA256: "2a36a2edcb7a5b9252f638fd5e4cf84d7f6e066092654220dc0b63e5cf4e9723",
	},
	"php-8.4": {
		URL:    "https://windows.php.net/downloads/releases/archives/php-8.4.4-nts-Win32-vs17-x64.zip",
		SHA256: "f2269ec0bdd74c1ba8db58c13b559592cc0ea2acbb0c3886a7ca5b7f2e05f1b1",
	},
	"php-8.3": {
		URL:    "https://windows.php.net/downloads/releases/archives/php-8.3.17-nts-Win32-vs16-x64.zip",
		SHA256: "630fec87d618c3f221864255cc2076471421dfaf7352779af3b7f8c76e00e8bd",
	},
	"php-8.2": {
		URL:    "https://windows.php.net/downloads/releases/archives/php-8.2.27-nts-Win32-vs16-x64.zip",
		SHA256: "ab8acf92d3b32425d82cea66c190e56969e61f15b7878730a112446d97521515",
	},
	"php-8.1": {
		URL:    "https://windows.php.net/downloads/releases/archives/php-8.1.31-nts-Win32-vs16-x64.zip",
		SHA256: "77feab37229359a897f602aa86dc5aff0cc7e414c0f21afc3098ae4e991a9e8f",
	},
	"mysql": {
		URL:    "https://archive.mariadb.org/mariadb-10.11.10/winx64-packages/mariadb-10.11.10-winx64.zip",
		SHA256: "5626478891370011b568e3776ba76d840846c2a08261b14b82de97f0608311cf",
	},
	"redis": {
		URL:    "https://github.com/tporadowski/redis/releases/download/v5.0.14.1/Redis-x64-5.0.14.1.zip",
		SHA256: "018ea18a35876383cbb5f4cd0258adfc87747cf9d619bce1cf73a2e36f720ccf",
	},
	"adminer": {
		URL:    "https://github.com/vrana/adminer/releases/download/v4.8.1/adminer-4.8.1.php",
		SHA256: "2fd7e6d8f987b243ab1839249551f62adce19704c47d3d0c8dd9e57ea5b9c6b3",
	},
	"phpmyadmin": {
		URL:    "https://files.phpmyadmin.net/phpMyAdmin/5.2.2/phpMyAdmin-5.2.2-all-languages.zip",
		SHA256: "6b99534f72ffb1d7275f50d23ca4141e1495c97d7cadb73a41d6dc580ed5ce29",
	},
}

var defaultURLs = func() map[string]string {
	m := make(map[string]string, len(defaultArtifacts))
	for k, v := range defaultArtifacts {
		m[k] = v.URL
	}
	return m
}()

func normalizeBinaryName(name string) string {
	n := strings.ToLower(strings.TrimSpace(name))
	if n == "mariadb" {
		return "mysql"
	}
	if n == "phpmyadmin" || n == "pma" {
		return "phpmyadmin"
	}
	if strings.HasPrefix(n, "php") {
		v := strings.TrimPrefix(n, "php")
		v = strings.TrimPrefix(v, "-")
		v = strings.TrimSpace(v)
		if len(v) == 2 && v[0] >= '7' && v[0] <= '9' && v[1] >= '0' && v[1] <= '9' {
			return fmt.Sprintf("php-%c.%c", v[0], v[1])
		}
		if strings.Contains(v, ".") {
			parts := strings.Split(v, ".")
			if len(parts) >= 2 {
				return fmt.Sprintf("php-%s.%s", parts[0], parts[1])
			}
		}
		if v != "" {
			return fmt.Sprintf("php-%s", v)
		}
		return "php-8.4"
	}
	return n
}

// GetDownloadArtifact returns the artifact metadata (URL and SHA-256) for known portable runtimes.
func GetDownloadArtifact(name string) (Artifact, error) {
	norm := normalizeBinaryName(name)
	art, exists := defaultArtifacts[norm]
	if exists {
		return art, nil
	}

	// For future or unlisted PHP versions (e.g. php-8.6, php-9.0), provide dynamic official pattern
	if strings.HasPrefix(norm, "php-") {
		ver := strings.TrimPrefix(norm, "php-")
		return Artifact{
			URL:    fmt.Sprintf("https://windows.php.net/downloads/releases/php-%s.0-nts-Win32-vs17-x64.zip", ver),
			SHA256: "",
		}, nil
	}

	return Artifact{}, fmt.Errorf("no known download artifact for binary: %s", name)
}

// GetDownloadURL returns the direct download URL for known portable runtimes
func GetDownloadURL(name string) (string, error) {
	art, err := GetDownloadArtifact(name)
	if err != nil {
		return "", err
	}
	return art.URL, nil
}

// GetExpectedSHA256 returns the configured SHA-256 checksum for a binary name or download URL.
// Returns an empty string if no known checksum is configured.
func GetExpectedSHA256(nameOrURL string) string {
	target := strings.TrimSpace(nameOrURL)
	if target == "" {
		return ""
	}

	// Try normalized binary name
	norm := normalizeBinaryName(target)
	if art, exists := defaultArtifacts[norm]; exists {
		return art.SHA256
	}

	// Try matching URL directly
	for _, art := range defaultArtifacts {
		if art.URL == target {
			return art.SHA256
		}
	}

	// Try matching by filename in URL (e.g. php-8.4.4-nts-Win32-vs17-x64.zip)
	targetFilename := filepath.Base(target)
	if targetFilename != "" && targetFilename != "." && targetFilename != "/" {
		for _, art := range defaultArtifacts {
			if filepath.Base(art.URL) == targetFilename {
				return art.SHA256
			}
		}
	}

	return ""
}

// ProgressCallback reports downloaded bytes, total expected bytes, and calculated percentage
type ProgressCallback func(downloaded int64, total int64, percent int)

type progressWriter struct {
	total      int64
	downloaded int64
	callback   ProgressCallback
	lastReport time.Time
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	n := len(p)
	pw.downloaded += int64(n)

	now := time.Now()
	if pw.callback != nil {
		if now.Sub(pw.lastReport) >= 100*time.Millisecond || (pw.total > 0 && pw.downloaded >= pw.total) {
			pw.lastReport = now
			var percent int
			if pw.total > 0 {
				percent = int((pw.downloaded * 100) / pw.total)
				if percent > 100 {
					percent = 100
				}
			}
			pw.callback(pw.downloaded, pw.total, percent)
		}
	}

	return n, nil
}

// DownloadFile streams a remote binary to a temporary file, computes its SHA-256 checksum,
// verifies integrity before placing the file into destPath, and reports progress.
// On hash mismatch, the temporary file is deleted and an error is returned.
func DownloadFile(ctx context.Context, url string, destPath string, onProgress ProgressCallback, expectedSHA256 ...string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "LaraGal-Downloader/1.0")

	client := &http.Client{
		Timeout: 30 * time.Minute,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("failed to download %s: HTTP %s", url, resp.Status)
	}

	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	// Create temp file in the same directory so os.Rename is atomic (same volume)
	tempFile, err := os.CreateTemp(destDir, ".tmp-download-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary download file: %w", err)
	}
	tempPath := tempFile.Name()

	success := false
	defer func() {
		if !success {
			_ = tempFile.Close()
			_ = os.Remove(tempPath)
		}
	}()

	hasher := sha256.New()
	pw := &progressWriter{
		total:      resp.ContentLength,
		callback:   onProgress,
		lastReport: time.Now(),
	}

	writer := io.MultiWriter(tempFile, hasher, pw)
	if _, err := io.Copy(writer, resp.Body); err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("failed to finalize temporary download file: %w", err)
	}

	// Determine expected SHA-256 hash
	var expected string
	if len(expectedSHA256) > 0 && strings.TrimSpace(expectedSHA256[0]) != "" {
		expected = strings.TrimSpace(expectedSHA256[0])
	} else {
		expected = GetExpectedSHA256(url)
	}

	actualHash := hex.EncodeToString(hasher.Sum(nil))

	if expected != "" {
		if !strings.EqualFold(expected, actualHash) {
			return fmt.Errorf("SHA-256 checksum mismatch for %s: expected %s, got %s", url, expected, actualHash)
		}
	} else {
		log.Printf("WARNING: no SHA-256 checksum configured for %s; skipping integrity verification", url)
	}

	// Safely move verified file into destPath (handling Windows existing file replacement)
	_ = os.Remove(destPath)
	if err := os.Rename(tempPath, destPath); err != nil {
		return fmt.Errorf("failed to move verified file to destination %s: %w", destPath, err)
	}

	success = true

	if onProgress != nil && resp.ContentLength > 0 {
		onProgress(resp.ContentLength, resp.ContentLength, 100)
	}

	return nil
}

func detectSingleRootDir(files []*zip.File) string {
	if len(files) == 0 {
		return ""
	}
	var prefix string
	for _, f := range files {
		clean := filepath.ToSlash(filepath.Clean(f.Name))
		if clean == "." || clean == "/" {
			continue
		}
		parts := strings.Split(clean, "/")
		if len(parts) == 1 && !f.FileInfo().IsDir() {
			return ""
		}
		if prefix == "" {
			prefix = parts[0]
		} else if parts[0] != prefix {
			return ""
		}
	}
	return prefix
}

// ExtractZip extracts zip contents and automatically strips a redundant single top-level root folder
func ExtractZip(zipPath string, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	rootDir := detectSingleRootDir(r.File)

	for _, f := range r.File {
		cleanName := filepath.ToSlash(filepath.Clean(f.Name))
		if rootDir != "" {
			if cleanName == rootDir {
				continue
			}
			cleanName = strings.TrimPrefix(cleanName, rootDir+"/")
		}

		if cleanName == "" || cleanName == "." {
			continue
		}

		targetPath := filepath.Join(destDir, filepath.FromSlash(cleanName))

		// Guard against zip-slip
		if !strings.HasPrefix(filepath.Clean(targetPath), filepath.Clean(destDir)) {
			return fmt.Errorf("illegal file path in zip: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			_ = os.MkdirAll(targetPath, f.Mode())
			continue
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}

		outFile, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()

		if err != nil {
			return err
		}
	}

	return nil
}

func IsBinaryInstalled(binDir string, exeName string) bool {
	path := filepath.Join(binDir, exeName)
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
