package projects

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/laragal/laragal/internal/winutil"
)

// DevServerInfo holds the detected frontend scripts and tooling for a project
type DevServerInfo struct {
	HasPackageJSON bool   `json:"has_package_json"`
	ScriptCommand  string `json:"script_command"`  // e.g. "dev", "start", "serve"
	PackageManager string `json:"package_manager"` // "pnpm", "yarn", "npm"
	Framework      string `json:"framework"`       // "Vite", "Next.js", "Nuxt", "Astro", "Tailwind", "Node"
	DefaultPort    int    `json:"default_port"`
	IsRunning      bool   `json:"is_running"`
	PID            int    `json:"pid,omitempty"`
}

// NodeEnvInfo reports system Node.js and package manager versions and paths
type NodeEnvInfo struct {
	Available      bool   `json:"available"`
	NodeVersion    string `json:"node_version"`
	NpmVersion     string `json:"npm_version"`
	NodePath       string `json:"node_path"`
	NpmPath        string `json:"npm_path"`
	PackageManager string `json:"default_pm"`
}

type packageJSONModel struct {
	Scripts         map[string]string `json:"scripts"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

// DetectDevServer inspects project directory for package.json and suitable dev server scripts
func DetectDevServer(projectDir string) DevServerInfo {
	pkgPath := filepath.Join(projectDir, "package.json")
	data, err := os.ReadFile(pkgPath)
	if err != nil {
		return DevServerInfo{HasPackageJSON: false}
	}

	var pkg packageJSONModel
	if err := json.Unmarshal(data, &pkg); err != nil {
		return DevServerInfo{HasPackageJSON: true}
	}

	info := DevServerInfo{
		HasPackageJSON: true,
		PackageManager: "npm",
		DefaultPort:    5173,
		Framework:      "Node.js",
	}

	// Detect lockfiles to determine preferred package manager
	if _, err := os.Stat(filepath.Join(projectDir, "pnpm-lock.yaml")); err == nil {
		info.PackageManager = "pnpm"
	} else if _, err := os.Stat(filepath.Join(projectDir, "yarn.lock")); err == nil {
		info.PackageManager = "yarn"
	} else if _, err := os.Stat(filepath.Join(projectDir, "bun.lockb")); err == nil {
		info.PackageManager = "bun"
	}

	// Detect Framework
	allDeps := make(map[string]bool)
	for k := range pkg.Dependencies {
		allDeps[strings.ToLower(k)] = true
	}
	for k := range pkg.DevDependencies {
		allDeps[strings.ToLower(k)] = true
	}

	switch {
	case allDeps["vite"]:
		info.Framework = "Vite"
		info.DefaultPort = 5173
	case allDeps["next"]:
		info.Framework = "Next.js"
		info.DefaultPort = 3000
	case allDeps["nuxt"]:
		info.Framework = "Nuxt"
		info.DefaultPort = 3000
	case allDeps["astro"]:
		info.Framework = "Astro"
		info.DefaultPort = 4321
	case allDeps["@angular/core"]:
		info.Framework = "Angular"
		info.DefaultPort = 4200
	case allDeps["react-scripts"]:
		info.Framework = "React"
		info.DefaultPort = 3000
	case allDeps["tailwindcss"]:
		info.Framework = "Tailwind"
		info.DefaultPort = 3000
	}

	// Pick script command: prefer dev > start > serve > watch
	candidates := []string{"dev", "start", "serve", "watch"}
	for _, c := range candidates {
		if _, ok := pkg.Scripts[c]; ok {
			info.ScriptCommand = c
			break
		}
	}

	return info
}

// FindNodeBinary locates node.exe on the system
func FindNodeBinary() (string, error) {
	if p, err := exec.LookPath("node.exe"); err == nil {
		return p, nil
	}
	if p, err := exec.LookPath("node"); err == nil {
		return p, nil
	}

	commonPaths := []string{
		`C:\Program Files\nodejs\node.exe`,
		`C:\Program Files (x86)\nodejs\node.exe`,
	}
	for _, p := range commonPaths {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("node binary not found on system")
}

// FindPackageManagerCmd locates npm.cmd, pnpm.cmd, or yarn.cmd
func FindPackageManagerCmd(pm string) string {
	cmdName := pm + ".cmd"
	if p, err := exec.LookPath(cmdName); err == nil {
		return p
	}
	if p, err := exec.LookPath(pm); err == nil {
		return p
	}

	commonNpm := `C:\Program Files\nodejs\npm.cmd`
	if pm == "npm" {
		if _, err := os.Stat(commonNpm); err == nil {
			return commonNpm
		}
	}
	return pm
}

// DetectNodeEnvironment returns system Node.js and npm metadata
func DetectNodeEnvironment() NodeEnvInfo {
	nodePath, err := FindNodeBinary()
	if err != nil {
		return NodeEnvInfo{Available: false}
	}

	info := NodeEnvInfo{
		Available:      true,
		NodePath:       nodePath,
		PackageManager: "npm",
	}

	// Get node version
	if out, err := winutil.Command(nodePath, "--version").CombinedOutput(); err == nil {
		info.NodeVersion = strings.TrimSpace(string(out))
	}

	npmCmd := FindPackageManagerCmd("npm")
	info.NpmPath = npmCmd
	if out, err := winutil.Command(npmCmd, "--version").CombinedOutput(); err == nil {
		info.NpmVersion = strings.TrimSpace(string(out))
	}

	return info
}

// IsDevPortListening checks if a port has an active TCP listener
func IsDevPortListening(port int) bool {
	if port <= 0 {
		return false
	}
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 150*time.Millisecond)
	if err == nil {
		_ = conn.Close()
		return true
	}
	return false
}
