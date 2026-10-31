package adminer

import (
	_ "embed"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

//go:embed adminer.php
var DefaultAdminerScript []byte

const (
	DefaultDir      = "data/adminer"
	DefaultFile     = "index.php"
	DefaultCoreFile = "adminer-core.php"
	DefaultPort     = 3307
)

// IsInstalled checks if data/adminer/index.php is present
func IsInstalled(baseDir string) bool {
	targetFile := filepath.Join(baseDir, DefaultDir, DefaultFile)
	info, err := os.Stat(targetFile)
	return err == nil && !info.IsDir()
}

// EnsureInstalled writes the embedded adminer.php and a seamless passwordless wrapper to data/adminer/
func EnsureInstalled(baseDir string) (string, error) {
	targetDir := filepath.Join(baseDir, DefaultDir)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create adminer directory: %w", err)
	}

	coreFile := filepath.Join(targetDir, DefaultCoreFile)
	if _, err := os.Stat(coreFile); err != nil {
		if len(DefaultAdminerScript) == 0 {
			return "", fmt.Errorf("embedded adminer script is empty")
		}
		if err := os.WriteFile(coreFile, DefaultAdminerScript, 0644); err != nil {
			return "", fmt.Errorf("failed to write adminer core script: %w", err)
		}
	}

	indexFile := filepath.Join(targetDir, DefaultFile)
	wrapperContent := `<?php
/**
 * LaraGal Adminer Local Development Wrapper
 * Automatically allows seamless passwordless login for local root/development database access
 * and automatically creates the requested database if it does not exist yet.
 */
@ini_set('display_errors', '0');
@ini_set('display_startup_errors', '0');
@error_reporting(0);

// Pre-authenticate session for local development
if (session_status() !== PHP_SESSION_ACTIVE) {
    @session_start();
}

if (isset($_GET['sqlite'])) {
    if (!isset($_SESSION['pwds']['sqlite'])) {
        $_SESSION['pwds']['sqlite'] = array('' => array('' => ''));
    }
    if (!empty($_GET['db'])) {
        $_SESSION['db']['sqlite'][''][''][$_GET['db']] = true;
    }
} else {
    if (!isset($_SESSION['pwds']['server']['127.0.0.1']['root'])) {
        $_SESSION['pwds']['server']['127.0.0.1']['root'] = '';
    }
    // If a specific database is requested in URL, ensure it exists in MySQL
    if (!empty($_GET['db'])) {
        $rawDb = trim((string)$_GET['db']);
        $normDb = preg_replace('/[^a-zA-Z0-9_]/', '_', $rawDb);
        $normDb = trim(preg_replace('/_+/', '_', $normDb), '_');
        if ($normDb !== '' && preg_match('/^[a-zA-Z0-9_]+$/', $normDb)) {
            try {
                $mysqli = @new mysqli("127.0.0.1", "root", "", "", 3306);
                if (!$mysqli->connect_errno) {
                    @$mysqli->query("CREATE DATABASE IF NOT EXISTS " . chr(96) . $normDb . chr(96) . " CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci");
                    $mysqli->close();
                }
            } catch (\Throwable $e) {}
        }
        if ($rawDb !== $normDb && !empty($normDb)) {
            $_GET['db'] = $normDb;
        }
        $_SESSION['db']['server']['127.0.0.1']['root'][$_GET['db']] = true;
    }
}

function adminer_object() {
    class AdminerCustomization extends Adminer {
        function login($login, $password) {
            // Allow root and local development users to connect without password
            return true;
        }
        function credentials() {
            if (isset($_GET['sqlite'])) {
                return array('', '', '');
            }
            return array('127.0.0.1', 'root', '');
        }
        function name() {
            return 'LaraGal Database Manager (Adminer)';
        }
        function database() {
            if (isset($_GET['sqlite']) && !empty($_GET['db'])) {
                return $_GET['db'];
            }
            return parent::database();
        }
    }
    return new AdminerCustomization;
}

if (file_exists(__DIR__ . '/adminer-core.php')) {
    include __DIR__ . '/adminer-core.php';
} elseif (file_exists(__DIR__ . '/adminer.php')) {
    include __DIR__ . '/adminer.php';
}
`
	if err := os.WriteFile(indexFile, []byte(wrapperContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write adminer wrapper: %w", err)
	}

	return indexFile, nil
}

// BuildURL constructs the pre-authenticated Adminer URL (supporting MySQL and SQLite)
func BuildURL(domain string, port int, dbName string) string {
	baseURL := fmt.Sprintf("http://localhost:%d", port)
	if port <= 0 {
		if domain != "" {
			baseURL = fmt.Sprintf("http://%s", domain)
		} else {
			baseURL = fmt.Sprintf("http://localhost:%d", DefaultPort)
		}
	}

	cleanDB := strings.TrimSpace(dbName)
	lower := strings.ToLower(cleanDB)
	if strings.HasSuffix(lower, ".sqlite") || strings.HasSuffix(lower, ".db") {
		return fmt.Sprintf("%s/?sqlite=&username=&db=%s", baseURL, url.QueryEscape(cleanDB))
	}

	query := "server=127.0.0.1&username=root"
	if cleanDB != "" {
		query += fmt.Sprintf("&db=%s", url.QueryEscape(cleanDB))
	}

	return fmt.Sprintf("%s/?%s", baseURL, query)
}
