package php

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var phpVersionRegex = regexp.MustCompile(`(?i)PHP\s+(\d+\.\d+)`)
var dirVersionRegex = regexp.MustCompile(`(?:php-?|^)(\d+\.\d+)`)
var opcacheZendExtRegex = regexp.MustCompile(`(?i)^\s*(;?)\s*(zend_extension\s*=\s*"?([a-zA-Z0-9_-]*opcache[a-zA-Z0-9_.-]*)"?)`)

// DetectPHPVersion attempts to detect the PHP version from the directory path or binary.
func DetectPHPVersion(phpDir string) string {
	base := filepath.Base(phpDir)
	if m := dirVersionRegex.FindStringSubmatch(base); len(m) > 1 {
		return m[1]
	}
	if m := dirVersionRegex.FindStringSubmatch(phpDir); len(m) > 1 {
		return m[1]
	}
	for _, exeName := range []string{"php.exe", "php-cgi.exe"} {
		exePath := filepath.Join(phpDir, exeName)
		if _, err := os.Stat(exePath); err == nil {
			cmd := exec.Command(exePath, "-v")
			if out, err := cmd.CombinedOutput(); err == nil {
				if m := phpVersionRegex.FindStringSubmatch(string(out)); len(m) > 1 {
					return m[1]
				}
			}
		}
	}
	return ""
}

// IsPHP85OrHigher returns true if the detected or supplied version is >= 8.5
func IsPHP85OrHigher(ver string) bool {
	clean := strings.TrimSpace(ver)
	clean = strings.TrimPrefix(clean, "php-")
	clean = strings.TrimPrefix(clean, "PHP-")
	clean = strings.TrimPrefix(clean, "v")
	clean = strings.TrimPrefix(clean, "V")
	if m := dirVersionRegex.FindStringSubmatch(clean); len(m) > 1 {
		clean = m[1]
	}

	parts := strings.Split(clean, ".")
	if len(parts) == 0 {
		return false
	}
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return false
	}
	minor := 0
	if len(parts) > 1 {
		minor, _ = strconv.Atoi(parts[1])
	}
	return major > 8 || (major == 8 && minor >= 5)
}

func processOPcacheConfig(content string, is85 bool) string {
	lines := strings.Split(content, "\n")
	foundZendExt := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		m := opcacheZendExtRegex.FindStringSubmatch(trimmed)
		if len(m) > 0 {
			foundZendExt = true
			isCommented := m[1] == ";"
			extDirective := m[2]
			if is85 {
				if !isCommented {
					lines[i] = ";" + line
				}
			} else {
				if isCommented {
					lines[i] = extDirective
				}
			}
		}
	}

	result := strings.Join(lines, "\n")

	if !foundZendExt {
		if is85 {
			if strings.Contains(result, "[opcache]") {
				result = strings.Replace(result, "[opcache]", ";zend_extension=opcache\n[opcache]", 1)
			}
		} else {
			if strings.Contains(result, "[opcache]") {
				result = strings.Replace(result, "[opcache]", "zend_extension=opcache\n[opcache]", 1)
			} else {
				result += "\nzend_extension=opcache\n"
			}
		}
	}

	if !strings.Contains(result, "[opcache]") {
		opcacheBlock := `
[opcache]
opcache.enable = 1
opcache.enable_cli = 1
opcache.memory_consumption = 128
opcache.interned_strings_buffer = 8
opcache.max_accelerated_files = 10000
`
		result += opcacheBlock
	}

	return result
}

// EnsurePHPIni generates a production/development php.ini with essential web extensions enabled
// and updates existing php.ini to ensure required extensions (e.g. mysqli, pdo_mysql) are active.
// For PHP >= 8.5, OPcache is compiled into core, so zend_extension=opcache is commented out or skipped
// while [opcache] configuration settings are retained.
func EnsurePHPIni(phpDir string, versionHint ...string) error {
	iniPath := filepath.Join(phpDir, "php.ini")
	devPath := filepath.Join(phpDir, "php.ini-development")
	prodPath := filepath.Join(phpDir, "php.ini-production")

	version := ""
	if len(versionHint) > 0 && strings.TrimSpace(versionHint[0]) != "" {
		version = strings.TrimSpace(versionHint[0])
	} else {
		version = DetectPHPVersion(phpDir)
	}
	is85 := IsPHP85OrHigher(version)

	var content string
	if data, err := os.ReadFile(iniPath); err == nil {
		content = string(data)
	} else if data, err := os.ReadFile(devPath); err == nil {
		content = string(data)
	} else if data, err := os.ReadFile(prodPath); err == nil {
		content = string(data)
	} else {
		// Minimal fallback php.ini
		opcacheExtLine := "zend_extension=opcache"
		if is85 {
			opcacheExtLine = ";zend_extension=opcache"
		}
		content = fmt.Sprintf(`[PHP]
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

%s
[opcache]
opcache.enable = 1
opcache.enable_cli = 1
opcache.memory_consumption = 128
opcache.interned_strings_buffer = 8
opcache.max_accelerated_files = 10000
`, opcacheExtLine)
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

	// Handle OPcache for PHP >= 8.5 vs < 8.5
	content = processOPcacheConfig(content, is85)

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
