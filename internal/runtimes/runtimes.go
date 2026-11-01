package runtimes

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

// Runtime represents a binary runtime package with its official integrity checksum and metadata.
type Runtime struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	URL     string `json:"url"`
	SHA256  string `json:"sha256"`
	EOL     bool   `json:"eol"`
	EOLDate string `json:"eol_date,omitempty"`
	PHPPort int    `json:"php_port,omitempty"`
}

// catalog defines the single canonical source of truth for all bundled runtimes in LaraGal.
var catalog = map[string]Runtime{
	"caddy": {
		Name:    "caddy",
		Version: "2.8.4",
		URL:     "https://github.com/caddyserver/caddy/releases/download/v2.8.4/caddy_2.8.4_windows_amd64.zip",
		SHA256:  "2631fbe1c4f6fb08783ed6dc6ac238d6655b4c923abe2bce1433dbd7fd079630",
		EOL:     false,
		PHPPort: 0,
	},
	"mailpit": {
		Name:    "mailpit",
		Version: "1.21.8",
		URL:     "https://github.com/axllent/mailpit/releases/download/v1.21.8/mailpit-windows-amd64.zip",
		SHA256:  "934dbd469da0f61ec7d1619e2d3c5a8c6179b7beab4a6d1900c3916ee2d547fb",
		EOL:     false,
		PHPPort: 0,
	},
	"php-8.5": {
		Name:    "php-8.5",
		Version: "8.5.12-1",
		URL:     "https://windows.php.net/downloads/releases/php-8.5.12-1-nts-Win32-vs17-x64.zip",
		SHA256:  "2a36a2edcb7a5b9252f638fd5e4cf84d7f6e066092654220dc0b63e5cf4e9723",
		EOL:     false,
		PHPPort: 9085,
	},
	"php-8.4": {
		Name:    "php-8.4",
		Version: "8.4.27-1",
		URL:     "https://windows.php.net/downloads/releases/php-8.4.27-1-nts-Win32-vs17-x64.zip",
		SHA256:  "dbae154a0a89791f36c3d48e68bae797878333e7d6cb5721e74617de59b63ab7",
		EOL:     false,
		PHPPort: 9084,
	},
	"php-8.3": {
		Name:    "php-8.3",
		Version: "8.3.36-1",
		URL:     "https://windows.php.net/downloads/releases/php-8.3.36-1-nts-Win32-vs16-x64.zip",
		SHA256:  "1a097d52b7a68622cdc723073b61cc70caf79c941218ff65b2d8441dbc6b10d4",
		EOL:     false,
		PHPPort: 9083,
	},
	"php-8.2": {
		Name:    "php-8.2",
		Version: "8.2.35-1",
		URL:     "https://windows.php.net/downloads/releases/php-8.2.35-1-nts-Win32-vs16-x64.zip",
		SHA256:  "ebe75134c03869826cb1714d31fd5eb96afefa6d9fd35d786971c03020e7bce2",
		EOL:     false,
		EOLDate: "2026-12-31",
		PHPPort: 9082,
	},
	"php-8.1": {
		Name:    "php-8.1",
		Version: "8.1.34",
		URL:     "https://windows.php.net/downloads/releases/php-8.1.34-nts-Win32-vs16-x64.zip",
		SHA256:  "9cfe246cb144076c16f5913a3ef88a474c3dd7e60f0f0c8bb95faf68674016cc",
		EOL:     true,
		EOLDate: "2025-12-31",
		PHPPort: 9081,
	},
	"mysql": {
		Name:    "mysql",
		Version: "10.11.10",
		URL:     "https://archive.mariadb.org/mariadb-10.11.10/winx64-packages/mariadb-10.11.10-winx64.zip",
		SHA256:  "5626478891370011b568e3776ba76d840846c2a08261b14b82de97f0608311cf",
		EOL:     false,
		PHPPort: 0,
	},
	"redis": {
		Name:    "redis",
		Version: "5.0.14.1",
		URL:     "https://github.com/tporadowski/redis/releases/download/v5.0.14.1/Redis-x64-5.0.14.1.zip",
		SHA256:  "018ea18a35876383cbb5f4cd0258adfc87747cf9d619bce1cf73a2e36f720ccf",
		EOL:     false,
		PHPPort: 0,
	},
	"adminer": {
		Name:    "adminer",
		Version: "4.8.1",
		URL:     "https://github.com/vrana/adminer/releases/download/v4.8.1/adminer-4.8.1.php",
		SHA256:  "2fd7e6d8f987b243ab1839249551f62adce19704c47d3d0c8dd9e57ea5b9c6b3",
		EOL:     false,
		PHPPort: 0,
	},
	"phpmyadmin": {
		Name:    "phpmyadmin",
		Version: "5.2.2",
		URL:     "https://files.phpmyadmin.net/phpMyAdmin/5.2.2/phpMyAdmin-5.2.2-all-languages.zip",
		SHA256:  "6b99534f72ffb1d7275f50d23ca4141e1495c97d7cadb73a41d6dc580ed5ce29",
		EOL:     false,
		PHPPort: 0,
	},
}

// NormalizeName converts any alias or version string into the canonical catalog key.
func NormalizeName(name string) string {
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

// All returns all runtimes in the catalog ordered by canonical priority.
func All() []Runtime {
	order := []string{
		"caddy",
		"php-8.5",
		"php-8.4",
		"php-8.3",
		"php-8.2",
		"php-8.1",
		"mysql",
		"redis",
		"mailpit",
		"phpmyadmin",
		"adminer",
	}

	result := make([]Runtime, 0, len(catalog))
	seen := make(map[string]bool)
	for _, key := range order {
		if rt, ok := catalog[key]; ok {
			result = append(result, rt)
			seen[key] = true
		}
	}
	for key, rt := range catalog {
		if !seen[key] {
			result = append(result, rt)
		}
	}
	return result
}

// CatalogMap returns a copy of the canonical catalog map.
func CatalogMap() map[string]Runtime {
	res := make(map[string]Runtime, len(catalog))
	for k, v := range catalog {
		res[k] = v
	}
	return res
}

// Get returns the runtime by canonical name or alias.
func Get(name string) (Runtime, bool) {
	norm := NormalizeName(name)
	rt, exists := catalog[norm]
	return rt, exists
}

// GetURL returns the download URL for a given runtime name or alias.
func GetURL(name string) (string, error) {
	norm := NormalizeName(name)
	if rt, ok := catalog[norm]; ok {
		return rt.URL, nil
	}

	// For future/unlisted PHP versions, provide dynamic pattern
	if strings.HasPrefix(norm, "php-") {
		ver := strings.TrimPrefix(norm, "php-")
		return fmt.Sprintf("https://windows.php.net/downloads/releases/php-%s.0-nts-Win32-vs17-x64.zip", ver), nil
	}

	return "", fmt.Errorf("no known download artifact for binary: %s", name)
}

// GetExpectedSHA256 returns the configured SHA-256 for a name, alias, URL, or filename.
func GetExpectedSHA256(nameOrURL string) string {
	target := strings.TrimSpace(nameOrURL)
	if target == "" {
		return ""
	}

	// 1. Try normalized catalog lookup
	norm := NormalizeName(target)
	if rt, exists := catalog[norm]; exists {
		return rt.SHA256
	}

	// 2. Try exact URL match
	for _, rt := range catalog {
		if rt.URL == target {
			return rt.SHA256
		}
	}

	// 3. Try filename in URL match
	targetFilename := filepath.Base(target)
	if targetFilename != "" && targetFilename != "." && targetFilename != "/" {
		for _, rt := range catalog {
			if filepath.Base(rt.URL) == targetFilename {
				return rt.SHA256
			}
		}
	}

	return ""
}

// NormalizePHPVersion returns standard minor version string (e.g. "8.5", "8.4", "8.3", "8.2", "8.1").
func NormalizePHPVersion(version string) string {
	v := strings.TrimSpace(strings.ToLower(version))
	v = strings.TrimPrefix(v, "php")
	v = strings.TrimPrefix(v, "-")
	v = strings.TrimSpace(v)
	if len(v) == 2 && v[0] >= '7' && v[0] <= '9' && v[1] >= '0' && v[1] <= '9' {
		return fmt.Sprintf("%c.%c", v[0], v[1])
	}
	if strings.Contains(v, ".") {
		parts := strings.Split(v, ".")
		if len(parts) >= 2 {
			return fmt.Sprintf("%s.%s", parts[0], parts[1])
		}
	}
	if v != "" {
		return v
	}
	return "8.4"
}

// PHPPortForVersion returns the FastCGI port for a PHP version string.
// Checks catalog first, then uses dynamic 9000 + (major * 10) + minor.
func PHPPortForVersion(version string) int {
	normMinor := NormalizePHPVersion(version)
	key := fmt.Sprintf("php-%s", normMinor)
	if rt, ok := catalog[key]; ok && rt.PHPPort > 0 {
		return rt.PHPPort
	}

	parts := strings.Split(normMinor, ".")
	if len(parts) == 2 {
		major, err1 := strconv.Atoi(parts[0])
		minor, err2 := strconv.Atoi(parts[1])
		if err1 == nil && err2 == nil && major >= 7 && major <= 15 {
			port := 9000 + (major * 10) + minor
			if port >= 9000 && port <= 9999 {
				return port
			}
		}
	}

	return 9083
}

// PHPPortMapping returns the map of PHP minor versions to FastCGI ports.
func PHPPortMapping() map[string]int {
	mapping := make(map[string]int)
	for _, rt := range catalog {
		if strings.HasPrefix(rt.Name, "php-") && rt.PHPPort > 0 {
			minor := strings.TrimPrefix(rt.Name, "php-")
			mapping[minor] = rt.PHPPort
		}
	}
	return mapping
}
