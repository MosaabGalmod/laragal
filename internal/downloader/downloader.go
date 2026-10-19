package downloader

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var defaultURLs = map[string]string{
	"caddy":   "https://github.com/caddyserver/caddy/releases/download/v2.8.4/caddy_2.8.4_windows_amd64.zip",
	"mailpit": "https://github.com/axllent/mailpit/releases/download/v1.21.8/mailpit-windows-amd64.zip",
	"php-8.4": "https://windows.php.net/downloads/releases/php-8.4.4-nts-Win32-vs17-x64.zip",
	"php-8.3": "https://windows.php.net/downloads/releases/php-8.3.17-nts-Win32-vs16-x64.zip",
}

func GetDownloadURL(name string) (string, error) {
	url, exists := defaultURLs[strings.ToLower(name)]
	if !exists {
		return "", fmt.Errorf("no known download url for binary: %s", name)
	}
	return url, nil
}

func ExtractZip(zipPath string, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	for _, f := range r.File {
		targetPath := filepath.Join(destDir, f.Name)

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
