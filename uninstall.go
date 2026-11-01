package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/laragal/laragal/internal/config"
	"github.com/laragal/laragal/internal/runtimes"
	"github.com/laragal/laragal/internal/services/caddy"
)

var phpVersionRegex = regexp.MustCompile(`^php-\d+\.\d+$`)

// UninstallError represents a refusal to uninstall a runtime, containing a translation key and optional params.
type UninstallError struct {
	Key    string            `json:"key"`
	Params map[string]string `json:"params,omitempty"`
}

func (e *UninstallError) Error() string {
	return e.Key
}

// IsRuntimeManaged returns true if the specified runtime name is a PHP runtime located inside LaraGal's bin directory.
func (a *App) IsRuntimeManaged(name string) bool {
	name = strings.TrimSpace(name)
	if !phpVersionRegex.MatchString(name) {
		return false
	}
	baseBin := a.getBaseBinDir()
	targetDir := filepath.Join(baseBin, name)
	exe1 := filepath.Join(targetDir, "php.exe")
	exe2 := filepath.Join(targetDir, "php-cgi.exe")

	if info, err := os.Stat(exe1); err == nil && !info.IsDir() {
		return true
	}
	if info, err := os.Stat(exe2); err == nil && !info.IsDir() {
		return true
	}
	return false
}

// GetUninstallBlockers returns the list of project domains currently using the specified runtime.
func (a *App) GetUninstallBlockers(name string) []string {
	name = strings.TrimSpace(name)
	ver := strings.TrimPrefix(name, "php-")
	normTarget := runtimes.NormalizePHPVersion(ver)

	a.sitesMu.RLock()
	sitesList := make([]caddy.SiteDef, len(a.sites))
	copy(sitesList, a.sites)
	a.sitesMu.RUnlock()

	if len(sitesList) == 0 {
		cachePath := a.getCacheFilePath()
		if data, err := os.ReadFile(cachePath); err == nil {
			var cached []caddy.SiteDef
			if err := json.Unmarshal(data, &cached); err == nil {
				sitesList = cached
			}
		}
	}

	seen := make(map[string]bool)
	var blockers []string
	for _, site := range sitesList {
		if runtimes.NormalizePHPVersion(site.PHPVersion) == normTarget {
			domain := site.Domain
			if domain == "" {
				domain = site.Name
			}
			if domain != "" && !seen[domain] {
				seen[domain] = true
				blockers = append(blockers, domain)
			}
		}
	}
	if blockers == nil {
		return []string{}
	}
	return blockers
}

// countInstalledPHPVersions returns the set of all unique installed PHP versions (managed + external).
func (a *App) countInstalledPHPVersions() map[string]bool {
	installed := make(map[string]bool)

	// 1. Managed on-disk in bin directory
	baseBin := a.getBaseBinDir()
	if entries, err := os.ReadDir(baseBin); err == nil {
		for _, e := range entries {
			if e.IsDir() && phpVersionRegex.MatchString(e.Name()) {
				cgi := filepath.Join(baseBin, e.Name(), "php-cgi.exe")
				exe := filepath.Join(baseBin, e.Name(), "php.exe")
				if info, err := os.Stat(cgi); err == nil && !info.IsDir() {
					v := strings.TrimPrefix(e.Name(), "php-")
					installed[runtimes.NormalizePHPVersion(v)] = true
				} else if info, err := os.Stat(exe); err == nil && !info.IsDir() {
					v := strings.TrimPrefix(e.Name(), "php-")
					installed[runtimes.NormalizePHPVersion(v)] = true
				}
			}
		}
	}

	// 2. WinGet packages
	if localApp := os.Getenv("LOCALAPPDATA"); localApp != "" {
		wingetGlob := filepath.Join(localApp, "Microsoft", "WinGet", "Packages")
		if entries, err := os.ReadDir(wingetGlob); err == nil {
			for _, e := range entries {
				lower := strings.ToLower(e.Name())
				if strings.Contains(lower, "php") {
					for _, catalogVer := range []string{"8.1", "8.2", "8.3", "8.4", "8.5"} {
						if strings.Contains(lower, catalogVer) {
							wingetExe := filepath.Join(wingetGlob, e.Name(), "php.exe")
							wingetCgi := filepath.Join(wingetGlob, e.Name(), "php-cgi.exe")
							if _, statErr := os.Stat(wingetExe); statErr == nil {
								installed[catalogVer] = true
							} else if _, statErr := os.Stat(wingetCgi); statErr == nil {
								installed[catalogVer] = true
							}
						}
					}
				}
			}
		}
	}

	return installed
}

// UninstallRuntime safely uninstalls a managed PHP runtime (php-X.Y) from LaraGal.
func (a *App) UninstallRuntime(name string) (UIMessage, error) {
	name = strings.TrimSpace(name)

	// ONLY for php-X.Y. NOT available for mysql/MariaDB or other services.
	if !phpVersionRegex.MatchString(name) {
		return UIMessage{}, fmt.Errorf("uninstall is only supported for managed PHP runtimes (php-X.Y)")
	}

	// Path safety checks: traversal refusal
	if strings.Contains(name, "/") || strings.Contains(name, "\\") || strings.Contains(name, "..") {
		return UIMessage{}, fmt.Errorf("invalid runtime path: path traversal is not allowed")
	}

	absBin, err := filepath.Abs(a.getBaseBinDir())
	if err != nil {
		return UIMessage{}, fmt.Errorf("failed to resolve bin directory: %w", err)
	}

	targetDir := filepath.Join(absBin, name)
	cleanTarget := filepath.Clean(targetDir)

	if filepath.Dir(cleanTarget) != absBin || filepath.Base(cleanTarget) != name {
		return UIMessage{}, fmt.Errorf("invalid runtime path: target must be a direct child of bin directory")
	}

	// Never touch data directory
	absData, _ := filepath.Abs(a.getDataDir())
	if cleanTarget == absData || strings.HasPrefix(cleanTarget, absData+string(filepath.Separator)) {
		return UIMessage{}, fmt.Errorf("refusing to touch data directory")
	}

	// Never follow symlinks/junctions out of bin
	if lstat, err := os.Lstat(cleanTarget); err == nil {
		if lstat.Mode()&os.ModeSymlink != 0 {
			evalTarget, err := filepath.EvalSymlinks(cleanTarget)
			if err != nil || filepath.Dir(evalTarget) != absBin || strings.HasPrefix(evalTarget, absData) {
				return UIMessage{}, fmt.Errorf("refusing to follow symlink or junction outside bin")
			}
		}
	}

	// Managed check: must be managed by LaraGal
	ver := strings.TrimPrefix(name, "php-")
	normVer := runtimes.NormalizePHPVersion(ver)
	if !a.IsRuntimeManaged(name) {
		wingetCmd := fmt.Sprintf("winget uninstall --id PHP.PHP.%s", normVer)
		return UIMessage{
			Key:    "runtime_uninstall_external",
			Params: map[string]string{"command": wingetCmd},
		}, &UninstallError{
			Key:    "runtime_uninstall_external",
			Params: map[string]string{"command": wingetCmd},
		}
	}

	// In-use check: project blockers
	blockers := a.GetUninstallBlockers(name)
	if len(blockers) > 0 {
		domainsStr := strings.Join(blockers, ", ")
		return UIMessage{
			Key:    "runtime_uninstall_in_use",
			Params: map[string]string{"domains": domainsStr},
		}, &UninstallError{
			Key:    "runtime_uninstall_in_use",
			Params: map[string]string{"domains": domainsStr},
		}
	}

	// Last installed PHP version or default PHP version in config check
	defVer := ""
	if a.cfg != nil && a.cfg.DefaultPHP != "" {
		defVer = runtimes.NormalizePHPVersion(a.cfg.DefaultPHP)
	}
	if defVer != "" && defVer == normVer {
		return UIMessage{
			Key: "runtime_uninstall_last_or_default",
		}, &UninstallError{
			Key: "runtime_uninstall_last_or_default",
		}
	}

	installedPHPs := a.countInstalledPHPVersions()
	if len(installedPHPs) <= 1 {
		return UIMessage{
			Key: "runtime_uninstall_last_or_default",
		}, &UninstallError{
			Key: "runtime_uninstall_last_or_default",
		}
	}

	// Stop running service in supervisor first
	if a.supervisor != nil {
		st, ok := a.supervisor.GetServiceStatus(name)
		if ok && st.Running {
			_ = a.supervisor.StopService(name)
			port := runtimes.PHPPortForVersion(normVer)
			if port > 0 {
				a.waitForPort(port, false, 3*time.Second)
			}
		}
		a.supervisor.RemoveService(name)
	}
	a.clearUserStopped(name)

	// Delete folder
	if err := os.RemoveAll(cleanTarget); err != nil {
		return UIMessage{}, fmt.Errorf("failed to delete runtime directory: %w", err)
	}

	return UIMessage{
		Key:    "runtime_uninstall_success",
		Params: map[string]string{"name": name},
	}, nil
}

// GetHiddenRuntimes returns the list of runtime keys hidden by the user.
func (a *App) GetHiddenRuntimes() []string {
	if a.cfg == nil || a.cfg.HiddenRuntimes == nil {
		return []string{}
	}
	res := make([]string, len(a.cfg.HiddenRuntimes))
	copy(res, a.cfg.HiddenRuntimes)
	return res
}

// SetRuntimeHidden toggles whether a runtime card is hidden in the UI catalog.
func (a *App) SetRuntimeHidden(name string, hidden bool) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("runtime name cannot be empty")
	}
	if a.cfg == nil {
		a.cfg = config.NewDefaultConfig()
	}

	newList := make([]string, 0, len(a.cfg.HiddenRuntimes))
	found := false
	for _, h := range a.cfg.HiddenRuntimes {
		if strings.EqualFold(h, name) {
			found = true
			if hidden {
				newList = append(newList, h)
			}
		} else {
			newList = append(newList, h)
		}
	}
	if hidden && !found {
		newList = append(newList, name)
	}
	a.cfg.HiddenRuntimes = newList

	if a.cfgPath != "" {
		return a.cfg.Save(a.cfgPath)
	}
	return nil
}
