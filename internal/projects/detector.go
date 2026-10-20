package projects

import (
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

// DetectProject inspects a folder to determine whether it is a web application and its specific stack
func DetectProject(projectDir string, defaultPHP string, domainSuffix string) DetectedProject {
	cleanDir := filepath.Clean(projectDir)
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
		if dirExists(filepath.Join(cleanDir, "public")) {
			result.DocRoot = filepath.Join(cleanDir, "public")
		}
		return result
	}

	// 2. Check for WordPress
	if fileExists(filepath.Join(cleanDir, "wp-config.php")) || dirExists(filepath.Join(cleanDir, "wp-content")) {
		result.Type = TypeWordPress
		result.TypeLabel = "WordPress"
		result.IsWeb = true
		result.PHPVersion = defaultPHP
		result.PHPPort = 9083
		result.Domain = name + domainSuffix
		return result
	}

	// 3. Check composer.json for PHP frameworks
	composerPath := filepath.Join(cleanDir, "composer.json")
	if fileExists(composerPath) {
		data, _ := os.ReadFile(composerPath)
		content := strings.ToLower(string(data))
		if strings.Contains(content, "laravel/framework") {
			result.Type = TypeLaravel
			result.TypeLabel = "Laravel"
			result.IsWeb = true
			result.PHPVersion = defaultPHP
			result.PHPPort = 9083
			result.Domain = name + domainSuffix
			if dirExists(filepath.Join(cleanDir, "public")) {
				result.DocRoot = filepath.Join(cleanDir, "public")
			}
			return result
		}
	}

	// 4. Check for PHP entry points (index.php in public/ or root)
	publicIndexPHP := filepath.Join(cleanDir, "public", "index.php")
	rootIndexPHP := filepath.Join(cleanDir, "index.php")
	if fileExists(publicIndexPHP) {
		result.Type = TypePHP
		result.TypeLabel = "PHP"
		result.IsWeb = true
		result.PHPVersion = defaultPHP
		result.PHPPort = 9083
		result.Domain = name + domainSuffix
		result.DocRoot = filepath.Join(cleanDir, "public")
		return result
	}
	if fileExists(rootIndexPHP) {
		result.Type = TypePHP
		result.TypeLabel = "PHP"
		result.IsWeb = true
		result.PHPVersion = defaultPHP
		result.PHPPort = 9083
		result.Domain = name + domainSuffix
		return result
	}

	// 5. Check for Static HTML sites (index.html in root, dist/, or public/)
	distIndexHTML := filepath.Join(cleanDir, "dist", "index.html")
	publicIndexHTML := filepath.Join(cleanDir, "public", "index.html")
	rootIndexHTML := filepath.Join(cleanDir, "index.html")

	if fileExists(distIndexHTML) {
		result.Type = TypeStatic
		result.TypeLabel = "HTML / ثابت"
		result.IsWeb = true
		result.Domain = name + domainSuffix
		result.DocRoot = filepath.Join(cleanDir, "dist")
		return result
	}
	if fileExists(publicIndexHTML) {
		result.Type = TypeStatic
		result.TypeLabel = "HTML / ثابت"
		result.IsWeb = true
		result.Domain = name + domainSuffix
		result.DocRoot = filepath.Join(cleanDir, "public")
		return result
	}
	if fileExists(rootIndexHTML) {
		result.Type = TypeStatic
		result.TypeLabel = "HTML / ثابت"
		result.IsWeb = true
		result.Domain = name + domainSuffix
		return result
	}

	// 6. Check for Node.js web applications (package.json with build scripts or web frameworks)
	pkgPath := filepath.Join(cleanDir, "package.json")
	if fileExists(pkgPath) {
		data, _ := os.ReadFile(pkgPath)
		content := strings.ToLower(string(data))
		isWebNode := strings.Contains(content, "vite") ||
			strings.Contains(content, "next") ||
			strings.Contains(content, "nuxt") ||
			strings.Contains(content, "react") ||
			strings.Contains(content, "vue") ||
			strings.Contains(content, "astro") ||
			strings.Contains(content, "svelte") ||
			strings.Contains(content, "express")

		if isWebNode {
			result.Type = TypeNode
			result.TypeLabel = "Node.js"
			result.IsWeb = true
			result.Domain = name + domainSuffix
			if dirExists(filepath.Join(cleanDir, "dist")) {
				result.DocRoot = filepath.Join(cleanDir, "dist")
			} else if dirExists(filepath.Join(cleanDir, "build")) {
				result.DocRoot = filepath.Join(cleanDir, "build")
			}
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

	// 8. Otherwise: folder is NOT a website (e.g. skills, docs, markdown, CLI tools)
	return result
}
