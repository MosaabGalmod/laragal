package php

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// EnsurePHPIni generates a production/development php.ini with essential web extensions enabled
// and updates existing php.ini to ensure required extensions (e.g. mysqli, pdo_mysql) are active.
func EnsurePHPIni(phpDir string) error {
	iniPath := filepath.Join(phpDir, "php.ini")
	devPath := filepath.Join(phpDir, "php.ini-development")
	prodPath := filepath.Join(phpDir, "php.ini-production")

	var content string
	if data, err := os.ReadFile(iniPath); err == nil {
		content = string(data)
	} else if data, err := os.ReadFile(devPath); err == nil {
		content = string(data)
	} else if data, err := os.ReadFile(prodPath); err == nil {
		content = string(data)
	} else {
		// Minimal fallback php.ini
		content = `[PHP]
engine = On
short_open_tag = Off
precision = 14
output_buffering = 4096
zlib.output_compression = Off
implicit_flush = Off
max_execution_time = 120
max_input_time = 60
memory_limit = 512M
error_reporting = E_ALL & ~E_DEPRECATED & ~E_USER_DEPRECATED & ~E_STRICT
display_errors = On
display_startup_errors = On
post_max_size = 100M
upload_max_filesize = 100M
default_socket_timeout = 60

extension_dir = "ext"
extension=curl
extension=fileinfo
extension=mbstring
extension=openssl
extension=pdo_mysql
extension=mysqli
extension=zip
extension=gd
`
		return os.WriteFile(iniPath, []byte(content), 0644)
	}

	// Required extensions that must be active
	extensions := []string{
		"curl",
		"fileinfo",
		"mbstring",
		"openssl",
		"pdo_mysql",
		"mysqli",
		"gd",
		"pdo_sqlite",
		"sqlite3",
		"zip",
	}

	// Ensure extension_dir uses clean path (preferring absolute if ext directory exists)
	extPath := "ext"
	if absExt, err := filepath.Abs(filepath.Join(phpDir, "ext")); err == nil {
		if info, statErr := os.Stat(absExt); statErr == nil && info.IsDir() {
			extPath = filepath.ToSlash(absExt)
		}
	}

	if strings.Contains(content, "extension_dir =") {
		// Update existing extension_dir
		lines := strings.Split(content, "\n")
		for i, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "extension_dir =") || strings.HasPrefix(trimmed, ";extension_dir =") {
				lines[i] = fmt.Sprintf("extension_dir = \"%s\"", extPath)
			}
		}
		content = strings.Join(lines, "\n")
	} else {
		content += fmt.Sprintf("\nextension_dir = \"%s\"\n", extPath)
	}

	// Ensure all required extensions are uncommented
	for _, ext := range extensions {
		if ext == "zip" {
			// Only enable zip extension if the dll actually exists in ext folder
			// (PHP 8.1+ often compiles zip directly into php8.dll, so php_zip.dll does not exist)
			zipDll := filepath.Join(phpDir, "ext", "php_zip.dll")
			if _, statErr := os.Stat(zipDll); statErr != nil {
				// Comment out or remove extension=zip to prevent startup warning
				content = strings.ReplaceAll(content, "\nextension=zip\n", "\n;extension=zip\n")
				continue
			}
		}

		commented := ";" + "extension=" + ext
		active := "extension=" + ext
		if strings.Contains(content, commented) {
			content = strings.ReplaceAll(content, commented, active)
		} else if !strings.Contains(content, active) {
			content += "\n" + active + "\n"
		}
	}

	// Ensure deprecation notices are suppressed to keep phpMyAdmin and frameworks clean
	if strings.Contains(content, "error_reporting = E_ALL") && !strings.Contains(content, "~E_DEPRECATED") {
		content = strings.ReplaceAll(content, "error_reporting = E_ALL", "error_reporting = E_ALL & ~E_DEPRECATED & ~E_USER_DEPRECATED & ~E_STRICT")
	}

	// Ensure memory_limit is at least 512M
	if strings.Contains(content, "memory_limit = 128M") {
		content = strings.ReplaceAll(content, "memory_limit = 128M", "memory_limit = 512M")
	}

	return os.WriteFile(iniPath, []byte(content), 0644)
}
