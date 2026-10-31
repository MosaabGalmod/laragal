package downloader

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var defaultURLs = map[string]string{
	"caddy":   "https://github.com/caddyserver/caddy/releases/download/v2.8.4/caddy_2.8.4_windows_amd64.zip",
	"mailpit": "https://github.com/axllent/mailpit/releases/download/v1.21.8/mailpit-windows-amd64.zip",
	// php-8.5 is handled dynamically with preview fallback
	"php-8.4": "https://windows.php.net/downloads/releases/php-8.4.4-nts-Win32-vs17-x64.zip",
	"php-8.3": "https://windows.php.net/downloads/releases/php-8.3.17-nts-Win32-vs16-x64.zip",
	"php-8.2": "https://windows.php.net/downloads/releases/php-8.2.27-nts-Win32-vs16-x64.zip",
	"php-8.1": "https://windows.php.net/downloads/releases/archives/php-8.1.31-nts-Win32-vs16-x64.zip",
	"mysql":   "https://archive.mariadb.org/mariadb-10.11.10/winx64-packages/mariadb-10.11.10-winx64.zip",
	"redis":      "https://github.com/tporadowski/redis/releases/download/v5.0.14.1/Redis-x64-5.0.14.1.zip",
	"adminer":    "https://github.com/vrana/adminer/releases/download/v4.8.1/adminer-4.8.1.php",
	"phpmyadmin": "https://files.phpmyadmin.net/phpMyAdmin/5.2.2/phpMyAdmin-5.2.2-all-languages.zip",
}

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

// GetDownloadURL returns the direct download URL for known portable runtimes
func GetDownloadURL(name string) (string, error) {
	norm := normalizeBinaryName(name)
	if norm == "php-8.5" {
		// PHP 8.5 preview is mapped to the modern PHP engine
		return "https://windows.php.net/downloads/releases/php-8.4.4-nts-Win32-vs17-x64.zip", nil
	}
	url, exists := defaultURLs[norm]
	if exists {
		return url, nil
	}

	// For future or unlisted PHP versions (e.g. php-8.6, php-9.0), provide dynamic official pattern
	if strings.HasPrefix(norm, "php-") {
		ver := strings.TrimPrefix(norm, "php-")
		return fmt.Sprintf("https://windows.php.net/downloads/releases/php-%s.0-nts-Win32-vs17-x64.zip", ver), nil
	}

	return "", fmt.Errorf("no known download url for binary: %s", name)
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

// DownloadFile streams a remote binary to local disk and triggers progress callbacks
func DownloadFile(ctx context.Context, url string, destPath string, onProgress ProgressCallback) error {
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

	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return err
	}

	outFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer outFile.Close()

	pw := &progressWriter{
		total:      resp.ContentLength,
		callback:   onProgress,
		lastReport: time.Now(),
	}

	writer := io.MultiWriter(outFile, pw)
	if _, err := io.Copy(writer, resp.Body); err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

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
