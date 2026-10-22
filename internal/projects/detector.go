package projects

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type ProjectType string

const (
	TypeLaravel   ProjectType = "laravel"
	TypeWordPress ProjectType = "wordpress"
	TypePHP       ProjectType = "php"
	TypeStatic    ProjectType = "static"
	TypeNode      ProjectType = "node"
	TypeGo        ProjectType = "go"
	TypeRust      ProjectType = "rust"
	TypeDotNet    ProjectType = "dotnet"
	TypePython    ProjectType = "python"
	TypeRuby      ProjectType = "ruby"
	TypeJava      ProjectType = "java"
	TypeFlutter   ProjectType = "flutter"
	TypeProxy     ProjectType = "proxy"
	TypeNonWeb    ProjectType = "non_web"
)

type DetectedProject struct {
	Name        string      `json:"name"`
	Path        string      `json:"path"`
	DocRoot     string      `json:"doc_root"`
	Type        ProjectType `json:"type"`
	TypeLabel   string      `json:"type_label"`
	PHPVersion  string      `json:"php_version"`
	PHPPort     int         `json:"php_port"`
	ProxyPort   int         `json:"proxy_port,omitempty"`
	IsWeb       bool        `json:"is_web"`
	Domain      string      `json:"domain"`
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func hasFilesWithExt(dir string, ext string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(strings.ToLower(e.Name()), ext) {
			return true
		}
	}
	return false
}

func findFirstExistingDir(baseDir string, subDirs ...string) string {
	for _, sub := range subDirs {
		target := filepath.Join(baseDir, sub)
		if dirExists(target) {
			return target
		}
	}
	return baseDir
}

func parseEnvPort(dir string) int {
	envFiles := []string{".env", ".env.local", ".env.development"}
	for _, envFile := range envFiles {
		content, err := os.ReadFile(filepath.Join(dir, envFile))
		if err != nil {
			continue
		}
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 && strings.TrimSpace(parts[0]) == "PORT" {
				val := strings.Trim(strings.TrimSpace(parts[1]), "\"'")
				if p, err := strconv.Atoi(val); err == nil && p > 0 && p < 65536 {
					return p
				}
			}
		}
	}
	return 0
}

// PHPPortForVersion returns the FastCGI port for a given PHP version string
func PHPPortForVersion(version string) int {
	v := strings.TrimSpace(strings.ToLower(version))
	v = strings.TrimPrefix(v, "php")
	v = strings.TrimSpace(v)
	switch {
	case strings.HasPrefix(v, "8.4"):
		return 9084
	case strings.HasPrefix(v, "8.3"):
		return 9083
	case strings.HasPrefix(v, "8.2"):
		return 9082
	case strings.HasPrefix(v, "8.1"):
		return 9081
	default:
		return 9083
	}
}

// NormalizePHPVersion returns a standard string like "8.3", "8.4", etc.
func NormalizePHPVersion(version string) string {
	v := strings.TrimSpace(strings.ToLower(version))
	v = strings.TrimPrefix(v, "php")
	v = strings.TrimSpace(v)
	switch {
	case strings.HasPrefix(v, "8.4"):
		return "8.4"
	case strings.HasPrefix(v, "8.3"):
		return "8.3"
	case strings.HasPrefix(v, "8.2"):
		return "8.2"
	case strings.HasPrefix(v, "8.1"):
		return "8.1"
	default:
		if v != "" {
			return v
		}
		return "8.3"
	}
}

type composerFile struct {
	Require map[string]string `json:"require"`
}

// DetectComposerPHPVersion parses composer.json and extracts the required PHP version
func DetectComposerPHPVersion(composerPath string) string {
	data, err := os.ReadFile(composerPath)
	if err != nil {
		return ""
	}
	var comp composerFile
	if err := json.Unmarshal(data, &comp); err != nil {
		return ""
	}
	phpReq, ok := comp.Require["php"]
	if !ok || phpReq == "" {
		return ""
	}
	return ParsePHPConstraint(phpReq)
}

// ParsePHPConstraint maps a composer PHP constraint like "^8.4", "~8.2.0", "8.1.*" to a normalized version string
func ParsePHPConstraint(constraint string) string {
	c := strings.ToLower(strings.TrimSpace(constraint))
	if c == "" {
		return ""
	}

	// Direct match checks from highest to lowest supported versions
	if strings.Contains(c, "8.4") {
		return "8.4"
	}
	if strings.Contains(c, "8.3") {
		return "8.3"
	}
	if strings.Contains(c, "8.2") {
		return "8.2"
	}
	if strings.Contains(c, "8.1") {
		return "8.1"
	}

	// Legacy constraints
	if strings.Contains(c, "8.0") || strings.Contains(c, "7.4") || strings.Contains(c, "7.3") {
		return "8.1"
	}

	return ""
}

// CleanDomainName generates a clean, RFC-compliant domain name replacing dots and underscores
func CleanDomainName(folderName string, domainSuffix string) string {
	clean := strings.TrimSpace(folderName)
	clean = strings.ToLower(clean)
	// Replace dots, underscores, and spaces with hyphens to ensure valid single-level DNS labels
	clean = strings.ReplaceAll(clean, ".", "-")
	clean = strings.ReplaceAll(clean, "_", "-")
	clean = strings.ReplaceAll(clean, " ", "-")
	clean = strings.Trim(clean, "-")
	if clean == "" {
		clean = "app"
	}
	if !strings.HasPrefix(domainSuffix, ".") {
		domainSuffix = "." + domainSuffix
	}
	return clean + domainSuffix
}

// DetectPlatformCheckPHPVersion inspects vendor/composer/platform_check.php for installed PHP requirements
func DetectPlatformCheckPHPVersion(projectDir string) string {
	candidatePaths := []string{
		filepath.Join(projectDir, "vendor", "composer", "platform_check.php"),
		filepath.Join(projectDir, "backend", "vendor", "composer", "platform_check.php"),
	}
	for _, p := range candidatePaths {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		content := string(data)
		switch {
		case strings.Contains(content, "PHP_VERSION_ID >= 804") || strings.Contains(content, "PHP_VERSION_ID >= 805"):
			return "8.4"
		case strings.Contains(content, "PHP_VERSION_ID >= 803"):
			return "8.3"
		case strings.Contains(content, "PHP_VERSION_ID >= 802"):
			return "8.2"
		case strings.Contains(content, "PHP_VERSION_ID >= 801"):
			return "8.1"
		}
	}
	return ""
}

// DetectProjectPHPVersion detects the required PHP version from:
// 0. vendor/composer/platform_check.php (installed dependencies)
// 1. .php-version file
// 2. .laragal file
// 3. composer.json require.php
// 4. Fallback to defaultPHP
func DetectProjectPHPVersion(projectDir string, defaultPHP string) (string, int) {
	// 0. Check vendor/composer/platform_check.php
	if platformVer := DetectPlatformCheckPHPVersion(projectDir); platformVer != "" {
		return platformVer, PHPPortForVersion(platformVer)
	}

	// 1. Check .php-version file
	phpVerFile := filepath.Join(projectDir, ".php-version")
	if fileExists(phpVerFile) {
		if data, err := os.ReadFile(phpVerFile); err == nil {
			trimmed := strings.TrimSpace(string(data))
			if trimmed != "" {
				norm := NormalizePHPVersion(trimmed)
				return norm, PHPPortForVersion(norm)
			}
		}
	}

	// 2. Check .laragal file
	laragalFile := filepath.Join(projectDir, ".laragal")
	if fileExists(laragalFile) {
		if data, err := os.ReadFile(laragalFile); err == nil {
			trimmed := strings.TrimSpace(string(data))
			if trimmed != "" {
				var lcfg struct {
					PHP string `json:"php"`
				}
				if err := json.Unmarshal(data, &lcfg); err == nil && lcfg.PHP != "" {
					norm := NormalizePHPVersion(lcfg.PHP)
					return norm, PHPPortForVersion(norm)
				}
				norm := NormalizePHPVersion(trimmed)
				return norm, PHPPortForVersion(norm)
			}
		}
	}

	// 3. Check composer.json require.php
	composerPath := filepath.Join(projectDir, "composer.json")
	if fileExists(composerPath) {
		if detected := DetectComposerPHPVersion(composerPath); detected != "" {
			return detected, PHPPortForVersion(detected)
		}
	}

	// 4. Fallback to defaultPHP
	normDefault := NormalizePHPVersion(defaultPHP)
	return normDefault, PHPPortForVersion(normDefault)
}

// detectSingleDir checks whether a specific directory contains web artifacts
func detectSingleDir(cleanDir string, defaultPHP string, domainSuffix string) DetectedProject {
	name := filepath.Base(cleanDir)

	domain := CleanDomainName(name, domainSuffix)

	result := DetectedProject{
		Name:       name,
		Path:       cleanDir,
		DocRoot:    cleanDir,
		Type:       TypeNonWeb,
		TypeLabel:  "غير ويب / ملفات",
		PHPVersion: "",
		PHPPort:    0,
		IsWeb:      false,
		Domain:     "",
	}

	// 0a. Ignore LaraGal desktop suite repository itself
	if strings.EqualFold(name, "laragal") || fileExists(filepath.Join(cleanDir, "wails.json")) {
		return result
	}

	// 0b. Ignore non-web skills, agent rule repositories, and documentation suites
	if fileExists(filepath.Join(cleanDir, "SKILL.md")) ||
		fileExists(filepath.Join(cleanDir, ".claude-plugin")) ||
		(dirExists(filepath.Join(cleanDir, ".agents")) && !fileExists(filepath.Join(cleanDir, "artisan")) && !fileExists(filepath.Join(cleanDir, "public", "index.php")) && !fileExists(filepath.Join(cleanDir, "index.php"))) {
		if !fileExists(filepath.Join(cleanDir, "artisan")) && !fileExists(filepath.Join(cleanDir, "public", "index.php")) && !fileExists(filepath.Join(cleanDir, "index.php")) {
			return result
		}
	}

	// 1. Check for Laravel (artisan or composer.json with laravel/framework)
	if fileExists(filepath.Join(cleanDir, "artisan")) {
		phpVer, phpPort := DetectProjectPHPVersion(cleanDir, defaultPHP)
		result.Type = TypeLaravel
		result.TypeLabel = "Laravel"
		result.IsWeb = true
		result.PHPVersion = phpVer
		result.PHPPort = phpPort
		result.Domain = domain
		result.DocRoot = findFirstExistingDir(cleanDir, "public")
		return result
	}

	// 2. Check for WordPress
	if fileExists(filepath.Join(cleanDir, "wp-config.php")) ||
		fileExists(filepath.Join(cleanDir, "wp-load.php")) ||
		fileExists(filepath.Join(cleanDir, "wp-login.php")) ||
		fileExists(filepath.Join(cleanDir, "wp-settings.php")) ||
		dirExists(filepath.Join(cleanDir, "wp-content")) {
		phpVer, phpPort := DetectProjectPHPVersion(cleanDir, defaultPHP)
		result.Type = TypeWordPress
		result.TypeLabel = "WordPress"
		result.IsWeb = true
		result.PHPVersion = phpVer
		result.PHPPort = phpPort
		result.Domain = domain
		return result
	}

	// 3. Check composer.json (Laravel, Symfony, WordPress Bedrock, or any PHP composer project)
	composerPath := filepath.Join(cleanDir, "composer.json")
	if fileExists(composerPath) {
		phpVer, phpPort := DetectProjectPHPVersion(cleanDir, defaultPHP)
		data, _ := os.ReadFile(composerPath)
		content := strings.ToLower(string(data))
		if strings.Contains(content, "laravel/framework") || strings.Contains(content, "laravel/lumen-framework") {
			result.Type = TypeLaravel
			result.TypeLabel = "Laravel"
			result.IsWeb = true
			result.PHPVersion = phpVer
			result.PHPPort = phpPort
			result.Domain = domain
			result.DocRoot = findFirstExistingDir(cleanDir, "public")
			return result
		}
		if strings.Contains(content, "roots/bedrock") || strings.Contains(content, "johnpbloch/wordpress") {
			result.Type = TypeWordPress
			result.TypeLabel = "WordPress"
			result.IsWeb = true
			result.PHPVersion = phpVer
			result.PHPPort = phpPort
			result.Domain = domain
			result.DocRoot = findFirstExistingDir(cleanDir, "web", "public")
			return result
		}

		// Any other project with composer.json is a PHP web application
		result.Type = TypePHP
		result.TypeLabel = "PHP"
		result.IsWeb = true
		result.PHPVersion = phpVer
		result.PHPPort = phpPort
		result.Domain = domain
		result.DocRoot = findFirstExistingDir(cleanDir, "public", "web", "public_html", "htdocs", "webroot")
		return result
	}

	// 4. Check for PHP entry points in common public roots or project root
	phpEntryCandidates := []string{
		filepath.Join(cleanDir, "public", "index.php"),
		filepath.Join(cleanDir, "web", "index.php"),
		filepath.Join(cleanDir, "public_html", "index.php"),
		filepath.Join(cleanDir, "htdocs", "index.php"),
		filepath.Join(cleanDir, "webroot", "index.php"),
		filepath.Join(cleanDir, "index.php"),
	}
	for _, entryPath := range phpEntryCandidates {
		if fileExists(entryPath) {
			phpVer, phpPort := DetectProjectPHPVersion(cleanDir, defaultPHP)
			result.Type = TypePHP
			result.TypeLabel = "PHP"
			result.IsWeb = true
			result.PHPVersion = phpVer
			result.PHPPort = phpPort
			result.Domain = domain
			result.DocRoot = filepath.Dir(entryPath)
			return result
		}
	}

	// 4b. Check for Go Web Applications (go.mod with web/, cmd/server, templates/, etc.)
	if fileExists(filepath.Join(cleanDir, "go.mod")) {
		hasGoWeb := dirExists(filepath.Join(cleanDir, "web")) ||
			dirExists(filepath.Join(cleanDir, "cmd", "server")) ||
			dirExists(filepath.Join(cleanDir, "templates")) ||
			dirExists(filepath.Join(cleanDir, "web", "templates")) ||
			dirExists(filepath.Join(cleanDir, "web", "static"))
		if hasGoWeb {
			result.Type = TypeGo
			result.TypeLabel = "Go / ويب"
			result.IsWeb = true
			result.Domain = name + domainSuffix
			result.Path = cleanDir
			result.DocRoot = findFirstExistingDir(cleanDir, "web", "public", "dist", "static")
			result.PHPPort = 0
			result.ProxyPort = 8080
			return result
		}
	}

	// 4c. Check for Rust Web Applications (Cargo.toml)
	if fileExists(filepath.Join(cleanDir, "Cargo.toml")) {
		hasRustWeb := dirExists(filepath.Join(cleanDir, "src")) ||
			fileExists(filepath.Join(cleanDir, "src", "main.rs"))
		if hasRustWeb {
			result.Type = TypeRust
			result.TypeLabel = "Rust / ويب"
			result.IsWeb = true
			result.Domain = name + domainSuffix
			result.Path = cleanDir
			result.DocRoot = findFirstExistingDir(cleanDir, "static", "public", "dist", "")
			result.PHPPort = 0
			result.ProxyPort = 8080
			return result
		}
	}

	// 4d. Check for .NET / C# Web Applications (*.csproj with Web SDK or wwwroot)
	hasDotNetWeb := false
	if dirExists(filepath.Join(cleanDir, "wwwroot")) {
		hasDotNetWeb = true
	} else if hasFilesWithExt(cleanDir, ".csproj") || hasFilesWithExt(cleanDir, ".fsproj") {
		if entries, err := os.ReadDir(cleanDir); err == nil {
			for _, e := range entries {
				if strings.HasSuffix(strings.ToLower(e.Name()), ".csproj") || strings.HasSuffix(strings.ToLower(e.Name()), ".fsproj") {
					if data, readErr := os.ReadFile(filepath.Join(cleanDir, e.Name())); readErr == nil {
						content := strings.ToLower(string(data))
						if strings.Contains(content, "microsoft.net.sdk.web") || strings.Contains(content, "aspnetcore") {
							hasDotNetWeb = true
							break
						}
					}
				}
			}
		}
	}
	if hasDotNetWeb {
		result.Type = TypeDotNet
		result.TypeLabel = ".NET / C#"
		result.IsWeb = true
		result.Domain = domain
		result.Path = cleanDir
		result.DocRoot = findFirstExistingDir(cleanDir, "wwwroot", "public", "")
		result.PHPPort = 0
		result.ProxyPort = 5000
		return result
	}

	// 4e. Check for Python Web Applications (manage.py, wsgi.py, asgi.py, requirements.txt, pyproject.toml)
	isPyWeb := fileExists(filepath.Join(cleanDir, "manage.py")) ||
		fileExists(filepath.Join(cleanDir, "wsgi.py")) ||
		fileExists(filepath.Join(cleanDir, "asgi.py")) ||
		dirExists(filepath.Join(cleanDir, ".streamlit"))
	pyPort := 8000
	pyLabel := "Python / Web"
	if dirExists(filepath.Join(cleanDir, ".streamlit")) {
		isPyWeb = true
		pyPort = 8501
		pyLabel = "Streamlit"
	}
	if !isPyWeb && (fileExists(filepath.Join(cleanDir, "requirements.txt")) || fileExists(filepath.Join(cleanDir, "pyproject.toml"))) {
		reqContent := ""
		if data, err := os.ReadFile(filepath.Join(cleanDir, "requirements.txt")); err == nil {
			reqContent += strings.ToLower(string(data))
		}
		if data, err := os.ReadFile(filepath.Join(cleanDir, "pyproject.toml")); err == nil {
			reqContent += strings.ToLower(string(data))
		}
		if strings.Contains(reqContent, "streamlit") {
			isPyWeb = true
			pyPort = 8501
			pyLabel = "Streamlit"
		} else if strings.Contains(reqContent, "fastapi") {
			isPyWeb = true
			pyPort = 8000
			pyLabel = "FastAPI"
		} else if strings.Contains(reqContent, "flask") {
			isPyWeb = true
			pyPort = 5000
			pyLabel = "Flask"
		} else if strings.Contains(reqContent, "django") || strings.Contains(reqContent, "uvicorn") || strings.Contains(reqContent, "tornado") || strings.Contains(reqContent, "sanic") {
			isPyWeb = true
			pyPort = 8000
			pyLabel = "Python / Web"
		}
	}
	if isPyWeb {
		result.Type = TypePython
		result.TypeLabel = pyLabel
		result.IsWeb = true
		result.Domain = domain
		result.Path = cleanDir
		result.DocRoot = findFirstExistingDir(cleanDir, "static", "public", "")
		result.PHPPort = 0
		result.ProxyPort = pyPort
		return result
	}

	// 4f. Check for Ruby on Rails / Sinatra (Gemfile, config.ru)
	if fileExists(filepath.Join(cleanDir, "Gemfile")) || fileExists(filepath.Join(cleanDir, "config.ru")) {
		result.Type = TypeRuby
		result.TypeLabel = "Ruby on Rails"
		result.IsWeb = true
		result.Domain = domain
		result.Path = cleanDir
		result.DocRoot = findFirstExistingDir(cleanDir, "public", "")
		result.PHPPort = 0
		result.ProxyPort = 3000
		return result
	}

	// 4g. Check for Java / Kotlin (pom.xml, build.gradle, build.gradle.kts)
	if fileExists(filepath.Join(cleanDir, "pom.xml")) || fileExists(filepath.Join(cleanDir, "build.gradle")) || fileExists(filepath.Join(cleanDir, "build.gradle.kts")) {
		result.Type = TypeJava
		result.TypeLabel = "Java / Spring"
		result.IsWeb = true
		result.Domain = domain
		result.Path = cleanDir
		result.DocRoot = findFirstExistingDir(cleanDir, "src/main/resources/static", "src/main/resources/public", "public", "")
		result.PHPPort = 0
		result.ProxyPort = 8080
		return result
	}

	// 4h. Check for Flutter Web (pubspec.yaml with flutter:)
	pubspecPath := filepath.Join(cleanDir, "pubspec.yaml")
	if fileExists(pubspecPath) {
		data, _ := os.ReadFile(pubspecPath)
		if strings.Contains(string(data), "flutter:") {
			hasWebTarget := dirExists(filepath.Join(cleanDir, "build", "web")) || dirExists(filepath.Join(cleanDir, "web"))
			if hasWebTarget {
				result.Type = TypeFlutter
				result.TypeLabel = "Flutter Web"
				result.IsWeb = true
				result.Domain = domain
				result.Path = cleanDir
				result.DocRoot = findFirstExistingDir(cleanDir, "build/web", "web", "")
				result.PHPPort = 0
				return result
			}
		}
	}

	// 5. Check for Node.js web applications (package.json or frontend config files) - CHECKED BEFORE GENERIC STATIC
	frontendConfigFiles := []string{
		"vite.config.js", "vite.config.ts", "vite.config.mjs",
		"next.config.js", "next.config.ts", "next.config.mjs",
		"nuxt.config.js", "nuxt.config.ts",
		"astro.config.mjs", "astro.config.ts",
		"svelte.config.js", "svelte.config.ts",
		"webpack.config.js", "angular.json",
	}
	hasFrontendConfig := false
	for _, cfg := range frontendConfigFiles {
		if fileExists(filepath.Join(cleanDir, cfg)) {
			hasFrontendConfig = true
			break
		}
	}

	pkgPath := filepath.Join(cleanDir, "package.json")
	if hasFrontendConfig || fileExists(pkgPath) {
		pkgContent := ""
		if data, err := os.ReadFile(pkgPath); err == nil {
			pkgContent = strings.ToLower(string(data))
		}
		isWebNode := hasFrontendConfig ||
			strings.Contains(pkgContent, "\"vite\"") ||
			strings.Contains(pkgContent, "\"next\"") ||
			strings.Contains(pkgContent, "\"nuxt\"") ||
			strings.Contains(pkgContent, "\"react\"") ||
			strings.Contains(pkgContent, "\"vue\"") ||
			strings.Contains(pkgContent, "\"astro\"") ||
			strings.Contains(pkgContent, "\"svelte\"") ||
			strings.Contains(pkgContent, "\"express\"") ||
			strings.Contains(pkgContent, "\"hono\"") ||
			strings.Contains(pkgContent, "\"fastify\"") ||
			strings.Contains(pkgContent, "\"remix\"") ||
			strings.Contains(pkgContent, "\"@angular/core\"") ||
			strings.Contains(pkgContent, "\"gatsby\"") ||
			(strings.Contains(pkgContent, "\"dev\":") && (dirExists(filepath.Join(cleanDir, "src")) || dirExists(filepath.Join(cleanDir, "public"))))

		if isWebNode {
			result.Type = TypeNode
			result.IsWeb = true
			result.Domain = domain
			result.Path = cleanDir
			result.PHPPort = 0

			// Next.js App
			if fileExists(filepath.Join(cleanDir, "next.config.js")) ||
				fileExists(filepath.Join(cleanDir, "next.config.ts")) ||
				fileExists(filepath.Join(cleanDir, "next.config.mjs")) ||
				strings.Contains(pkgContent, "\"next\"") {
				result.TypeLabel = "Next.js"
				result.ProxyPort = 3000
				if envPort := parseEnvPort(cleanDir); envPort > 0 {
					result.ProxyPort = envPort
				}
				result.DocRoot = findFirstExistingDir(cleanDir, "public", "")
				return result
			}

			// Nuxt App
			if fileExists(filepath.Join(cleanDir, "nuxt.config.js")) ||
				fileExists(filepath.Join(cleanDir, "nuxt.config.ts")) ||
				strings.Contains(pkgContent, "\"nuxt\"") {
				result.TypeLabel = "Nuxt"
				result.ProxyPort = 3000
				if envPort := parseEnvPort(cleanDir); envPort > 0 {
					result.ProxyPort = envPort
				}
				result.DocRoot = findFirstExistingDir(cleanDir, "public", "")
				return result
			}

			// Built Static SPA with index.html in dist, out, or build
			for _, buildSub := range []string{"dist", "out", "build"} {
				if fileExists(filepath.Join(cleanDir, buildSub, "index.html")) {
					result.Type = TypeStatic
					result.TypeLabel = "HTML / ثابت"
					result.DocRoot = filepath.Join(cleanDir, buildSub)
					result.ProxyPort = 0
					return result
				}
			}

			// Vite / Frontend Dev Server
			if fileExists(filepath.Join(cleanDir, "vite.config.js")) ||
				fileExists(filepath.Join(cleanDir, "vite.config.ts")) ||
				fileExists(filepath.Join(cleanDir, "vite.config.mjs")) ||
				strings.Contains(pkgContent, "\"vite\"") {
				result.TypeLabel = "Vite"
				result.ProxyPort = 5173
				if envPort := parseEnvPort(cleanDir); envPort > 0 {
					result.ProxyPort = envPort
				}
				result.DocRoot = findFirstExistingDir(cleanDir, "public", "")
				return result
			}

			// Node API / Backend (Express, Fastify, Nest, Hono)
			result.TypeLabel = "Node.js"
			result.ProxyPort = 3000
			if envPort := parseEnvPort(cleanDir); envPort > 0 {
				result.ProxyPort = envPort
			}
			result.DocRoot = findFirstExistingDir(cleanDir, "public", "")
			return result
		}
	}

	// 6. Check for offline scripts with lone index.html that are NOT web sites (e.g. fix_hidden_charecters)
	isOfflineScript := (fileExists(filepath.Join(cleanDir, "app.py")) || fileExists(filepath.Join(cleanDir, "cli.py")) || fileExists(filepath.Join(cleanDir, "cleaner.js"))) &&
		!fileExists(filepath.Join(cleanDir, "manage.py")) &&
		!fileExists(filepath.Join(cleanDir, "wsgi.py")) &&
		!fileExists(filepath.Join(cleanDir, "asgi.py"))
	if isOfflineScript {
		pyWeb := false
		if reqData, err := os.ReadFile(filepath.Join(cleanDir, "requirements.txt")); err == nil {
			c := strings.ToLower(string(reqData))
			pyWeb = strings.Contains(c, "fastapi") || strings.Contains(c, "django") || strings.Contains(c, "flask") || strings.Contains(c, "streamlit") || strings.Contains(c, "uvicorn")
		}
		if !pyWeb {
			if pyData, err := os.ReadFile(filepath.Join(cleanDir, "app.py")); err == nil {
				c := strings.ToLower(string(pyData))
				pyWeb = strings.Contains(c, "fastapi") || strings.Contains(c, "flask") || strings.Contains(c, "streamlit") || strings.Contains(c, "django")
			}
		}
		if !pyWeb {
			return result
		}
	}

	// 7. Check for Static HTML sites (index.html, index.htm, default.html) in dist/, public/, build/, or root
	htmlDirCandidates := []string{"dist", "public", "build", "out", "web", ""}
	htmlFileNames := []string{"index.html", "index.htm", "default.html", "default.htm"}
	for _, sub := range htmlDirCandidates {
		targetDir := cleanDir
		if sub != "" {
			targetDir = filepath.Join(cleanDir, sub)
		}
		for _, fName := range htmlFileNames {
			if fileExists(filepath.Join(targetDir, fName)) {
				result.Type = TypeStatic
				result.TypeLabel = "HTML / ثابت"
				result.IsWeb = true
				result.Domain = domain
				result.DocRoot = targetDir
				result.PHPPort = 0
				return result
			}
		}
	}

	// 8. Check if folder contains any .php files anywhere in top-level
	entries, err := os.ReadDir(cleanDir)
	if err == nil {
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(strings.ToLower(e.Name()), ".php") {
				phpVer, phpPort := DetectProjectPHPVersion(cleanDir, defaultPHP)
				result.Type = TypePHP
				result.TypeLabel = "PHP"
				result.IsWeb = true
				result.PHPVersion = phpVer
				result.PHPPort = phpPort
				result.Domain = domain
				return result
			}
		}
	}

	// 9. Otherwise: directory does not contain direct web artifacts
	return result
}

// DetectProject inspects a folder to determine whether it is a web application or contains a web section
func DetectProject(projectDir string, defaultPHP string, domainSuffix string) DetectedProject {
	cleanDir := filepath.Clean(projectDir)
	name := filepath.Base(cleanDir)
	domain := CleanDomainName(name, domainSuffix)

	// 0. Ignore LaraGal desktop suite repository itself
	if strings.EqualFold(name, "laragal") || fileExists(filepath.Join(cleanDir, "wails.json")) {
		return DetectedProject{
			Name:       name,
			Path:       cleanDir,
			DocRoot:    cleanDir,
			Type:       TypeNonWeb,
			TypeLabel:  "غير ويب / ملفات",
			PHPVersion: "",
			PHPPort:    0,
			IsWeb:      false,
			Domain:     "",
		}
	}

	// 1. Direct root check
	rootResult := detectSingleDir(cleanDir, defaultPHP, domainSuffix)
	if rootResult.IsWeb {
		return rootResult
	}

	// 2. Subdirectory check for multi-tier projects and monorepos
	webSubDirs := []string{
		"backend",
		"services",
		"server",
		"api",
		"web",
		"frontend",
		"client",
		"ui",
		"site",
		"apps",
		"admin",
		"app",
	}

	for _, sub := range webSubDirs {
		subPath := filepath.Join(cleanDir, sub)
		if !dirExists(subPath) {
			continue
		}

		subResult := detectSingleDir(subPath, defaultPHP, domainSuffix)
		if subResult.IsWeb {
			return DetectedProject{
				Name:        name,
				Path:        cleanDir,
				DocRoot:     subResult.DocRoot,
				Type:        subResult.Type,
				TypeLabel:   fmt.Sprintf("%s (%s)", subResult.TypeLabel, sub),
				PHPVersion:  subResult.PHPVersion,
				PHPPort:     subResult.PHPPort,
				ProxyPort:   subResult.ProxyPort,
				IsWeb:       true,
				Domain:      domain,
			}
		}

		// Check one level deeper inside container folders like apps/ or services/
		if sub == "apps" || sub == "services" {
			subEntries, err := os.ReadDir(subPath)
			if err == nil {
				for _, se := range subEntries {
					if se.IsDir() {
						nestedPath := filepath.Join(subPath, se.Name())
						nestedResult := detectSingleDir(nestedPath, defaultPHP, domainSuffix)
						if nestedResult.IsWeb {
							return DetectedProject{
								Name:        name,
								Path:        cleanDir,
								DocRoot:     nestedResult.DocRoot,
								Type:        nestedResult.Type,
								TypeLabel:   fmt.Sprintf("%s (%s/%s)", nestedResult.TypeLabel, sub, se.Name()),
								PHPVersion:  nestedResult.PHPVersion,
								PHPPort:     nestedResult.PHPPort,
								ProxyPort:   nestedResult.ProxyPort,
								IsWeb:       true,
								Domain:      domain,
							}
						}
					}
				}
			}
		}
	}

	// 3. Genuine non-web folder
	return rootResult
}
