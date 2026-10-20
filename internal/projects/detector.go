package projects

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ProjectType string

const (
	TypeLaravel   ProjectType = "laravel"
	TypeWordPress ProjectType = "wordpress"
	TypePHP       ProjectType = "php"
	TypeStatic    ProjectType = "static"
	TypeNode      ProjectType = "node"
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

// detectSingleDir checks whether a specific directory contains web artifacts
func detectSingleDir(cleanDir string, defaultPHP string, domainSuffix string) DetectedProject {
	name := filepath.Base(cleanDir)

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

	// 1. Check for Laravel (artisan or composer.json with laravel/framework)
	if fileExists(filepath.Join(cleanDir, "artisan")) {
		result.Type = TypeLaravel
		result.TypeLabel = "Laravel"
		result.IsWeb = true
		result.PHPVersion = defaultPHP
		result.PHPPort = 9083
		result.Domain = name + domainSuffix
		result.DocRoot = findFirstExistingDir(cleanDir, "public")
		return result
	}

	// 2. Check for WordPress
	if fileExists(filepath.Join(cleanDir, "wp-config.php")) ||
		fileExists(filepath.Join(cleanDir, "wp-load.php")) ||
		fileExists(filepath.Join(cleanDir, "wp-login.php")) ||
		fileExists(filepath.Join(cleanDir, "wp-settings.php")) ||
		dirExists(filepath.Join(cleanDir, "wp-content")) {
		result.Type = TypeWordPress
		result.TypeLabel = "WordPress"
		result.IsWeb = true
		result.PHPVersion = defaultPHP
		result.PHPPort = 9083
		result.Domain = name + domainSuffix
		return result
	}

	// 3. Check composer.json (Laravel, Symfony, WordPress Bedrock, or any PHP composer project)
	composerPath := filepath.Join(cleanDir, "composer.json")
	if fileExists(composerPath) {
		data, _ := os.ReadFile(composerPath)
		content := strings.ToLower(string(data))
		if strings.Contains(content, "laravel/framework") || strings.Contains(content, "laravel/lumen-framework") {
			result.Type = TypeLaravel
			result.TypeLabel = "Laravel"
			result.IsWeb = true
			result.PHPVersion = defaultPHP
			result.PHPPort = 9083
			result.Domain = name + domainSuffix
			result.DocRoot = findFirstExistingDir(cleanDir, "public")
			return result
		}
		if strings.Contains(content, "roots/bedrock") || strings.Contains(content, "johnpbloch/wordpress") {
			result.Type = TypeWordPress
			result.TypeLabel = "WordPress"
			result.IsWeb = true
			result.PHPVersion = defaultPHP
			result.PHPPort = 9083
			result.Domain = name + domainSuffix
			result.DocRoot = findFirstExistingDir(cleanDir, "web", "public")
			return result
		}

		// Any other project with composer.json is a PHP web application
		result.Type = TypePHP
		result.TypeLabel = "PHP"
		result.IsWeb = true
		result.PHPVersion = defaultPHP
		result.PHPPort = 9083
		result.Domain = name + domainSuffix
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
			result.Type = TypePHP
			result.TypeLabel = "PHP"
			result.IsWeb = true
			result.PHPVersion = defaultPHP
			result.PHPPort = 9083
			result.Domain = name + domainSuffix
			result.DocRoot = filepath.Dir(entryPath)
			return result
		}
	}

	// 5. Check for Static HTML sites (index.html, index.htm, default.html) in dist/, public/, build/, or root
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
				result.Domain = name + domainSuffix
				result.DocRoot = targetDir
				result.PHPPort = 0
				return result
			}
		}
	}

	// 5b. Check for web template directories with HTML templates (e.g. templates/*.html, views/*.html, pages/*.html)
	for _, tplSub := range []string{"templates", "views", "pages", "templates/html", "dist", "static"} {
		tplDir := filepath.Join(cleanDir, tplSub)
		if hasFilesWithExt(tplDir, ".html") || hasFilesWithExt(tplDir, ".htm") {
			result.Type = TypeStatic
			result.TypeLabel = "HTML / ويب"
			result.IsWeb = true
			result.Domain = name + domainSuffix
			result.DocRoot = cleanDir
			result.PHPPort = 0
			return result
		}
	}

	// 5c. Check if directory contains a static/ or assets/ folder alongside templates/ or locales/ (like Go/Python web structures)
	if (dirExists(filepath.Join(cleanDir, "static")) || dirExists(filepath.Join(cleanDir, "assets"))) &&
		(dirExists(filepath.Join(cleanDir, "templates")) || dirExists(filepath.Join(cleanDir, "views")) || dirExists(filepath.Join(cleanDir, "locales"))) {
		result.Type = TypeStatic
		result.TypeLabel = "HTML / ويب"
		result.IsWeb = true
		result.Domain = name + domainSuffix
		result.DocRoot = cleanDir
		result.PHPPort = 0
		return result
	}

	// 5d. Check for Python Web applications (Django, Flask, FastAPI)
	if fileExists(filepath.Join(cleanDir, "manage.py")) ||
		fileExists(filepath.Join(cleanDir, "wsgi.py")) ||
		fileExists(filepath.Join(cleanDir, "asgi.py")) {
		result.Type = TypeStatic
		result.TypeLabel = "Python / Web"
		result.IsWeb = true
		result.Domain = name + domainSuffix
		result.DocRoot = findFirstExistingDir(cleanDir, "static", "public", "")
		result.PHPPort = 0
		return result
	}

	// 6. Check for Node.js web applications (package.json or frontend config files)
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
		isWebNode := hasFrontendConfig
		if !isWebNode && fileExists(pkgPath) {
			data, _ := os.ReadFile(pkgPath)
			content := strings.ToLower(string(data))
			isWebNode = strings.Contains(content, "vite") ||
				strings.Contains(content, "next") ||
				strings.Contains(content, "nuxt") ||
				strings.Contains(content, "react") ||
				strings.Contains(content, "vue") ||
				strings.Contains(content, "astro") ||
				strings.Contains(content, "svelte") ||
				strings.Contains(content, "express") ||
				strings.Contains(content, "hono") ||
				strings.Contains(content, "fastify") ||
				strings.Contains(content, "remix") ||
				strings.Contains(content, "angular") ||
				strings.Contains(content, "gatsby") ||
				strings.Contains(content, "webpack") ||
				strings.Contains(content, "tailwind") ||
				strings.Contains(content, "\"dev\":") ||
				strings.Contains(content, "\"start\":") ||
				strings.Contains(content, "\"build\":")
		}

		if isWebNode {
			result.Type = TypeNode
			result.TypeLabel = "Node.js"
			result.IsWeb = true
			result.Domain = name + domainSuffix
			result.DocRoot = findFirstExistingDir(cleanDir, "dist", "build", "public", "out")
			result.PHPPort = 0
			return result
		}
	}

	// 7. Check if folder contains any .php files anywhere in top-level
	entries, err := os.ReadDir(cleanDir)
	if err == nil {
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(strings.ToLower(e.Name()), ".php") {
				result.Type = TypePHP
				result.TypeLabel = "PHP"
				result.IsWeb = true
				result.PHPVersion = defaultPHP
				result.PHPPort = 9083
				result.Domain = name + domainSuffix
				return result
			}
		}
	}

	// 8. Check if folder contains any .html / .htm files in top-level
	if err == nil {
		for _, e := range entries {
			if !e.IsDir() {
				lowerName := strings.ToLower(e.Name())
				if strings.HasSuffix(lowerName, ".html") || strings.HasSuffix(lowerName, ".htm") {
					result.Type = TypeStatic
					result.TypeLabel = "HTML / ثابت"
					result.IsWeb = true
					result.Domain = name + domainSuffix
					result.PHPPort = 0
					return result
				}
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

	// 1. Direct root check
	rootResult := detectSingleDir(cleanDir, defaultPHP, domainSuffix)
	if rootResult.IsWeb {
		return rootResult
	}

	// 2. Subdirectory check for multi-tier projects and monorepos
	webSubDirs := []string{
		"backend",
		"web",
		"server",
		"api",
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
				IsWeb:       true,
				Domain:      name + domainSuffix,
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
								IsWeb:       true,
								Domain:      name + domainSuffix,
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
