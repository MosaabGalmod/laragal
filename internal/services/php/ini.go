package php

import (
	"os"
	"path/filepath"
	"strings"
)

// EnsurePHPIni generates a production/development php.ini with essential web extensions enabled if php.ini is missing
func EnsurePHPIni(phpDir string) error {
	iniPath := filepath.Join(phpDir, "php.ini")
	if _, err := os.Stat(iniPath); err == nil {
		return nil // php.ini already exists, do not overwrite user customization
	}

	devPath := filepath.Join(phpDir, "php.ini-development")
	prodPath := filepath.Join(phpDir, "php.ini-production")

	var baseContent string
	if data, err := os.ReadFile(devPath); err == nil {
		baseContent = string(data)
	} else if data, err := os.ReadFile(prodPath); err == nil {
		baseContent = string(data)
	} else {
		// Minimal fallback php.ini
		baseContent = `[PHP]
engine = On
short_open_tag = Off
precision = 14
output_buffering = 4096
zlib.output_compression = Off
implicit_flush = Off
max_execution_time = 120
max_input_time = 60
memory_limit = 512M
error_reporting = E_ALL & ~E_DEPRECATED & ~E_STRICT
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
`
		return os.WriteFile(iniPath, []byte(baseContent), 0644)
	}

	// Un-comment crucial extensions
	replacements := map[string]string{
		";extension_dir = \"ext\"": "extension_dir = \"ext\"",
		";extension=curl":          "extension=curl",
		";extension=fileinfo":      "extension=fileinfo",
		";extension=mbstring":      "extension=mbstring",
		";extension=openssl":       "extension=openssl",
		";extension=pdo_mysql":     "extension=pdo_mysql",
		";extension=mysqli":        "extension=mysqli",
	}

	content := baseContent
	for target, rep := range replacements {
		content = strings.ReplaceAll(content, target, rep)
	}

	// Ensure memory_limit is at least 256M for composer / laravel
	if strings.Contains(content, "memory_limit = 128M") {
		content = strings.ReplaceAll(content, "memory_limit = 128M", "memory_limit = 512M")
	}

	return os.WriteFile(iniPath, []byte(content), 0644)
}
