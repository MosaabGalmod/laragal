package phpmyadmin

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

const (
	DefaultDir  = "data/phpmyadmin"
	DefaultFile = "index.php"
	DefaultPort = 3308
)

// IsInstalled checks if data/phpmyadmin/index.php is present
func IsInstalled(baseDir string) bool {
	targetFile := filepath.Join(baseDir, DefaultDir, DefaultFile)
	info, err := os.Stat(targetFile)
	return err == nil && !info.IsDir()
}

// PatchIndexPHPSilenceDeprecations ensures index.php silences PHP 8.4 deprecations before session starts
func PatchIndexPHPSilenceDeprecations(baseDir string) error {
	indexPath := filepath.Join(baseDir, DefaultDir, DefaultFile)
	data, err := os.ReadFile(indexPath)
	if err != nil {
		return err
	}
	content := string(data)
	if strings.Contains(content, "Silence deprecation warnings in PHP 8.4+") {
		return nil
	}

	target := "declare(strict_types=1);\n"
	replacement := `declare(strict_types=1);

// Silence deprecation warnings in PHP 8.4+ before headers/session are initialized
@ini_set('display_errors', '0');
@ini_set('display_startup_errors', '0');
error_reporting(E_ALL & ~E_DEPRECATED & ~E_USER_DEPRECATED);
`
	if strings.Contains(content, target) {
		content = strings.Replace(content, target, replacement, 1)
		return os.WriteFile(indexPath, []byte(content), 0644)
	}
	return nil
}

// EnsureConfig generates a seamless config.inc.php for local MariaDB/MySQL root connection
func EnsureConfig(baseDir string) error {
	targetDir := filepath.Join(baseDir, DefaultDir)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return err
	}

	_ = PatchIndexPHPSilenceDeprecations(baseDir)

	configFile := filepath.Join(targetDir, "config.inc.php")
	if _, err := os.Stat(configFile); err == nil {
		return nil
	}

	// Generate 32-character blowfish secret
	secretBytes := make([]byte, 16)
	_, _ = rand.Read(secretBytes)
	blowfishSecret := hex.EncodeToString(secretBytes)

	configContent := fmt.Sprintf(`<?php
/**
 * LaraGal phpMyAdmin Configuration
 * Automatically provisioned for zero-friction local development.
 */
declare(strict_types=1);

// Silence deprecation warnings in PHP 8.4+
@ini_set('display_errors', '0');
@ini_set('display_startup_errors', '0');
error_reporting(E_ALL & ~E_DEPRECATED & ~E_USER_DEPRECATED);

$cfg['blowfish_secret'] = '%s';

$i = 0;
$i++;
$cfg['Servers'][$i]['auth_type'] = 'config';
$cfg['Servers'][$i]['host'] = '127.0.0.1';
$cfg['Servers'][$i]['port'] = '3306';
$cfg['Servers'][$i]['user'] = 'root';
$cfg['Servers'][$i]['password'] = '';
$cfg['Servers'][$i]['AllowNoPassword'] = true;
$cfg['Servers'][$i]['compress'] = false;
$cfg['Servers'][$i]['hide_db'] = '^(information_schema|performance_schema|mysql|sys)$';

$cfg['UploadDir'] = '';
$cfg['SaveDir'] = '';
$cfg['ShowPhpInfo'] = true;
$cfg['CheckConfigurationPermissions'] = false;
`, blowfishSecret)

	return os.WriteFile(configFile, []byte(configContent), 0644)
}

// BuildURL constructs the target phpMyAdmin URL
func BuildURL(domain string, port int, dbName string) string {
	baseURL := fmt.Sprintf("http://localhost:%d", port)
	if port <= 0 {
		if domain != "" {
			baseURL = fmt.Sprintf("http://%s", domain)
		} else {
			baseURL = fmt.Sprintf("http://localhost:%d", DefaultPort)
		}
	}

	if dbName != "" {
		return fmt.Sprintf("%s/index.php?route=/database/structure&db=%s", baseURL, url.QueryEscape(dbName))
	}
	return fmt.Sprintf("%s/", baseURL)
}
