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

	"github.com/laragal/laragal/internal/runtimes"
)

// Artifact represents a downloadable binary runtime package with its official integrity checksum.
type Artifact struct {
	URL    string `json:"url"`
	SHA256 string `json:"sha256"`
}

// GetDownloadArtifact returns the artifact metadata (URL and SHA-256) for known portable runtimes.
func GetDownloadArtifact(name string) (Artifact, error) {
	norm := runtimes.NormalizeName(name)
	if rt, exists := runtimes.Get(norm); exists {
		return Artifact{
			URL:    rt.URL,
			SHA256: rt.SHA256,
		}, nil
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
	return runtimes.GetExpectedSHA256(nameOrURL)
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
