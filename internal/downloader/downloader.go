package downloader

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
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

	return Artifact{}, fmt.Errorf("unknown runtime or version: %s (no verified artifact found)", name)
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

func getArchiveFallbackURL(rawURL string) string {
	if strings.Contains(rawURL, "/downloads/releases/") && !strings.Contains(rawURL, "/downloads/releases/archives/") {
		return strings.Replace(rawURL, "/downloads/releases/", "/downloads/releases/archives/", 1)
	}
	return ""
}

// DownloadFile streams a remote binary to a temporary file, computes its checksum (SHA-256 or SHA-512),
// verifies integrity before placing the file into destPath, and reports progress.
// Supports passing checksum as value (auto-detects sha256 vs sha512) or algo + value ("sha512", "<hash>").
// On hash mismatch, the temporary file is deleted and an error is returned.
func DownloadFile(ctx context.Context, url string, destPath string, onProgress ProgressCallback, checksumArgs ...string) error {
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

	// For windows.php.net release URLs, on HTTP 404 retry under /downloads/releases/archives/
	if resp.StatusCode == http.StatusNotFound {
		if archiveURL := getArchiveFallbackURL(url); archiveURL != "" {
			_ = resp.Body.Close()
			archiveReq, aErr := http.NewRequestWithContext(ctx, "GET", archiveURL, nil)
			if aErr == nil {
				archiveReq.Header.Set("User-Agent", "LaraGal-Downloader/1.0")
				if aResp, doErr := client.Do(archiveReq); doErr == nil {
					resp = aResp
				}
			}
		}
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

	// Determine expected checksum and hashing algorithm (sha256 vs sha512)
	var expected string
	algo := "sha256"

	if len(checksumArgs) == 1 {
		val := strings.TrimSpace(checksumArgs[0])
		if len(val) == 128 {
			algo = "sha512"
			expected = val
		} else {
			expected = val
		}
	} else if len(checksumArgs) >= 2 {
		arg1 := strings.ToLower(strings.TrimSpace(checksumArgs[0]))
		arg2 := strings.ToLower(strings.TrimSpace(checksumArgs[1]))
		if arg1 == "sha512" || arg1 == "sha256" {
			algo = arg1
			expected = strings.TrimSpace(checksumArgs[1])
		} else if arg2 == "sha512" || arg2 == "sha256" {
			algo = arg2
			expected = strings.TrimSpace(checksumArgs[0])
		} else {
			expected = strings.TrimSpace(checksumArgs[0])
		}
	} else {
		expected = GetExpectedSHA256(url)
	}

	var hasher hash.Hash
	if strings.EqualFold(algo, "sha512") {
		hasher = sha512.New()
	} else {
		hasher = sha256.New()
	}

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

	actualHash := hex.EncodeToString(hasher.Sum(nil))

	if expected != "" {
		if !strings.EqualFold(expected, actualHash) {
			return fmt.Errorf("%s checksum mismatch for %s: expected %s, got %s", strings.ToUpper(algo), url, expected, actualHash)
		}
	} else {
		log.Printf("WARNING: no %s checksum configured for %s; skipping integrity verification", strings.ToUpper(algo), url)
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
	if prefix == "bin" || prefix == "lib" || prefix == "ext" || prefix == "share" || prefix == "etc" {
		return ""
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
