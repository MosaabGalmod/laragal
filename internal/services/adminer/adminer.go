package adminer

import (
	_ "embed"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
)

//go:embed adminer.php
var DefaultAdminerScript []byte

const (
	DefaultDir  = "data/adminer"
	DefaultFile = "index.php"
	DefaultPort = 3307
)

// EnsureInstalled writes the embedded adminer.php to data/adminer/index.php if not present
func EnsureInstalled(baseDir string) (string, error) {
	targetDir := filepath.Join(baseDir, "data", "adminer")
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create adminer directory: %w", err)
	}

	targetFile := filepath.Join(targetDir, DefaultFile)
	if _, err := os.Stat(targetFile); err == nil {
		return targetFile, nil
	}

	if len(DefaultAdminerScript) == 0 {
		return "", fmt.Errorf("embedded adminer script is empty")
	}

	if err := os.WriteFile(targetFile, DefaultAdminerScript, 0644); err != nil {
		return "", fmt.Errorf("failed to write adminer script: %w", err)
	}

	return targetFile, nil
}

// BuildURL constructs the pre-authenticated Adminer URL
func BuildURL(domain string, port int, dbName string) string {
	baseURL := fmt.Sprintf("http://localhost:%d", port)
	if port <= 0 {
		if domain != "" {
			baseURL = fmt.Sprintf("http://%s", domain)
		} else {
			baseURL = fmt.Sprintf("http://localhost:%d", DefaultPort)
		}
	}

	query := "server=127.0.0.1&username=root"
	if dbName != "" {
		query += fmt.Sprintf("&db=%s", url.QueryEscape(dbName))
	}

	return fmt.Sprintf("%s/?%s", baseURL, query)
}
