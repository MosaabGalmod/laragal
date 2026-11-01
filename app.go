package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/laragal/laragal/internal/certs"
	"github.com/laragal/laragal/internal/config"
	"github.com/laragal/laragal/internal/downloader"
	"github.com/laragal/laragal/internal/hosts"
	"github.com/laragal/laragal/internal/privilege"
	"github.com/laragal/laragal/internal/process"
	"github.com/laragal/laragal/internal/projects"
	"github.com/laragal/laragal/internal/runtimes"
	"github.com/laragal/laragal/internal/services/adminer"
	"github.com/laragal/laragal/internal/services/caddy"
	"github.com/laragal/laragal/internal/services/mysql"
	"github.com/laragal/laragal/internal/services/php"
	"github.com/laragal/laragal/internal/services/phpmyadmin"
	"github.com/laragal/laragal/internal/updater"
	"github.com/laragal/laragal/internal/winutil"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)


type UIMessage struct {
	Key    string            `json:"key"`
	Params map[string]string `json:"params,omitempty"`
}

type FrameCheck struct {
	Frameable bool   `json:"frameable"`
	Reason    string `json:"reason"`
	Status    int    `json:"status"`
}

type App struct {
	ctx               context.Context
	cfg               *config.AppConfig
	supervisor        *process.Supervisor
	sites             []caddy.SiteDef
	sitesMu           sync.RWMutex
	cfgPath           string
	userStopped       map[string]bool
	userStoppedMu     sync.RWMutex
	phpReleasesURL    string
	baseBinDir        string
	tempDir           string
	dataDir           string
	checkClient       *http.Client
	caddyRootCertPath string
	registry          updater.Registry
}

func (a *App) getBaseBinDir() string {
	if a.baseBinDir != "" {
		return a.baseBinDir
	}
	return "bin"
}

func (a *App) getTempDir() string {
	if a.tempDir != "" {
		return a.tempDir
	}
	return filepath.Join("data", "tmp")
}

func (a *App) getDataDir() string {
	if a.dataDir != "" {
		return a.dataDir
	}
	return "data"
}

func (a *App) getActivePHPVersion() string {
	if a.supervisor != nil {
		for _, rt := range runtimes.All() {
			if strings.HasPrefix(rt.Name, "php-") {
				if st, ok := a.supervisor.GetServiceStatus(rt.Name); ok && st.Running {
					return strings.TrimPrefix(rt.Name, "php-")
				}
			}
		}
	}
	defVer := "8.4"
	if a.cfg != nil && a.cfg.DefaultPHP != "" {
		defVer = a.cfg.DefaultPHP
	}
	return defVer
}

func (a *App) getRegistry() updater.Registry {
	reg := a.registry
	if reg == nil {
		reg = updater.DefaultRegistry
	}
	if pmaSrc, ok := reg["phpmyadmin"].(*updater.PhpMyAdminSource); ok && pmaSrc != nil {
		pmaSrc.ActivePHPVersion = a.getActivePHPVersion()
	}
	if mariaSrc, ok := reg["mysql"].(*updater.MariaDBSource); ok && mariaSrc != nil {
		if installed := a.detectInstalledVersion("mysql"); installed != "" {
			mariaSrc.InstalledVersion = installed
		}
	}
	return reg
}

func (a *App) SetRegistry(r updater.Registry) {
	a.registry = r
}

func (a *App) SetPHPReleasesURL(u string) {
	a.phpReleasesURL = u
}

func NewApp() *App {
	cfgPath := "laragal.json"
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		cfg = config.NewDefaultConfig()
		_ = cfg.Save(cfgPath)
	}

	sup := process.NewSupervisor()
	app := &App{
		cfg:         cfg,
		supervisor:  sup,
		sites:       make([]caddy.SiteDef, 0),
		cfgPath:     cfgPath,
		userStopped: make(map[string]bool),
	}
	sup.OnExit = func(name string) {
		if app.ctx != nil {
			runtime.EventsEmit(app.ctx, "services-updated", app.GetServices())
		}
	}
	app.loadSitesCache()
	return app
}

func (a *App) getCacheFilePath() string {
	return filepath.Join("data", "sites_cache.json")
}

func (a *App) loadSitesCache() {
	data, err := os.ReadFile(a.getCacheFilePath())
	if err != nil {
		return
	}
	var cached []caddy.SiteDef
	if err := json.Unmarshal(data, &cached); err == nil && len(cached) > 0 {
		a.sitesMu.Lock()
		a.sites = cached
		a.sitesMu.Unlock()
	}
}

func (a *App) saveSitesCache() {
	a.sitesMu.RLock()
	count := len(a.sites)
	var data []byte
	var err error
	if count > 0 {
		data, err = json.MarshalIndent(a.sites, "", "  ")
	}
	a.sitesMu.RUnlock()

	if count == 0 || err != nil {
		return
	}
	_ = os.MkdirAll("data", 0755)
	_ = os.WriteFile(a.getCacheFilePath(), data, 0644)
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	_ = os.MkdirAll(a.cfg.ProjectsPath, 0755)
	a.initTray()

	a.loadSitesCache()

	// 1. Scan projects synchronously at the start of background services goroutine so UI has sites in < 20ms and Caddyfile is written
	go func() {
		sites := a.ScanProjects()
		a.saveSitesCache()
		if a.ctx != nil {
			runtime.EventsEmit(a.ctx, "sites-updated", sites)
		}

		// 2. Launch background services asynchronously (Caddyfile is already written)
		a.StartServices()
		_ = a.EnsureAdminerInstalled()
		if a.ctx != nil {
			runtime.EventsEmit(a.ctx, "services-updated", a.GetServices())
		}

		// 3. Check runtime updates in background without blocking UI
		go func() {
			updates := a.CheckRuntimeUpdates()
			if a.ctx != nil {
				runtime.EventsEmit(a.ctx, "runtime-updates", updates)
			}
		}()
	}()
}

func (a *App) shutdown(ctx context.Context) {
	_ = a.supervisor.StopAll()
}

func (a *App) findCaddyExecutable() string {
	localCaddy := filepath.Join("bin", "caddy", "caddy.exe")
	if _, err := os.Stat(localCaddy); err == nil {
		return localCaddy
	}
	if path, err := exec.LookPath("caddy.exe"); err == nil {
		return path
	}
	if path, err := exec.LookPath("caddy"); err == nil {
		return path
	}
	return ""
}

func (a *App) findPHPExecutable() string {
	candidates := []string{
		filepath.Join("bin", "php-8.4", "php-cgi.exe"),
		filepath.Join("bin", "php-8.3", "php-cgi.exe"),
		filepath.Join("bin", "php", "php-cgi.exe"),
	}

	if localApp := os.Getenv("LOCALAPPDATA"); localApp != "" {
		wingetGlob := filepath.Join(localApp, "Microsoft", "WinGet", "Packages")
		if entries, err := os.ReadDir(wingetGlob); err == nil {
			for _, e := range entries {
				if strings.Contains(strings.ToLower(e.Name()), "php") {
					wingetPHP := filepath.Join(wingetGlob, e.Name(), "php-cgi.exe")
					if _, statErr := os.Stat(wingetPHP); statErr == nil {
						candidates = append(candidates, wingetPHP)
					}
				}
			}
		}
	}

	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}

	if path, err := exec.LookPath("php-cgi.exe"); err == nil {
		return path
	}
	if path, err := exec.LookPath("php-cgi"); err == nil {
		return path
	}
	return ""
}

func (a *App) findPHPExecutableForVersion(version string) string {
	normVer := projects.NormalizePHPVersion(version)
	baseBin := a.getBaseBinDir()
	candidates := []string{
		filepath.Join(baseBin, fmt.Sprintf("php-%s", normVer), "php-cgi.exe"),
		filepath.Join(baseBin, fmt.Sprintf("php%s", strings.ReplaceAll(normVer, ".", "")), "php-cgi.exe"),
		filepath.Join(baseBin, "php", "php-cgi.exe"),
	}

	if localApp := os.Getenv("LOCALAPPDATA"); localApp != "" {
		wingetGlob := filepath.Join(localApp, "Microsoft", "WinGet", "Packages")
		if entries, err := os.ReadDir(wingetGlob); err == nil {
			for _, e := range entries {
				lower := strings.ToLower(e.Name())
				if strings.Contains(lower, "php") && strings.Contains(lower, normVer) {
					wingetPHP := filepath.Join(wingetGlob, e.Name(), "php-cgi.exe")
					if _, statErr := os.Stat(wingetPHP); statErr == nil {
						candidates = append(candidates, wingetPHP)
					}
				}
			}
		}
	}

	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}

	if normVer == "8.5" {
		if fb := a.findPHPExecutableForVersion("8.4"); fb != "" {
			return fb
		}
	}

	return a.findPHPExecutable()
}

func (a *App) findRedisExecutable() string {
	candidates := []string{
		filepath.Join("bin", "redis", "redis-server.exe"),
	}
	if a.baseBinDir != "" {
		candidates = append([]string{
			filepath.Join(a.baseBinDir, "redis", "redis-server.exe"),
		}, candidates...)
		if matches, err := filepath.Glob(filepath.Join(a.baseBinDir, "redis", "*", "redis-server.exe")); err == nil {
			candidates = append(candidates, matches...)
		}
	}
	if matches, err := filepath.Glob(filepath.Join("bin", "redis", "*", "redis-server.exe")); err == nil {
		candidates = append(candidates, matches...)
	}

	if localApp := os.Getenv("LOCALAPPDATA"); localApp != "" {
		wingetGlob := filepath.Join(localApp, "Microsoft", "WinGet", "Packages")
		if entries, err := os.ReadDir(wingetGlob); err == nil {
			for _, e := range entries {
				if strings.Contains(strings.ToLower(e.Name()), "redis") {
					packageDir := filepath.Join(wingetGlob, e.Name())
					subEntries, _ := os.ReadDir(packageDir)
					for _, se := range subEntries {
						if se.IsDir() {
							c := filepath.Join(packageDir, se.Name(), "redis-server.exe")
							if _, statErr := os.Stat(c); statErr == nil {
								candidates = append(candidates, c)
							}
						}
					}
					cDirect := filepath.Join(packageDir, "redis-server.exe")
					if _, statErr := os.Stat(cDirect); statErr == nil {
						candidates = append(candidates, cDirect)
					}
				}
			}
		}
	}

	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}

	if path, err := exec.LookPath("redis-server.exe"); err == nil {
		return path
	}
	if path, err := exec.LookPath("redis-server"); err == nil {
		return path
	}
	return ""
}

func (a *App) findMySQLExecutable() string {
	candidates := []string{
		filepath.Join("bin", "mysql", "bin", "mysqld.exe"),
		filepath.Join("bin", "mysql", "mysqld.exe"),
		filepath.Join("bin", "mariadb", "bin", "mysqld.exe"),
		filepath.Join("bin", "mariadb", "mysqld.exe"),
	}

	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}

	if path, err := exec.LookPath("mysqld.exe"); err == nil {
		return path
	}
	if path, err := exec.LookPath("mysqld"); err == nil {
		return path
	}
	return ""
}

func (a *App) getMySQLPendingUpgradeMarker() string {
	dataDir := filepath.Join("data", "mysql")
	if a.dataDir != "" {
		dataDir = filepath.Join(a.dataDir, "mysql")
	}
	return filepath.Join(dataDir, ".pending_upgrade")
}

func (a *App) markMySQLPendingUpgrade() {
	marker := a.getMySQLPendingUpgradeMarker()
	_ = os.MkdirAll(filepath.Dir(marker), 0755)
	_ = os.WriteFile(marker, []byte(time.Now().Format(time.RFC3339)), 0644)
}

func (a *App) isMySQLPendingUpgrade() bool {
	marker := a.getMySQLPendingUpgradeMarker()
	_, err := os.Stat(marker)
	return err == nil
}

func (a *App) clearMySQLPendingUpgrade() {
	marker := a.getMySQLPendingUpgradeMarker()
	_ = os.Remove(marker)
}

func (a *App) runMySQLUpgradeIfNeeded() error {
	if !a.isMySQLPendingUpgrade() {
		return nil
	}

	candidates := []string{
		filepath.Join(a.getBaseBinDir(), "mysql", "bin", "mariadb-upgrade.exe"),
		filepath.Join(a.getBaseBinDir(), "mysql", "bin", "mysql_upgrade.exe"),
		filepath.Join(a.getBaseBinDir(), "mysql", "mariadb-upgrade.exe"),
		filepath.Join(a.getBaseBinDir(), "mysql", "mysql_upgrade.exe"),
	}
	var upgradeExe string
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			upgradeExe = c
			break
		}
	}
	if upgradeExe == "" {
		if p, err := exec.LookPath("mariadb-upgrade.exe"); err == nil {
			upgradeExe = p
		} else if p, err := exec.LookPath("mysql_upgrade.exe"); err == nil {
			upgradeExe = p
		}
	}

	if upgradeExe == "" {
		a.clearMySQLPendingUpgrade()
		return nil
	}

	logsDir := filepath.Join("data", "logs")
	if a.dataDir != "" {
		logsDir = filepath.Join(a.dataDir, "logs")
	}
	_ = os.MkdirAll(logsDir, 0755)
	logFile := filepath.Join(logsDir, "mariadb-upgrade.log")

	cmd := winutil.Command(upgradeExe, "--port=3306")
	out, err := cmd.CombinedOutput()

	f, openErr := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if openErr == nil {
		_, _ = f.WriteString(fmt.Sprintf("\n--- mariadb-upgrade executed at %s ---\n", time.Now().Format(time.RFC3339)))
		_, _ = f.Write(out)
		_, _ = f.WriteString("\n")
		_ = f.Close()
	}

	a.clearMySQLPendingUpgrade()
	return err
}


func (a *App) markUserStopped(name string) {
	a.userStoppedMu.Lock()
	defer a.userStoppedMu.Unlock()
	if a.userStopped == nil {
		a.userStopped = make(map[string]bool)
	}
	a.userStopped[name] = true
}

func (a *App) clearUserStopped(name string) {
	a.userStoppedMu.Lock()
	defer a.userStoppedMu.Unlock()
	if a.userStopped != nil {
		delete(a.userStopped, name)
	}
}

func (a *App) isUserStopped(name string) bool {
	a.userStoppedMu.RLock()
	defer a.userStoppedMu.RUnlock()
	if a.userStopped == nil {
		return false
	}
	return a.userStopped[name]
}

func (a *App) ensurePHPServiceRunning(version string) error {
	if a.supervisor == nil {
		return nil
	}
	normVer := runtimes.NormalizePHPVersion(version)
	port := runtimes.PHPPortForVersion(normVer)
	serviceName := fmt.Sprintf("php-%s", normVer)
	if a.isUserStopped(serviceName) {
		return nil
	}

	status, exists := a.supervisor.GetServiceStatus(serviceName)
	if exists && status.Running {
		return nil
	}

	phpExe := a.findPHPExecutableForVersion(normVer)
	if phpExe == "" {
		return fmt.Errorf("PHP executable for %s not found", normVer)
	}

	_ = php.EnsurePHPIni(filepath.Dir(phpExe), normVer)

	phpEnv := []string{
		"PHP_FCGI_CHILDREN=8",
		"PHP_FCGI_MAX_REQUESTS=10000",
	}
	return a.supervisor.StartServiceWithEnv(serviceName, phpExe, []string{"-b", fmt.Sprintf("127.0.0.1:%d", port)}, phpEnv, port)
}

func (a *App) findCaddyRootCert() string {
	candidate1 := filepath.Join("data", "caddy", "pki", "authorities", "local", "root.crt")
	if _, err := os.Stat(candidate1); err == nil {
		return candidate1
	}

	if appData := os.Getenv("APPDATA"); appData != "" {
		candidate2 := filepath.Join(appData, "caddy", "pki", "authorities", "local", "root.crt")
		if _, err := os.Stat(candidate2); err == nil {
			_ = os.MkdirAll(filepath.Dir(candidate1), 0755)
			if data, readErr := os.ReadFile(candidate2); readErr == nil {
				_ = os.WriteFile(candidate1, data, 0644)
			}
			return candidate2
		}
	}

	if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
		candidate3 := filepath.Join(localAppData, "caddy", "pki", "authorities", "local", "root.crt")
		if _, err := os.Stat(candidate3); err == nil {
			_ = os.MkdirAll(filepath.Dir(candidate1), 0755)
			if data, readErr := os.ReadFile(candidate3); readErr == nil {
				_ = os.WriteFile(candidate1, data, 0644)
			}
			return candidate3
		}
	}

	return ""
}

func (a *App) StartServices() {
	phpEnv := []string{
		"PHP_FCGI_CHILDREN=8",
		"PHP_FCGI_MAX_REQUESTS=10000",
	}

	// 1. Start PHP FastCGI services from discovered / catalog PHP versions
	defVer := "8.4"
	if a.cfg != nil && a.cfg.DefaultPHP != "" {
		defVer = runtimes.NormalizePHPVersion(a.cfg.DefaultPHP)
	}
	versionsToStart := []string{defVer}
	for _, ver := range []string{"8.4", "8.3"} {
		if ver != defVer {
			versionsToStart = append(versionsToStart, ver)
		}
	}
	for _, ver := range versionsToStart {
		rtName := fmt.Sprintf("php-%s", ver)
		port := runtimes.PHPPortForVersion(ver)
		if phpExe := a.findPHPExecutableForVersion(ver); phpExe != "" {
			_ = php.EnsurePHPIni(filepath.Dir(phpExe), ver)
			_ = a.supervisor.StartServiceWithEnv(rtName, phpExe, []string{"-b", fmt.Sprintf("127.0.0.1:%d", port)}, phpEnv, port)
		}
	}

	// 2. Start Caddy on 443
	caddyExe := a.findCaddyExecutable()
	if caddyExe != "" {
		caddyDir := filepath.Join("data", "caddy")
		_ = os.MkdirAll(caddyDir, 0755)
		caddyfilePath := filepath.Join(caddyDir, "Caddyfile")
		if _, err := os.Stat(caddyfilePath); os.IsNotExist(err) {
			_ = a.ReloadCaddy()
		}
		_ = a.supervisor.StartService("caddy", caddyExe, []string{"run", "--config", caddyfilePath, "--adapter", "caddyfile"}, 443)
	}

	// 3. Start Mailpit (if present)
	mailpitExe := filepath.Join("bin", "mailpit", "mailpit.exe")
	if _, err := os.Stat(mailpitExe); err == nil {
		_ = a.supervisor.StartService("mailpit", mailpitExe, []string{"--smtp-bind-addr", "127.0.0.1:1025", "--ui-bind-addr", "127.0.0.1:8025"}, 1025)
	}

	// 4. Start MySQL/MariaDB if portable binary exists and port 3306 is not in use
	if !a.IsPortListening(3306) {
		if mysqldExe := a.findMySQLExecutable(); mysqldExe != "" {
			dataDir := filepath.Join("data", "mysql")
			if err := mysql.EnsureDataDir(mysqldExe, dataDir); err == nil {
				absDataDir, _ := filepath.Abs(dataDir)
				args := []string{
					fmt.Sprintf("--datadir=%s", absDataDir),
					"--port=3306",
					"--bind-address=127.0.0.1",
					"--console",
				}
				_ = a.supervisor.StartService("mysql", mysqldExe, args, 3306)
			}
		}
	}
}

func (a *App) ToggleService(name string) (bool, error) {
	var port int
	isPHP := strings.HasPrefix(name, "php-")
	if isPHP {
		ver := strings.TrimPrefix(name, "php-")
		port = runtimes.PHPPortForVersion(ver)
	} else {
		switch name {
		case "caddy":
			port = 443
		case "mysql":
			port = 3306
		case "redis":
			port = 6379
		case "mailpit":
			port = 1025
		}
	}

	status, exists := a.supervisor.GetServiceStatus(name)
	isRunning := (exists && status.Running)
	if port > 0 && a.IsPortListening(port) {
		isRunning = true
	}

	// 1. If currently running or port listening -> STOP IT
	if isRunning {
		a.markUserStopped(name)
		_ = a.supervisor.StopService(name)
		if !exists && port > 0 {
			_ = winutil.KillProcessesOnPort(port)
		}
		if port > 0 {
			if !a.waitForPort(port, false, 5*time.Second) {
				return true, fmt.Errorf("failed to stop service %s: port %d is still listening", name, port)
			}
		}
		return false, nil
	}

	// 2. If stopped -> START IT (clean lingering port first)
	a.clearUserStopped(name)
	if port > 0 && a.IsPortListening(port) {
		_ = winutil.KillProcessesOnPort(port)
		_ = a.waitForPort(port, false, 5*time.Second)
	}

	if isPHP {
		ver := strings.TrimPrefix(name, "php-")
		phpExe := a.findPHPExecutableForVersion(ver)
		if phpExe == "" {
			return false, fmt.Errorf("PHP executable for %s not found; download it from runtimes", ver)
		}
		_ = php.EnsurePHPIni(filepath.Dir(phpExe), ver)
		phpEnv := []string{
			"PHP_FCGI_CHILDREN=8",
			"PHP_FCGI_MAX_REQUESTS=10000",
		}
		err := a.supervisor.StartServiceWithEnv(name, phpExe, []string{"-b", fmt.Sprintf("127.0.0.1:%d", port)}, phpEnv, port)
		if err != nil {
			return false, err
		}
		if !a.waitForPort(port, true, 5*time.Second) {
			return false, fmt.Errorf("failed to start service %s: port %d did not respond within 5 seconds", name, port)
		}
		return true, nil
	}

	switch name {
	case "caddy":
		caddyExe := a.findCaddyExecutable()
		if caddyExe == "" {
			return false, fmt.Errorf("Caddy binary not found; install it from runtimes")
		}
		_ = a.ReloadCaddy()
		caddyDir := filepath.Join("data", "caddy")
		_ = os.MkdirAll(caddyDir, 0755)
		caddyfilePath := filepath.Join(caddyDir, "Caddyfile")
		err := a.supervisor.StartService("caddy", caddyExe, []string{"run", "--config", caddyfilePath, "--adapter", "caddyfile"}, 443)
		return err == nil, err

	case "mailpit":
		mailpitExe := filepath.Join("bin", "mailpit", "mailpit.exe")
		if _, err := os.Stat(mailpitExe); err != nil {
			return false, fmt.Errorf("Mailpit binary not found")
		}
		err := a.supervisor.StartService("mailpit", mailpitExe, []string{"--smtp-bind-addr", "127.0.0.1:1025", "--ui-bind-addr", "127.0.0.1:8025"}, 1025)
		return err == nil, err

	case "redis":
		redisExe := a.findRedisExecutable()
		if redisExe == "" {
			return false, fmt.Errorf("Redis binary not found on this system")
		}
		redisDir := filepath.Dir(redisExe)
		args := []string{}
		if _, err := os.Stat(filepath.Join(redisDir, "redis.conf")); err == nil {
			args = append(args, "redis.conf")
		} else if _, err := os.Stat(filepath.Join(redisDir, "redis.windows.conf")); err == nil {
			args = append(args, "redis.windows.conf")
		}
		err := a.supervisor.StartServiceInDir("redis", redisExe, args, redisDir, 6379)
		return err == nil, err

	case "mysql":
		mysqldExe := a.findMySQLExecutable()
		if mysqldExe == "" {
			if a.IsPortListening(3306) {
				if a.isMySQLPendingUpgrade() {
					_ = a.runMySQLUpgradeIfNeeded()
				}
				return true, nil
			}
			return false, fmt.Errorf("MySQL/MariaDB binary not found; install it from runtimes")
		}
		dataDir := filepath.Join("data", "mysql")
		if err := mysql.EnsureDataDir(mysqldExe, dataDir); err != nil {
			return false, fmt.Errorf("failed to initialize MySQL data directory: %w", err)
		}
		if a.IsPortListening(3306) {
			if a.isMySQLPendingUpgrade() {
				_ = a.runMySQLUpgradeIfNeeded()
			}
			return true, nil
		}
		absDataDir, _ := filepath.Abs(dataDir)
		args := []string{
			fmt.Sprintf("--datadir=%s", absDataDir),
			"--port=3306",
			"--bind-address=127.0.0.1",
			"--console",
		}
		err := a.supervisor.StartService("mysql", mysqldExe, args, 3306)
		if err != nil {
			return false, err
		}
		if a.isMySQLPendingUpgrade() {
			_ = a.waitForPort(3306, true, 5*time.Second)
			_ = a.runMySQLUpgradeIfNeeded()
		}
		return true, nil
	}

	return false, fmt.Errorf("unknown service: %s", name)
}

func (a *App) CheckPortConflict(port int) UIMessage {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 200*time.Millisecond)
	if err == nil {
		_ = conn.Close()
		for _, s := range a.GetServices() {
			if s.Running && s.Port == port {
				return UIMessage{}
			}
		}
		return UIMessage{
			Key:    "port_conflict_warning",
			Params: map[string]string{"port": fmt.Sprintf("%d", port)},
		}
	}
	return UIMessage{}
}

func (a *App) getDiscoveredPHPServices() []string {
	seen := make(map[string]bool)
	var list []string

	// 1. Catalog PHP runtimes
	for _, rt := range runtimes.All() {
		if strings.HasPrefix(rt.Name, "php-") {
			if !seen[rt.Name] {
				seen[rt.Name] = true
				list = append(list, rt.Name)
			}
		}
	}

	// 2. On-disk php-X.Y folders in baseBinDir containing php-cgi.exe (or php.exe)
	binDir := a.getBaseBinDir()
	if entries, err := os.ReadDir(binDir); err == nil {
		for _, e := range entries {
			if e.IsDir() && strings.HasPrefix(e.Name(), "php-") {
				cgi := filepath.Join(binDir, e.Name(), "php-cgi.exe")
				exe := filepath.Join(binDir, e.Name(), "php.exe")
				cgiExists := false
				if info, sErr := os.Stat(cgi); sErr == nil && !info.IsDir() {
					cgiExists = true
				} else if info, sErr := os.Stat(exe); sErr == nil && !info.IsDir() {
					cgiExists = true
				}
				if cgiExists && !seen[e.Name()] {
					seen[e.Name()] = true
					list = append(list, e.Name())
				}
			}
		}
	}

	// 3. Any php-* service currently registered or running in supervisor
	if a.supervisor != nil {
		for _, st := range a.supervisor.GetAllServices() {
			if strings.HasPrefix(st.Name, "php-") && !seen[st.Name] {
				seen[st.Name] = true
				list = append(list, st.Name)
			}
		}
	}

	return list
}

func (a *App) GetServices() []process.ServiceStatus {
	phpServices := a.getDiscoveredPHPServices()
	serviceNames := []string{"caddy"}
	serviceNames = append(serviceNames, phpServices...)
	serviceNames = append(serviceNames, "mysql", "redis", "mailpit")

	result := make([]process.ServiceStatus, 0, len(serviceNames))
	for _, name := range serviceNames {
		status, _ := a.supervisor.GetServiceStatus(name)
		var port int
		if strings.HasPrefix(name, "php-") {
			ver := strings.TrimPrefix(name, "php-")
			port = runtimes.PHPPortForVersion(ver)
		} else {
			switch name {
			case "caddy":
				port = 443
			case "mysql":
				port = 3306
			case "redis":
				port = 6379
			case "mailpit":
				port = 1025
			}
		}
		if port > 0 && !status.Running && a.IsPortListening(port) {
			status.Running = true
			status.Port = port
		}
		result = append(result, status)
	}
	return result
}

func (a *App) GetSites() []caddy.SiteDef {
	a.sitesMu.RLock()
	if len(a.sites) == 0 {
		a.sitesMu.RUnlock()
		return a.ScanProjects()
	}
	defer a.sitesMu.RUnlock()
	res := make([]caddy.SiteDef, len(a.sites))
	copy(res, a.sites)
	return res
}

func (a *App) AddSite(domain string, rootPath string, phpVersion string) error {
	domain = strings.TrimSpace(domain)
	if !strings.HasSuffix(domain, a.cfg.DomainSuffix) {
		domain += a.cfg.DomainSuffix
	}

	phpPort := php.AssignPortForVersion(phpVersion)

	for _, s := range a.sites {
		if s.Domain == domain {
			return fmt.Errorf("site %s already exists", domain)
		}
	}

	newSite := caddy.SiteDef{
		Name:        strings.TrimSuffix(domain, a.cfg.DomainSuffix),
		Domain:      domain,
		Path:        filepath.Dir(rootPath),
		RootPath:    rootPath,
		PHPPort:     phpPort,
		PHPVersion:  phpVersion,
		ProjectType: "laravel",
		TypeLabel:   "Laravel",
		IsWeb:       true,
	}

	a.sites = append(a.sites, newSite)

	_ = a.ReloadCaddy()

	webDomains := make([]string, 0, len(a.sites))
	for _, s := range a.sites {
		if s.IsWeb && s.Domain != "" {
			webDomains = append(webDomains, s.Domain)
		}
	}
	_ = hosts.UpdateSystemHosts(webDomains)

	return nil
}

func (a *App) InstallSSLTrust() (UIMessage, error) {
	rootCertPath := a.findCaddyRootCert()
	if rootCertPath == "" {
		return UIMessage{Key: "ssl_root_cert_not_found"}, nil
	}

	fingerprint, err := certs.ReadCertFingerprint(rootCertPath)
	if err != nil {
		return UIMessage{}, err
	}

	trusted, _ := certs.IsCertTrustedInWindowsStore(fingerprint)
	if trusted {
		return UIMessage{Key: "ssl_already_trusted"}, nil
	}

	// 1. Try native caddy trust if available
	caddyExe := a.findCaddyExecutable()
	if caddyExe != "" {
		_ = winutil.Command(caddyExe, "trust").Run()
		if trustedNow, _ := certs.IsCertTrustedInWindowsStore(fingerprint); trustedNow {
			return UIMessage{Key: "ssl_install_success"}, nil
		}
	}

	// 2. Direct certutil command
	cmdArgs := certs.BuildTrustCertCommand(rootCertPath)
	cmd := winutil.Command(cmdArgs[0], cmdArgs[1:]...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		// 3. Fallback to elevated PowerShell Import-Certificate
		elevatedCmd := privilege.BuildElevatedCertCommand(rootCertPath)
		elCmd := winutil.Command(elevatedCmd[0], elevatedCmd[1:]...)
		elOut, elErr := elCmd.CombinedOutput()
		if elErr != nil {
			return UIMessage{}, fmt.Errorf("%s %s: %w", string(out), string(elOut), elErr)
		}
	}

	return UIMessage{Key: "ssl_install_success"}, nil
}

func (a *App) CheckFrameable(domain string) (FrameCheck, error) {
	cleanDomain := strings.TrimSpace(domain)
	if cleanDomain == "" {
		return FrameCheck{}, fmt.Errorf("domain cannot be empty")
	}

	// 1. Only allow domains that belong to the current LaraGal site list
	sites := a.GetSites()
	matched := false
	for _, s := range sites {
		if strings.EqualFold(strings.TrimSpace(s.Domain), cleanDomain) {
			matched = true
			break
		}
	}
	if !matched {
		return FrameCheck{}, fmt.Errorf("domain %q is not in the known sites list", cleanDomain)
	}

	// 2. Setup TLS transport with Caddy local root cert
	var clientTransport http.RoundTripper
	if a.checkClient != nil && a.checkClient.Transport != nil {
		clientTransport = a.checkClient.Transport
	} else {
		rootCAs, err := x509.SystemCertPool()
		if err != nil || rootCAs == nil {
			rootCAs = x509.NewCertPool()
		}
		certPath := a.caddyRootCertPath
		if certPath == "" {
			certPath = a.findCaddyRootCert()
		}
		if certPath != "" {
			if pemData, err := os.ReadFile(certPath); err == nil {
				rootCAs.AppendCertsFromPEM(pemData)
			}
		}
		clientTransport = &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: rootCAs,
			},
		}
	}

	client := &http.Client{
		Transport: clientTransport,
		Timeout:   5 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("stopped after 10 redirects")
			}
			if !strings.EqualFold(req.URL.Host, via[0].URL.Host) {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	targetURL := fmt.Sprintf("https://%s/", strings.Trim(cleanDomain, "/"))

	// 3. Perform HEAD request with 5s timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	headReq, err := http.NewRequestWithContext(ctx, http.MethodHead, targetURL, nil)
	if err != nil {
		return FrameCheck{Frameable: false, Reason: "unreachable", Status: 0}, nil
	}

	resp, err := client.Do(headReq)
	if err != nil {
		return FrameCheck{Frameable: false, Reason: "unreachable", Status: 0}, nil
	}
	defer func() {
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()

	// 4. Fall back to GET if HEAD not allowed (405 or 501)
	if resp.StatusCode == http.StatusMethodNotAllowed || resp.StatusCode == http.StatusNotImplemented {
		getReq, getErr := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
		if getErr == nil {
			if getResp, doErr := client.Do(getReq); doErr == nil {
				_ = resp.Body.Close()
				resp = getResp
			}
		}
	}

	// 5. Check X-Frame-Options
	// Treat X-Frame-Options DENY or SAMEORIGIN as not frameable (the preview origin differs).
	for _, val := range resp.Header.Values("X-Frame-Options") {
		for _, part := range strings.Split(val, ",") {
			clean := strings.ToUpper(strings.TrimSpace(part))
			if clean == "DENY" || clean == "SAMEORIGIN" {
				return FrameCheck{
					Frameable: false,
					Reason:    "x_frame_options",
					Status:    resp.StatusCode,
				}, nil
			}
		}
	}

	// 6. Parse CSP frame-ancestors: not frameable unless it contains `*`.
	for _, csp := range resp.Header.Values("Content-Security-Policy") {
		directives := strings.Split(csp, ";")
		for _, directive := range directives {
			tokens := strings.Fields(directive)
			if len(tokens) > 0 && strings.EqualFold(tokens[0], "frame-ancestors") {
				hasWildcard := false
				for _, src := range tokens[1:] {
					if src == "*" {
						hasWildcard = true
						break
					}
				}
				if !hasWildcard {
					return FrameCheck{
						Frameable: false,
						Reason:    "csp_frame_ancestors",
						Status:    resp.StatusCode,
					}, nil
				}
			}
		}
	}

	// 7. No framing restrictions found -> frameable
	return FrameCheck{
		Frameable: true,
		Reason:    "",
		Status:    resp.StatusCode,
	}, nil
}

func (a *App) downloadAndInstallArtifact(name string, url string, expectedChecksum string, checksumAlgo ...string) (UIMessage, error) {
	tempDir := a.getTempDir()
	_ = os.MkdirAll(tempDir, 0755)
	destDir := filepath.Join(a.getBaseBinDir(), name)
	if strings.HasPrefix(strings.ToLower(name), "mariadb") || strings.HasPrefix(strings.ToLower(name), "mysql") {
		destDir = filepath.Join(a.getBaseBinDir(), "mysql")
	} else if strings.ToLower(name) == "phpmyadmin" {
		destDir = filepath.Join(a.getDataDir(), "phpmyadmin")
	}

	onProgress := func(downloaded int64, total int64, percent int) {
		if a.ctx != nil {
			runtime.EventsEmit(a.ctx, "download-progress", map[string]interface{}{
				"name":       name,
				"downloaded": downloaded,
				"total":      total,
				"percent":    percent,
			})
		}
	}

	ctx := context.Background()
	if a.ctx != nil {
		ctx = a.ctx
	}

	algo := "sha256"
	if len(checksumAlgo) > 0 && checksumAlgo[0] != "" {
		algo = checksumAlgo[0]
	}

	if strings.ToLower(name) == "adminer" {
		adminerDir := filepath.Join(a.getDataDir(), "adminer")
		_ = os.MkdirAll(adminerDir, 0755)
		adminerPath := filepath.Join(adminerDir, "adminer-core.php")
		if err := downloader.DownloadFile(ctx, url, adminerPath, onProgress, algo, expectedChecksum); err != nil {
			return UIMessage{}, err
		}
		base := "."
		if a.dataDir != "" {
			base = filepath.Dir(a.dataDir)
		}
		_, _ = adminer.EnsureInstalled(base)
		if a.ctx != nil {
			runtime.EventsEmit(a.ctx, "download-progress", map[string]interface{}{
				"name":       name,
				"downloaded": 100,
				"total":      100,
				"percent":    100,
				"completed":  true,
			})
		}
		return UIMessage{Key: "adminer_download_success", Params: map[string]string{"path": filepath.ToSlash(adminerPath)}}, nil
	}

	zipPath := filepath.Join(tempDir, fmt.Sprintf("%s.zip", name))
	if err := downloader.DownloadFile(ctx, url, zipPath, onProgress, algo, expectedChecksum); err != nil {
		_ = os.Remove(zipPath)
		return UIMessage{}, err
	}

	if err := downloader.InstallArtifactZip(name, zipPath, destDir); err != nil {
		_ = os.Remove(zipPath)
		return UIMessage{}, err
	}
	_ = os.Remove(zipPath)

	if strings.ToLower(name) == "phpmyadmin" {
		base := "."
		if a.dataDir != "" {
			base = filepath.Dir(a.dataDir)
		}
		_ = phpmyadmin.EnsureConfig(base)
	}

	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "download-progress", map[string]interface{}{
			"name":       name,
			"downloaded": 100,
			"total":      100,
			"percent":    100,
			"completed":  true,
		})
	}

	return UIMessage{Key: "runtime_install_success_path", Params: map[string]string{"name": name, "dest": filepath.ToSlash(destDir)}}, nil
}

// resolveArtifactInfo resolves an artifact for any name: live release (with checksum) if available, else catalog pinned.
// Returns (url, checksum, algo, version, majorChange, err).
func (a *App) resolveArtifactInfo(name string) (string, string, string, string, bool, error) {
	norm := runtimes.NormalizeName(name)
	ctx := context.Background()
	if a.ctx != nil {
		ctx = a.ctx
	}

	// 1. Try live release via the registry
	family := updater.RuntimeFamily(norm)

	// For PHP, feed releases for uncataloged versions are only allowed if newer than highest catalog minor
	if family == "php" {
		if _, inCatalog := runtimes.Get(norm); !inCatalog {
			minor := strings.TrimPrefix(norm, "php-")
			if !updater.IsVersionNewer(updater.HighestCatalogPHPMajorMinor(), minor) {
				return "", "", "", "", false, fmt.Errorf("unknown runtime or version: %s (no verified artifact found)", name)
			}
		}
	}

	var source updater.ReleaseSource
	if family == "php" && a.phpReleasesURL != "" {
		source = &updater.PHPSource{BaseURL: a.phpReleasesURL}
	} else if a.registry != nil {
		source = a.registry[family]
	} else if updater.DefaultRegistry != nil {
		source = updater.DefaultRegistry[family]
	}

	if source != nil {
		liveReleases, err := source.Latest(ctx, norm)
		if err == nil {
			for _, rel := range liveReleases {
				if rel.Name == norm || rel.Name == name {
					if rel.Checksum != "" && rel.URL != "" {
						algo := rel.ChecksumAlgo
						if algo == "" {
							algo = "sha256"
						}

						// Determine major change against currently installed or catalog version
						curVer := a.detectInstalledVersion(name)
						if curVer == "" {
							if catRt, ok := runtimes.Get(norm); ok {
								curVer = catRt.Version
							}
						}
						isMajor := rel.MajorChange
						if curVer != "" && rel.Version != "" {
							if updater.IsMajorChange(curVer, rel.Version) {
								isMajor = true
							}
						}
						if norm == "redis" && !strings.Contains(rel.Source, "tporadowski") {
							isMajor = true
						}

						return rel.URL, rel.Checksum, algo, rel.Version, isMajor, nil
					}
				}
			}
		}
	}

	// 2. Fall back to pinned catalog artifact
	if rt, exists := runtimes.Get(norm); exists {
		if rt.URL != "" && rt.SHA256 != "" {
			return rt.URL, rt.SHA256, "sha256", rt.Version, false, nil
		}
	}

	// 3. Names not in catalog are installable ONLY via a live release with a checksum
	return "", "", "", "", false, fmt.Errorf("unknown runtime or version: %s (no verified artifact or checksum available)", name)
}

func (a *App) resolveArtifact(name string) (string, string, string, error) {
	url, checksum, algo, _, _, err := a.resolveArtifactInfo(name)
	return url, checksum, algo, err
}

func (a *App) detectInstalledVersion(name string) string {
	return updater.DetectInstalledBinaryVersion(a.getBaseBinDir(), name, a.getDataDir())
}

func (a *App) DownloadBinary(name string) (UIMessage, error) {
	url, checksum, algo, err := a.resolveArtifact(name)
	if err != nil {
		return UIMessage{}, err
	}
	if checksum == "" {
		return UIMessage{}, fmt.Errorf("no verified checksum available for binary: %s", name)
	}
	return a.downloadAndInstallArtifact(name, url, checksum, algo)
}

func (a *App) resolveUpgradeArtifact(name string) (string, string, error) {
	url, checksum, _, _, _, err := a.resolveArtifactInfo(name)
	return url, checksum, err
}

// GetRuntimeCatalog returns all runtime artifacts from the catalog plus on-disk PHP runtimes.
func (a *App) GetRuntimeCatalog() []runtimes.Artifact {
	all := runtimes.All()
	seen := make(map[string]bool, len(all))
	for _, rt := range all {
		seen[rt.Name] = true
	}

	binDir := a.getBaseBinDir()
	if entries, err := os.ReadDir(binDir); err == nil {
		for _, e := range entries {
			if e.IsDir() && strings.HasPrefix(e.Name(), "php-") {
				cgiPath := filepath.Join(binDir, e.Name(), "php-cgi.exe")
				exePath := filepath.Join(binDir, e.Name(), "php.exe")
				cgiExists := false
				if info, statErr := os.Stat(cgiPath); statErr == nil && !info.IsDir() {
					cgiExists = true
				} else if info, statErr := os.Stat(exePath); statErr == nil && !info.IsDir() {
					cgiExists = true
				}
				if cgiExists && !seen[e.Name()] {
					seen[e.Name()] = true
					minor := strings.TrimPrefix(e.Name(), "php-")
					all = append(all, runtimes.Artifact{
						Name:    e.Name(),
						Version: minor,
						PHPPort: runtimes.PHPPortForVersion(minor),
					})
				}
			}
		}
	}
	return all
}

// CheckRuntimeUpdates inspects available version upgrades and security releases for managed runtimes
func (a *App) CheckRuntimeUpdates() []updater.RuntimeUpdateInfo {
	ctx := context.Background()
	if a.ctx != nil {
		ctx = a.ctx
	}
	opts := updater.CheckOptions{
		PHPReleasesURL: a.phpReleasesURL,
		BaseBinDir:     a.getBaseBinDir(),
		DataDir:        a.getDataDir(),
		LogDir:         filepath.Join(a.getDataDir(), "logs"),
		Registry:       a.getRegistry(),
	}
	return updater.CheckUpdates(ctx, opts)
}

// UpgradeRuntime stops running service if active, downloads latest release, preserves configs, and restarts service.
// Upgrades with MajorChange=true must NOT run from the normal Upgrade call:
// UpgradeRuntime(name) returns a clear error "major_change_requires_confirmation".
func (a *App) UpgradeRuntime(name string) (UIMessage, error) {
	return a.upgradeRuntimeInternal(name, false)
}

// UpgradeRuntimeConfirmed performs the runtime upgrade explicitly even if MajorChange=true.
func (a *App) UpgradeRuntimeConfirmed(name string) (UIMessage, error) {
	return a.upgradeRuntimeInternal(name, true)
}

func (a *App) upgradeRuntimeInternal(name string, confirmed bool) (UIMessage, error) {
	url, checksum, algo, _, majorChange, err := a.resolveArtifactInfo(name)
	if err != nil {
		return UIMessage{}, err
	}

	if majorChange && !confirmed {
		return UIMessage{}, fmt.Errorf("major_change_requires_confirmation")
	}

	family := updater.RuntimeFamily(runtimes.NormalizeName(name))
	serviceName := name
	if family == "mysql" {
		serviceName = "mysql"
	}

	wasRunning := false
	if a.supervisor != nil {
		if status, exists := a.supervisor.GetServiceStatus(serviceName); exists && status.Running {
			wasRunning = true
			_ = a.supervisor.StopService(serviceName)
		}
	}

	msg, err := a.downloadAndInstallArtifact(name, url, checksum, algo)
	if err != nil {
		if wasRunning {
			_, _ = a.ToggleService(serviceName)
		}
		return UIMessage{}, err
	}

	if family == "mysql" && majorChange {
		a.markMySQLPendingUpgrade()
	}

	if wasRunning {
		_, _ = a.ToggleService(serviceName)
	}

	return msg, nil
}


func (a *App) CreateProject(name string, projectType string, phpVersion string) error {
	projectDir := filepath.Join(a.cfg.ProjectsPath, name)
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return err
	}

	domain := name + a.cfg.DomainSuffix
	docRoot := projectDir

	if projectType == "laravel" {
		docRoot = filepath.Join(projectDir, "public")
		_ = os.MkdirAll(docRoot, 0755)

		envContent := `APP_NAME=Laravel
APP_ENV=local
APP_KEY=
APP_DEBUG=true
`
		cfg := projects.EnvConfig{
			AppURL:   fmt.Sprintf("https://%s", domain),
			DBName:   strings.ReplaceAll(name, "-", "_"),
			DBUser:   "root",
			DBPass:   "",
			MailHost: "127.0.0.1",
			MailPort: "1025",
		}
		injected := projects.InjectEnvVariables(envContent, cfg)
		_ = os.WriteFile(filepath.Join(projectDir, ".env"), []byte(injected), 0644)

		indexPHP := `<?php echo "<h1>LaraGal - Laravel App: ` + name + `</h1><p>Trusted SSL: https://` + domain + `</p>";`
		_ = os.WriteFile(filepath.Join(docRoot, "index.php"), []byte(indexPHP), 0644)
	} else {
		indexPHP := `<?php echo "<h1>LaraGal - Site: ` + name + `</h1><p>PHP Version: ` + phpVersion + `</p>";`
		_ = os.WriteFile(filepath.Join(docRoot, "index.php"), []byte(indexPHP), 0644)
	}

	return a.AddSite(domain, docRoot, phpVersion)
}

func (a *App) OpenFolder(folderPath string) error {
	cleanPath := filepath.FromSlash(folderPath)
	_ = os.MkdirAll(cleanPath, 0755)
	return winutil.Command("explorer", cleanPath).Start()
}

// waitForPort waits up to timeout (polling every 100ms) for the port to reach wantListening state.
// It returns the real final state (true if the port matches wantListening, false otherwise).
func (a *App) waitForPort(port int, wantListening bool, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for {
		if a.IsPortListening(port) == wantListening {
			return true
		}
		if time.Now().After(deadline) {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	return a.IsPortListening(port) == wantListening
}

func (a *App) IsPortListening(port int) bool {
	if port <= 0 {
		return false
	}
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 200*time.Millisecond)
	if err == nil {
		_ = conn.Close()
		return true
	}
	return false
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func findExecutableCandidate(dir string, name string) string {
	candidates := []string{
		filepath.Join(dir, name+".exe"),
		filepath.Join(dir, "server.exe"),
		filepath.Join(dir, "main.exe"),
		filepath.Join(dir, "app.exe"),
		filepath.Join(dir, "target", "release", name+".exe"),
		filepath.Join(dir, "target", "debug", name+".exe"),
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return ""
}

func (a *App) StartProjectApp(projectPath string, port int) (bool, error) {
	cleanPath := filepath.Clean(filepath.FromSlash(projectPath))
	projectName := filepath.Base(cleanPath)

	if a.IsPortListening(port) {
		return true, nil
	}

	// 1. Look for executable in project directory or build targets
	if targetExe := findExecutableCandidate(cleanPath, projectName); targetExe != "" {
		err := a.supervisor.StartServiceInDir(projectName, targetExe, []string{}, cleanPath, port)
		if err != nil {
			return false, err
		}
		// Wait up to 2 seconds for port to open
		for i := 0; i < 10; i++ {
			time.Sleep(200 * time.Millisecond)
			if a.IsPortListening(port) {
				return true, nil
			}
		}
		return true, nil
	}

	// 2. Check if Go source is available to run
	if _, err := os.Stat(filepath.Join(cleanPath, "go.mod")); err == nil {
		var runTarget string
		if _, err := os.Stat(filepath.Join(cleanPath, "cmd", "server")); err == nil {
			runTarget = "./cmd/server"
		} else {
			runTarget = "."
		}
		err := a.supervisor.StartServiceInDir(projectName, "go", []string{"run", runTarget}, cleanPath, port)
		if err != nil {
			return false, err
		}
		for i := 0; i < 15; i++ {
			time.Sleep(300 * time.Millisecond)
			if a.IsPortListening(port) {
				return true, nil
			}
		}
		return true, nil
	}

	// 3. Check for Rust application (cargo run --release)
	if _, err := os.Stat(filepath.Join(cleanPath, "Cargo.toml")); err == nil {
		err := a.supervisor.StartServiceInDir(projectName, "cargo", []string{"run", "--release"}, cleanPath, port)
		if err != nil {
			return false, err
		}
		for i := 0; i < 15; i++ {
			time.Sleep(300 * time.Millisecond)
			if a.IsPortListening(port) {
				return true, nil
			}
		}
		return true, nil
	}

	// 4. Check for .NET application (dotnet run)
	hasDotNet := false
	if entries, err := os.ReadDir(cleanPath); err == nil {
		for _, e := range entries {
			if strings.HasSuffix(strings.ToLower(e.Name()), ".csproj") || strings.HasSuffix(strings.ToLower(e.Name()), ".fsproj") {
				hasDotNet = true
				break
			}
		}
	}
	if hasDotNet || fileExists(filepath.Join(cleanPath, "Program.cs")) {
		err := a.supervisor.StartServiceInDir(projectName, "dotnet", []string{"run"}, cleanPath, port)
		if err != nil {
			return false, err
		}
		for i := 0; i < 15; i++ {
			time.Sleep(300 * time.Millisecond)
			if a.IsPortListening(port) {
				return true, nil
			}
		}
		return true, nil
	}

	// 5. Check for Python application (manage.py, main.py, app.py)
	if _, err := os.Stat(filepath.Join(cleanPath, "manage.py")); err == nil {
		portArg := fmt.Sprintf("%d", port)
		err := a.supervisor.StartServiceInDir(projectName, "python", []string{"manage.py", "runserver", portArg}, cleanPath, port)
		if err != nil {
			return false, err
		}
		for i := 0; i < 15; i++ {
			time.Sleep(300 * time.Millisecond)
			if a.IsPortListening(port) {
				return true, nil
			}
		}
		return true, nil
	}
	if _, err := os.Stat(filepath.Join(cleanPath, "main.py")); err == nil {
		err := a.supervisor.StartServiceInDir(projectName, "python", []string{"main.py"}, cleanPath, port)
		if err != nil {
			return false, err
		}
		for i := 0; i < 15; i++ {
			time.Sleep(300 * time.Millisecond)
			if a.IsPortListening(port) {
				return true, nil
			}
		}
		return true, nil
	}
	if _, err := os.Stat(filepath.Join(cleanPath, "app.py")); err == nil {
		err := a.supervisor.StartServiceInDir(projectName, "python", []string{"app.py"}, cleanPath, port)
		if err != nil {
			return false, err
		}
		for i := 0; i < 15; i++ {
			time.Sleep(300 * time.Millisecond)
			if a.IsPortListening(port) {
				return true, nil
			}
		}
		return true, nil
	}

	// 6. Check for Ruby application (bundle exec rails s or ruby app.rb)
	if _, err := os.Stat(filepath.Join(cleanPath, "Gemfile")); err == nil {
		portArg := fmt.Sprintf("%d", port)
		err := a.supervisor.StartServiceInDir(projectName, "bundle", []string{"exec", "rails", "s", "-p", portArg}, cleanPath, port)
		if err != nil {
			return false, err
		}
		for i := 0; i < 15; i++ {
			time.Sleep(300 * time.Millisecond)
			if a.IsPortListening(port) {
				return true, nil
			}
		}
		return true, nil
	}

	// 7. Check for Node.js application (package.json in project root or subservices)
	nodeDirs := []string{cleanPath}
	if entries, err := os.ReadDir(cleanPath); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				nodeDirs = append(nodeDirs, filepath.Join(cleanPath, e.Name()))
				if e.Name() == "services" || e.Name() == "apps" {
					if subEntries, err := os.ReadDir(filepath.Join(cleanPath, e.Name())); err == nil {
						for _, se := range subEntries {
							if se.IsDir() {
								nodeDirs = append(nodeDirs, filepath.Join(cleanPath, e.Name(), se.Name()))
							}
						}
					}
				}
			}
		}
	}
	for _, nDir := range nodeDirs {
		pkgFile := filepath.Join(nDir, "package.json")
		if _, err := os.Stat(pkgFile); err == nil {
			var cmdArgs []string
			if _, err := os.Stat(filepath.Join(nDir, "start-server.bat")); err == nil {
				cmdArgs = []string{"/c", filepath.Join(nDir, "start-server.bat")}
			} else {
				scriptToRun := "dev"
				if data, err := os.ReadFile(pkgFile); err == nil {
					content := string(data)
					if strings.Contains(content, "\"start\"") && !strings.Contains(content, "\"dev\"") {
						scriptToRun = "start"
					}
				}
				cmdArgs = []string{"/c", "npm", "run", scriptToRun}
			}
			err := a.supervisor.StartServiceInDir(projectName, "cmd", cmdArgs, nDir, port)
			if err != nil {
				return false, err
			}
			for i := 0; i < 20; i++ {
				time.Sleep(300 * time.Millisecond)
				if a.IsPortListening(port) {
					return true, nil
				}
			}
			return true, nil
		}
	}

	return false, fmt.Errorf("no executable or start script found for project %s", projectName)
}

func (a *App) StopProjectApp(projectName string) error {
	return a.supervisor.StopService(projectName)
}

func (a *App) OpenBrowser(targetURL string) error {
	cleanURL := strings.ToLower(targetURL)
	for _, s := range a.sites {
		if s.Domain != "" && strings.Contains(cleanURL, strings.ToLower(s.Domain)) {
			if (s.ProjectType == "go" || s.ProjectType == "proxy" || s.ProxyPort > 0) && s.ProxyPort > 0 {
				if !a.IsPortListening(s.ProxyPort) {
					_, _ = a.StartProjectApp(s.Path, s.ProxyPort)
				}
			}
			break
		}
	}

	return winutil.Command("rundll32", "url.dll,FileProtocolHandler", targetURL).Start()
}

func (a *App) OpenVSCode(folderPath string) error {
	cleanPath := filepath.FromSlash(folderPath)
	_ = os.MkdirAll(cleanPath, 0755)
	return winutil.Command("cmd", "/c", "code", cleanPath).Start()
}

func (a *App) MinimizeWindow() {
	if a.ctx != nil {
		runtime.WindowMinimise(a.ctx)
	}
}

func (a *App) ToggleMaximizeWindow() {
	if a.ctx != nil {
		runtime.WindowToggleMaximise(a.ctx)
	}
}

// CloseWindow hides the window to the background so running services (Caddy, PHP, MySQL, Mailpit) keep working
func (a *App) CloseWindow() {
	if a.ctx != nil {
		runtime.WindowHide(a.ctx)
	}
}

// QuitApp performs full application shutdown and stops background services
func (a *App) QuitApp() {
	if a.ctx != nil {
		runtime.Quit(a.ctx)
	}
}

// GetProjectsDirectory returns the configured projects root directory
func (a *App) GetProjectsDirectory() string {
	return a.cfg.ProjectsPath
}

// GetRecentDirectories returns the list of recently used project folders
func (a *App) GetRecentDirectories() []string {
	if len(a.cfg.RecentPaths) == 0 && a.cfg.ProjectsPath != "" {
		a.cfg.RecentPaths = []string{a.cfg.ProjectsPath}
	}
	return a.cfg.RecentPaths
}

func (a *App) addRecentDirectory(dir string) {
	cleanDir := filepath.Clean(dir)
	newRecent := []string{cleanDir}
	for _, p := range a.cfg.RecentPaths {
		if filepath.Clean(p) != cleanDir {
			newRecent = append(newRecent, p)
		}
		if len(newRecent) >= 5 {
			break
		}
	}
	a.cfg.RecentPaths = newRecent
}

// SetProjectsDirectory switches active projects directory and rescans projects
func (a *App) SetProjectsDirectory(dirPath string) ([]caddy.SiteDef, error) {
	cleanDir := filepath.Clean(dirPath)
	if err := os.MkdirAll(cleanDir, 0755); err != nil {
		return a.sites, err
	}

	a.cfg.ProjectsPath = cleanDir
	a.addRecentDirectory(cleanDir)
	_ = a.cfg.Save(a.cfgPath)

	sites := a.ScanProjects()
	return sites, nil
}

// SelectProjectsDirectory displays a directory chooser dialog with guaranteed PowerShell fallback and updates configuration
func (a *App) SelectProjectsDirectory(lang string) (string, error) {
	defaultDir := a.cfg.ProjectsPath
	if defaultDir == "" {
		if home, err := os.UserHomeDir(); err == nil && home != "" {
			defaultDir = filepath.Join(home, "laragal", "sites")
		} else {
			defaultDir = "C:\\laragal\\sites"
		}
	}
	_ = os.MkdirAll(defaultDir, 0755)

	// In Wails runtime, if DefaultDirectory is non-empty and does NOT exist on disk,
	// OpenDirectoryDialog fails with "default directory '...' does not exist".
	// Passing an empty string bypasses Wails validation completely.
	wailsDir := defaultDir
	if _, err := os.Stat(wailsDir); err != nil {
		wailsDir = ""
	}

	dialogTitle := "Select Projects Directory"
	if strings.ToLower(strings.TrimSpace(lang)) == "en" {
		dialogTitle = "Select Projects Directory"
	}

	var selectedDir string
	var err error

	if a.ctx != nil {
		selectedDir, err = runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
			DefaultDirectory: wailsDir,
			Title:            dialogTitle,
		})
	}

	// Only trigger PowerShell fallback if the native COM dialog genuinely failed with an error.
	// If the user simply cancelled or closed the window (err == nil && selectedDir == ""), do not pop up fallback.
	if err != nil {
		safeDir := defaultDir
		if _, statErr := os.Stat(safeDir); statErr != nil {
			safeDir = ""
		}
		var psScript string
		if safeDir != "" {
			psScript = fmt.Sprintf(`Add-Type -AssemblyName System.Windows.Forms; $f = New-Object System.Windows.Forms.FolderBrowserDialog; $f.Description = '%s'; $f.SelectedPath = '%s'; if ($f.ShowDialog() -eq [System.Windows.Forms.DialogResult]::OK) { Write-Output $f.SelectedPath }`, dialogTitle, strings.ReplaceAll(safeDir, `\`, `\\`))
		} else {
			psScript = fmt.Sprintf(`Add-Type -AssemblyName System.Windows.Forms; $f = New-Object System.Windows.Forms.FolderBrowserDialog; $f.Description = '%s'; if ($f.ShowDialog() -eq [System.Windows.Forms.DialogResult]::OK) { Write-Output $f.SelectedPath }`, dialogTitle)
		}
		out, psErr := winutil.Command("powershell", "-NoProfile", "-Command", psScript).Output()
		if psErr == nil {
			trimmed := strings.TrimSpace(string(out))
			if trimmed != "" {
				selectedDir = trimmed
			}
		}
	}

	if selectedDir == "" {
		return "", nil
	}

	cleanDir := filepath.Clean(selectedDir)
	_ = os.MkdirAll(cleanDir, 0755)
	a.cfg.ProjectsPath = cleanDir
	a.addRecentDirectory(cleanDir)
	_ = a.cfg.Save(a.cfgPath)

	a.ScanProjects()
	return a.cfg.ProjectsPath, nil
}

// SetProjectPHPVersion updates the PHP version for a specific project, ensures the PHP FastCGI service is running, and reloads Caddy
func (a *App) SetProjectPHPVersion(projectPathOrDomain string, version string) ([]caddy.SiteDef, error) {
	normVer := projects.NormalizePHPVersion(version)
	cleanKey := filepath.ToSlash(filepath.Clean(projectPathOrDomain))

	if a.cfg.ProjectPHP == nil {
		a.cfg.ProjectPHP = make(map[string]string)
	}
	a.cfg.ProjectPHP[cleanKey] = normVer
	_ = a.cfg.Save(a.cfgPath)

	// Ensure the corresponding PHP FastCGI service is running
	_ = a.ensurePHPServiceRunning(normVer)

	// Rescan projects to update ports and reload Caddy
	sites := a.ScanProjects()
	return sites, nil
}

// SaveProjectSettings updates manual project overrides (type, docroot, proxy_port, php_version, is_ignored)
func (a *App) SaveProjectSettings(projectPath string, settings config.ProjectCustomSettings) ([]caddy.SiteDef, error) {
	cleanKey := filepath.ToSlash(filepath.Clean(projectPath))

	if a.cfg.ProjectOverrides == nil {
		a.cfg.ProjectOverrides = make(map[string]config.ProjectCustomSettings)
	}
	a.cfg.ProjectOverrides[cleanKey] = settings

	if settings.PHPVersion != "" {
		if a.cfg.ProjectPHP == nil {
			a.cfg.ProjectPHP = make(map[string]string)
		}
		a.cfg.ProjectPHP[cleanKey] = projects.NormalizePHPVersion(settings.PHPVersion)
		_ = a.ensurePHPServiceRunning(settings.PHPVersion)
	}

	_ = a.cfg.Save(a.cfgPath)

	sites := a.ScanProjects()
	return sites, nil
}

type ProjectDiagnostics struct {
	ProjectPath    string      `json:"project_path"`
	Domain         string      `json:"domain"`
	ProjectType    string      `json:"project_type"`
	PHPVersion     string      `json:"php_version"`
	PHPPort        int         `json:"php_port"`
	ProxyPort      int         `json:"proxy_port"`
	DocRoot        string      `json:"doc_root"`
	DocRootExists  bool        `json:"doc_root_exists"`
	HasIndexFile   bool        `json:"has_index_file"`
	IndexFile      string      `json:"index_file,omitempty"`
	HasEnvFile     bool        `json:"has_env_file"`
	HasVendorDir   bool        `json:"has_vendor_dir"`
	HasNodeModules bool        `json:"has_node_modules"`
	IsPortActive   bool        `json:"is_port_active"`
	Status         string      `json:"status"` // "healthy", "warning", "error"
	Issues         []UIMessage `json:"issues"`
	Suggestions    []UIMessage `json:"suggestions"`
}

// ScanProjectDatabases searches for preexisting schemas (.sql) and local MySQL databases for the given project
func (a *App) ScanProjectDatabases(projectPath string) ([]mysql.DiscoveredDatabase, error) {
	return mysql.ScanAllDatabases(projectPath)
}

// ImportProjectDatabase imports a discovered schema or legacy database and syncs the project's .env
func (a *App) ImportProjectDatabase(projectPath string, candidate mysql.DiscoveredDatabase) (bool, error) {
	err := mysql.ImportDiscoveredDatabase(projectPath, candidate)
	if err != nil {
		return false, err
	}
	return true, nil
}

// DiagnoseProject analyzes a project configuration, checks disk files, ports, and env
func (a *App) DiagnoseProject(projectPath string) ProjectDiagnostics {
	cleanPath := filepath.Clean(projectPath)
	pName := filepath.Base(cleanPath)

	var matchedSite *caddy.SiteDef
	for i := range a.sites {
		if filepath.Clean(a.sites[i].Path) == cleanPath || filepath.Clean(a.sites[i].RootPath) == cleanPath || a.sites[i].Name == pName {
			matchedSite = &a.sites[i]
			break
		}
	}

	pType := "php"
	domain := pName + a.cfg.DomainSuffix
	docRoot := cleanPath
	phpPort := runtimes.PHPPortForVersion(a.cfg.DefaultPHP)
	phpVer := a.cfg.DefaultPHP
	proxyPort := 0

	if matchedSite != nil {
		pType = matchedSite.ProjectType
		domain = matchedSite.Domain
		docRoot = filepath.Clean(matchedSite.RootPath)
		phpPort = matchedSite.PHPPort
		phpVer = matchedSite.PHPVersion
		proxyPort = matchedSite.ProxyPort
	}

	diag := ProjectDiagnostics{
		ProjectPath: filepath.ToSlash(cleanPath),
		Domain:      domain,
		ProjectType: pType,
		PHPVersion:  phpVer,
		PHPPort:     phpPort,
		ProxyPort:   proxyPort,
		DocRoot:     filepath.ToSlash(docRoot),
		Status:      "healthy",
		Issues:      make([]UIMessage, 0),
		Suggestions: make([]UIMessage, 0),
	}

	// 1. Check Document Root
	if _, err := os.Stat(docRoot); err == nil {
		diag.DocRootExists = true
	} else {
		diag.DocRootExists = false
		diag.Status = "error"
		diag.Issues = append(diag.Issues, UIMessage{Key: "diag_issue_docroot_missing", Params: map[string]string{"base": filepath.Base(docRoot)}})
		diag.Suggestions = append(diag.Suggestions, UIMessage{Key: "diag_sug_docroot_fix"})
	}

	// Check index file in docRoot
	indexCandidates := []string{"index.php", "index.html", "index.htm"}
	for _, idx := range indexCandidates {
		if fileExists(filepath.Join(docRoot, idx)) {
			diag.HasIndexFile = true
			diag.IndexFile = idx
			break
		}
	}
	if !diag.HasIndexFile && pType != "go" && pType != "rust" && pType != "dotnet" && pType != "python" {
		diag.Status = "warning"
		diag.Issues = append(diag.Issues, UIMessage{Key: "diag_issue_index_missing"})
		diag.Suggestions = append(diag.Suggestions, UIMessage{Key: "diag_sug_index_fix"})
	}

	// 2. Check PHP projects (.env, vendor, fastcgi)
	if pType == "laravel" || pType == "wordpress" || pType == "php" {
		envPath := filepath.Join(cleanPath, ".env")
		if fileExists(envPath) {
			diag.HasEnvFile = true
		} else {
			diag.HasEnvFile = false
			if pType == "laravel" {
				diag.Status = "warning"
				diag.Issues = append(diag.Issues, UIMessage{Key: "diag_issue_env_missing"})
				if fileExists(filepath.Join(cleanPath, ".env.example")) {
					diag.Suggestions = append(diag.Suggestions, UIMessage{Key: "diag_sug_env_create"})
				} else {
					diag.Suggestions = append(diag.Suggestions, UIMessage{Key: "diag_sug_env_manual"})
				}
			}
		}

		vendorPath := filepath.Join(cleanPath, "vendor", "autoload.php")
		if fileExists(vendorPath) {
			diag.HasVendorDir = true
		} else {
			diag.HasVendorDir = false
			if fileExists(filepath.Join(cleanPath, "composer.json")) {
				diag.Status = "warning"
				diag.Issues = append(diag.Issues, UIMessage{Key: "diag_issue_vendor_missing"})
				diag.Suggestions = append(diag.Suggestions, UIMessage{Key: "diag_sug_composer_install"})
			}
		}

		if phpPort > 0 {
			if a.IsPortListening(phpPort) {
				diag.IsPortActive = true
			} else {
				diag.IsPortActive = false
				diag.Status = "error"
				diag.Issues = append(diag.Issues, UIMessage{Key: "diag_issue_php_port_inactive", Params: map[string]string{"port": fmt.Sprintf("%d", phpPort)}})
				diag.Suggestions = append(diag.Suggestions, UIMessage{Key: "diag_sug_php_service"})
			}
		}
	}

	// 3. Check Proxy projects (Go, Rust, .NET, Python, Node)
	if proxyPort > 0 {
		if a.IsPortListening(proxyPort) {
			diag.IsPortActive = true
		} else {
			diag.IsPortActive = false
			if diag.Status != "error" {
				diag.Status = "warning"
			}
			diag.Issues = append(diag.Issues, UIMessage{Key: "diag_issue_proxy_port_inactive", Params: map[string]string{"port": fmt.Sprintf("%d", proxyPort)}})
			diag.Suggestions = append(diag.Suggestions, UIMessage{Key: "diag_sug_start_app"})
		}
	}

	// 4. Node projects (node_modules)
	if pType == "node" {
		if dirExists(filepath.Join(cleanPath, "node_modules")) {
			diag.HasNodeModules = true
		} else if fileExists(filepath.Join(cleanPath, "package.json")) {
			diag.HasNodeModules = false
			diag.Status = "warning"
			diag.Issues = append(diag.Issues, UIMessage{Key: "diag_issue_node_modules_missing"})
			diag.Suggestions = append(diag.Suggestions, UIMessage{Key: "diag_sug_npm_install"})
		}
	}

	return diag
}

// FixProjectEnv creates .env from .env.example if missing
func (a *App) FixProjectEnv(projectPath string) (bool, error) {
	cleanPath := filepath.Clean(projectPath)
	envPath := filepath.Join(cleanPath, ".env")
	examplePath := filepath.Join(cleanPath, ".env.example")

	if fileExists(envPath) {
		return true, nil
	}

	if !fileExists(examplePath) {
		return false, fmt.Errorf(".env.example not found in project directory")
	}

	data, err := os.ReadFile(examplePath)
	if err != nil {
		return false, err
	}

	if err := os.WriteFile(envPath, data, 0644); err != nil {
		return false, err
	}

	return true, nil
}

// ScanProjects scans the projects directory, identifies project types, and populates existing projects
func (a *App) ScanProjects() []caddy.SiteDef {
	if a.cfg.ProjectsPath == "" {
		return a.sites
	}
	_ = os.MkdirAll(a.cfg.ProjectsPath, 0755)

	entries, err := os.ReadDir(a.cfg.ProjectsPath)
	if err != nil {
		return a.sites
	}

	scannedSites := make([]caddy.SiteDef, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		projectName := entry.Name()
		if strings.HasPrefix(projectName, ".") || strings.HasPrefix(projectName, "$") {
			continue
		}

		projectDir := filepath.Join(a.cfg.ProjectsPath, projectName)
		detected := projects.DetectProject(projectDir, a.cfg.DefaultPHP, a.cfg.DomainSuffix)

		cleanKey := filepath.ToSlash(projectDir)
		isWeb := detected.IsWeb
		projectType := string(detected.Type)
		typeLabel := detected.TypeLabel
		docRoot := detected.DocRoot
		phpVer := detected.PHPVersion
		phpPort := detected.PHPPort
		proxyPort := detected.ProxyPort

		// 1. Check if user configured manual overrides for this project
		if a.cfg.ProjectOverrides != nil {
			var override config.ProjectCustomSettings
			var hasOverride bool
			if ov, ok := a.cfg.ProjectOverrides[cleanKey]; ok {
				override = ov
				hasOverride = true
			} else if ov, ok := a.cfg.ProjectOverrides[projectName]; ok {
				override = ov
				hasOverride = true
			} else if ov, ok := a.cfg.ProjectOverrides[detected.Domain]; ok {
				override = ov
				hasOverride = true
			}

			if hasOverride {
				if override.IsIgnored || override.Type == "non_web" {
					isWeb = false
					projectType = "non_web"
					typeLabel = "Non-Web"
				} else {
					if override.Type != "" {
						projectType = override.Type
						isWeb = true
						switch override.Type {
						case "laravel":
							typeLabel = "Laravel"
						case "wordpress":
							typeLabel = "WordPress"
						case "php":
							typeLabel = "PHP"
						case "go":
							typeLabel = "Go"
						case "rust":
							typeLabel = "Rust"
						case "dotnet":
							typeLabel = ".NET"
						case "python":
							typeLabel = "Python"
						case "node":
							typeLabel = "Node.js"
						case "static":
							typeLabel = "Static"
						default:
							typeLabel = override.Type
						}
					}
					if override.DocRoot != "" {
						if filepath.IsAbs(override.DocRoot) {
							docRoot = filepath.Clean(override.DocRoot)
						} else {
							docRoot = filepath.Join(projectDir, override.DocRoot)
						}
					}
					if override.ProxyPort > 0 {
						proxyPort = override.ProxyPort
					}
					if override.PHPVersion != "" {
						phpVer = runtimes.NormalizePHPVersion(override.PHPVersion)
						phpPort = runtimes.PHPPortForVersion(phpVer)
					}
				}
			}
		}

		// 2. Standalone ProjectPHP overrides fallback
		if overrideVer, exists := a.cfg.ProjectPHP[cleanKey]; exists && overrideVer != "" {
			phpVer = runtimes.NormalizePHPVersion(overrideVer)
			phpPort = runtimes.PHPPortForVersion(phpVer)
		} else if overrideVer, exists := a.cfg.ProjectPHP[projectName]; exists && overrideVer != "" {
			phpVer = runtimes.NormalizePHPVersion(overrideVer)
			phpPort = runtimes.PHPPortForVersion(phpVer)
		} else if overrideVer, exists := a.cfg.ProjectPHP[detected.Domain]; exists && overrideVer != "" {
			phpVer = runtimes.NormalizePHPVersion(overrideVer)
			phpPort = runtimes.PHPPortForVersion(phpVer)
		}

		scannedSites = append(scannedSites, caddy.SiteDef{
			Name:        detected.Name,
			Domain:      detected.Domain,
			Path:        filepath.ToSlash(detected.Path),
			RootPath:    filepath.ToSlash(docRoot),
			PHPPort:     phpPort,
			PHPVersion:  phpVer,
			ProxyPort:   proxyPort,
			ProjectType: projectType,
			TypeLabel:   typeLabel,
			IsWeb:       isWeb,
		})
	}

	a.sitesMu.Lock()
	a.sites = scannedSites
	a.sitesMu.Unlock()

	// Write Caddyfile synchronously so callers (like StartServices) always find a valid Caddyfile
	_ = a.ReloadCaddy()

	// Ensure needed PHP FastCGI instances asynchronously in background
	go func(sites []caddy.SiteDef) {
		for i := range sites {
			s := &sites[i]
			if s.IsWeb && s.PHPPort > 0 && s.PHPVersion != "" {
				if a.findPHPExecutableForVersion(s.PHPVersion) == "" {
					s.PHPVersion = runtimes.NormalizePHPVersion(a.cfg.DefaultPHP)
					s.PHPPort = runtimes.PHPPortForVersion(s.PHPVersion)
				}
				_ = a.ensurePHPServiceRunning(s.PHPVersion)
			}
		}
	}(scannedSites)

	// 2. Update system hosts file asynchronously in background (prevents ipconfig /flushdns UI freezes)
	webDomains := make([]string, 0)
	for _, s := range scannedSites {
		if s.IsWeb && s.Domain != "" {
			webDomains = append(webDomains, s.Domain)
		}
	}
	suffix := ".test"
	if a.cfg != nil && a.cfg.DomainSuffix != "" {
		suffix = a.cfg.DomainSuffix
	}
	webDomains = append(webDomains, "adminer"+suffix)
	webDomains = append(webDomains, "phpmyadmin"+suffix)
	go func(domains []string) {
		_ = hosts.UpdateSystemHosts(domains)
	}(webDomains)

	return scannedSites
}

func (a *App) writeAtomicFile(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, perm); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	return nil
}

// ReloadCaddy writes the updated Caddyfile configuration and reloads Caddy if running
func (a *App) ReloadCaddy() error {
	caddyDir := filepath.Join("data", "caddy")
	_ = os.MkdirAll(caddyDir, 0755)
	caddyfilePath := filepath.Join(caddyDir, "Caddyfile")
	absAdminerPath, _ := filepath.Abs(filepath.Join("data", "adminer"))
	adminerDomain := "adminer.test"
	if a.cfg != nil && a.cfg.DomainSuffix != "" {
		adminerDomain = "adminer" + a.cfg.DomainSuffix
	}
	adminerCfg := caddy.AdminerConfig{
		Enabled:  true,
		Domain:   adminerDomain,
		RootPath: absAdminerPath,
		PHPPort:  a.getActivePHPPort(),
		Port:     3307,
	}
	pmaCfg := caddy.PhpMyAdminConfig{}
	if phpmyadmin.IsInstalled(".") {
		_ = phpmyadmin.EnsureConfig(".")
		absPmaPath, _ := filepath.Abs(filepath.Join("data", "phpmyadmin"))
		pmaDomain := "phpmyadmin.test"
		if a.cfg != nil && a.cfg.DomainSuffix != "" {
			pmaDomain = "phpmyadmin" + a.cfg.DomainSuffix
		}
		pmaCfg = caddy.PhpMyAdminConfig{
			Enabled:  true,
			Domain:   pmaDomain,
			RootPath: absPmaPath,
			Port:     phpmyadmin.DefaultPort,
			PHPPort:  a.getActivePHPPort(),
		}
	}
	a.sitesMu.RLock()
	sites := a.sites
	a.sitesMu.RUnlock()
	caddyContent := caddy.GenerateCaddyfileWithTools(sites, 2019, adminerCfg, pmaCfg)

	if err := a.writeAtomicFile(caddyfilePath, []byte(caddyContent), 0644); err != nil {
		return err
	}

	// Also write to data/Caddyfile for standard root access
	_ = a.writeAtomicFile(filepath.Join("data", "Caddyfile"), []byte(caddyContent), 0644)

	// Attempt live reload via Caddy admin API asynchronously so it never delays UI or caller
	go func(content string) {
		req, err := http.NewRequest("POST", "http://localhost:2019/load", strings.NewReader(content))
		if err == nil {
			req.Header.Set("Content-Type", "text/caddyfile")
			client := &http.Client{Timeout: 500 * time.Millisecond}
			resp, postErr := client.Do(req)
			if postErr == nil {
				_ = resp.Body.Close()
			}
		}
	}(caddyContent)

	return nil
}

// SyncHosts synchronizes the active web domains to the Windows hosts file and flushes DNS cache
func (a *App) SyncHosts() (UIMessage, error) {
	webDomains := make([]string, 0)
	for _, s := range a.sites {
		if s.IsWeb && s.Domain != "" {
			webDomains = append(webDomains, s.Domain)
		}
	}
	suffix := ".test"
	if a.cfg != nil && a.cfg.DomainSuffix != "" {
		suffix = a.cfg.DomainSuffix
	}
	webDomains = append(webDomains, "adminer"+suffix)
	webDomains = append(webDomains, "phpmyadmin"+suffix)
	if err := hosts.UpdateSystemHosts(webDomains); err != nil {
		return UIMessage{}, err
	}
	if len(webDomains) == 0 {
		return UIMessage{Key: "hosts_clean_no_sites"}, nil
	}

	return UIMessage{
		Key:    "hosts_synced_count",
		Params: map[string]string{"count": fmt.Sprintf("%d", len(webDomains))},
	}, nil
}

// getActivePHPPort returns the port of the currently running PHP FastCGI instance, or default port
func (a *App) getActivePHPPort() int {
	if a.supervisor != nil {
		for _, rt := range runtimes.All() {
			if strings.HasPrefix(rt.Name, "php-") && rt.PHPPort > 0 {
				if st, ok := a.supervisor.GetServiceStatus(rt.Name); ok && st.Running {
					return rt.PHPPort
				}
			}
		}
	}
	defVer := "8.4"
	if a.cfg != nil && a.cfg.DefaultPHP != "" {
		defVer = a.cfg.DefaultPHP
	}
	return runtimes.PHPPortForVersion(defVer)
}

// EnsureAdminerInstalled verifies that data/adminer/index.php is present or writes embedded release
func (a *App) EnsureAdminerInstalled() error {
	_, err := adminer.EnsureInstalled(".")
	return err
}

// OpenAdminer starts prerequisites (MySQL, PHP, Caddy) and opens Adminer in browser
func (a *App) OpenAdminer(dbName string) error {
	_ = a.EnsureAdminerInstalled()

	// 1. Ensure MySQL is running
	mysqlStatus, _ := a.supervisor.GetServiceStatus("mysql")
	if !mysqlStatus.Running && !a.IsPortListening(3306) {
		_, _ = a.ToggleService("mysql")
	}

	// 2. Ensure at least one PHP FastCGI instance is running
	phpRunning := false
	if a.supervisor != nil {
		for _, st := range a.supervisor.GetAllServices() {
			if strings.HasPrefix(st.Name, "php-") && st.Running {
				phpRunning = true
				break
			}
		}
	}
	if !phpRunning {
		defVer := "8.4"
		if a.cfg != nil && a.cfg.DefaultPHP != "" {
			defVer = a.cfg.DefaultPHP
		}
		_ = a.ensurePHPServiceRunning(defVer)
	}

	// 3. Ensure Caddy is running
	caddyStatus, _ := a.supervisor.GetServiceStatus("caddy")
	if !caddyStatus.Running {
		_, _ = a.ToggleService("caddy")
	}

	_ = a.ReloadCaddy()

	targetURL := adminer.BuildURL("adminer"+a.cfg.DomainSuffix, 3307, dbName)
	return a.OpenBrowser(targetURL)
}

// OpenPhpMyAdmin starts prerequisites (MySQL, PHP, Caddy) and opens phpMyAdmin in browser
func (a *App) OpenPhpMyAdmin(dbName string) error {
	if !phpmyadmin.IsInstalled(".") {
		return fmt.Errorf("phpMyAdmin is not installed in data/phpmyadmin")
	}

	_ = phpmyadmin.EnsureConfig(".")

	if a.supervisor != nil {
		// 1. Ensure MySQL is running
		mysqlStatus, _ := a.supervisor.GetServiceStatus("mysql")
		if !mysqlStatus.Running && !a.IsPortListening(3306) {
			_, _ = a.ToggleService("mysql")
		}

		// 2. Ensure at least one PHP FastCGI instance is running
		phpRunning := false
		if a.supervisor != nil {
			for _, st := range a.supervisor.GetAllServices() {
				if strings.HasPrefix(st.Name, "php-") && st.Running {
					phpRunning = true
					break
				}
			}
		}
		if !phpRunning {
			defVer := "8.4"
			if a.cfg != nil && a.cfg.DefaultPHP != "" {
				defVer = a.cfg.DefaultPHP
			}
			_ = a.ensurePHPServiceRunning(defVer)
		}

		// 3. Ensure Caddy is running
		caddyStatus, _ := a.supervisor.GetServiceStatus("caddy")
		if !caddyStatus.Running {
			_, _ = a.ToggleService("caddy")
		}
	}

	_ = a.ReloadCaddy()

	suffix := ".test"
	if a.cfg != nil && a.cfg.DomainSuffix != "" {
		suffix = a.cfg.DomainSuffix
	}
	targetURL := phpmyadmin.BuildURL("phpmyadmin"+suffix, phpmyadmin.DefaultPort, dbName)
	return a.OpenBrowser(targetURL)
}