package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/laragal/laragal/internal/certs"
	"github.com/laragal/laragal/internal/config"
	"github.com/laragal/laragal/internal/downloader"
	"github.com/laragal/laragal/internal/hosts"
	"github.com/laragal/laragal/internal/privilege"
	"github.com/laragal/laragal/internal/process"
	"github.com/laragal/laragal/internal/projects"
	"github.com/laragal/laragal/internal/services/caddy"
	"github.com/laragal/laragal/internal/services/php"
	"github.com/laragal/laragal/internal/winutil"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx        context.Context
	cfg        *config.AppConfig
	supervisor *process.Supervisor
	sites      []caddy.SiteDef
	cfgPath    string
}

func NewApp() *App {
	cfgPath := "laragal.json"
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		cfg = config.NewDefaultConfig()
		_ = cfg.Save(cfgPath)
	}

	return &App{
		cfg:        cfg,
		supervisor: process.NewSupervisor(),
		sites:      make([]caddy.SiteDef, 0),
		cfgPath:    cfgPath,
	}
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
		a.sites = cached
	}
}

func (a *App) saveSitesCache() {
	if len(a.sites) == 0 {
		return
	}
	_ = os.MkdirAll("data", 0755)
	if data, err := json.MarshalIndent(a.sites, "", "  "); err == nil {
		_ = os.WriteFile(a.getCacheFilePath(), data, 0644)
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	_ = os.MkdirAll(a.cfg.ProjectsPath, 0755)
	a.initTray()

	// Load cached sites immediately so the UI displays existing projects with 0ms delay!
	a.loadSitesCache()

	// Launch background services, project scanning, and hosts synchronization asynchronously
	// so the Wails window and Vue frontend render immediately with zero delay!
	go func() {
		a.StartServices()
		a.ScanProjects()
		a.saveSitesCache()
		if a.ctx != nil {
			runtime.EventsEmit(a.ctx, "services-updated", a.GetServices())
			runtime.EventsEmit(a.ctx, "sites-updated", a.sites)
		}
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
	// 1. Start PHP FastCGI (port 9083)
	phpExe := a.findPHPExecutable()
	if phpExe != "" {
		_ = a.supervisor.StartService("php-8.3", phpExe, []string{"-b", "127.0.0.1:9083"}, 9083)
	}

	// 2. Start Caddy on 443
	caddyExe := a.findCaddyExecutable()
	if caddyExe != "" {
		caddyDir := filepath.Join("data", "caddy")
		_ = os.MkdirAll(caddyDir, 0755)
		caddyfilePath := filepath.Join(caddyDir, "Caddyfile")
		_ = a.supervisor.StartService("caddy", caddyExe, []string{"run", "--config", caddyfilePath, "--adapter", "caddyfile"}, 443)
	}

	// 3. Start Mailpit (if present)
	mailpitExe := filepath.Join("bin", "mailpit", "mailpit.exe")
	if _, err := os.Stat(mailpitExe); err == nil {
		_ = a.supervisor.StartService("mailpit", mailpitExe, []string{"--smtp-bind-addr", "127.0.0.1:1025", "--ui-bind-addr", "127.0.0.1:8025"}, 1025)
	}
}

func (a *App) ToggleService(name string) (bool, error) {
	status, exists := a.supervisor.GetServiceStatus(name)
	if exists && status.Running {
		err := a.supervisor.StopService(name)
		return false, err
	}

	switch name {
	case "caddy":
		caddyExe := a.findCaddyExecutable()
		if caddyExe == "" {
			return false, fmt.Errorf("خادم Caddy غير مثبت بعد. يرجى تنزيله من قسم الحزم المحمولة.")
		}
		caddyDir := filepath.Join("data", "caddy")
		_ = os.MkdirAll(caddyDir, 0755)
		caddyfilePath := filepath.Join(caddyDir, "Caddyfile")
		err := a.supervisor.StartService("caddy", caddyExe, []string{"run", "--config", caddyfilePath, "--adapter", "caddyfile"}, 443)
		return err == nil, err

	case "php-8.3", "php-8.4":
		phpExe := a.findPHPExecutable()
		if phpExe == "" {
			return false, fmt.Errorf("محرك PHP غير متوفر. يرجى تنزيله من قسم الحزم المحمولة.")
		}
		port := 9083
		if name == "php-8.4" {
			port = 9084
		}
		err := a.supervisor.StartService(name, phpExe, []string{"-b", fmt.Sprintf("127.0.0.1:%d", port)}, port)
		return err == nil, err

	case "mailpit":
		mailpitExe := filepath.Join("bin", "mailpit", "mailpit.exe")
		if _, err := os.Stat(mailpitExe); err != nil {
			return false, fmt.Errorf("صائد البريد Mailpit غير مثبت بعد.")
		}
		err := a.supervisor.StartService("mailpit", mailpitExe, []string{"--smtp-bind-addr", "127.0.0.1:1025", "--ui-bind-addr", "127.0.0.1:8025"}, 1025)
		return err == nil, err
	}

	return false, fmt.Errorf("الخدمة %s غير معروفة", name)
}

func (a *App) CheckPortConflict(port int) string {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 200*time.Millisecond)
	if err == nil {
		_ = conn.Close()
		for _, s := range a.GetServices() {
			if s.Running && s.Port == port {
				return ""
			}
		}
		return fmt.Sprintf("تنبيه: المنفذ %d مستخدم حالياً من قِبل برنامج آخر على جهازك (مثل خادم محلي قديم).", port)
	}
	return ""
}

func (a *App) GetServices() []process.ServiceStatus {
	serviceNames := []string{"caddy", "php-8.3", "mysql", "redis", "mailpit"}
	result := make([]process.ServiceStatus, 0, len(serviceNames))
	for _, name := range serviceNames {
		status, _ := a.supervisor.GetServiceStatus(name)
		result = append(result, status)
	}
	return result
}

func (a *App) GetSites() []caddy.SiteDef {
	return a.sites
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

func (a *App) InstallSSLTrust() (string, error) {
	rootCertPath := a.findCaddyRootCert()
	if rootCertPath == "" {
		return "لم يتم العثور على شهادة الجذر لـ Caddy بعد. يرجى تشغيل خادم Caddy أولاً لتوليدها.", nil
	}

	fingerprint, err := certs.ReadCertFingerprint(rootCertPath)
	if err != nil {
		return "", err
	}

	trusted, _ := certs.IsCertTrustedInWindowsStore(fingerprint)
	if trusted {
		return "شهادة الجذر المحلية موثوقة بالفعل في نظام ويندوز.", nil
	}

	// 1. Try native caddy trust if available
	caddyExe := a.findCaddyExecutable()
	if caddyExe != "" {
		_ = winutil.Command(caddyExe, "trust").Run()
		if trustedNow, _ := certs.IsCertTrustedInWindowsStore(fingerprint); trustedNow {
			return "تم تثبيت شهادة الجذر في متجر شهادات ويندوز بنجاح.", nil
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
			return string(out) + " " + string(elOut), elErr
		}
	}

	return "تم تثبيت شهادة الجذر في متجر شهادات ويندوز بنجاح.", nil
}

func (a *App) DownloadBinary(name string) (string, error) {
	url, err := downloader.GetDownloadURL(name)
	if err != nil {
		return "", err
	}

	tempDir := filepath.Join("data", "tmp")
	_ = os.MkdirAll(tempDir, 0755)
	zipPath := filepath.Join(tempDir, fmt.Sprintf("%s.zip", name))
	destDir := filepath.Join("bin", name)

	// Download zip
	resp, err := winutil.Command("powershell", "-NoProfile", "-Command", fmt.Sprintf("Invoke-WebRequest -Uri '%s' -OutFile '%s'", url, zipPath)).CombinedOutput()
	if err != nil {
		return string(resp), err
	}

	// Extract zip
	if err := downloader.ExtractZip(zipPath, destDir); err != nil {
		return "", err
	}

	_ = os.Remove(zipPath)
	return fmt.Sprintf("تم تنزيل وتثبيت %s بنجاح في %s", name, destDir), nil
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

	return false, fmt.Errorf("لم يتم العثور على ملف تنفيذي أو أداة تشغيل للمشروع %s", projectName)
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

	dialogTitle := "اختر مجلد المشاريع"
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

		scannedSites = append(scannedSites, caddy.SiteDef{
			Name:        detected.Name,
			Domain:      detected.Domain,
			Path:        filepath.ToSlash(detected.Path),
			RootPath:    filepath.ToSlash(detected.DocRoot),
			PHPPort:     detected.PHPPort,
			PHPVersion:  detected.PHPVersion,
			ProxyPort:   detected.ProxyPort,
			ProjectType: string(detected.Type),
			TypeLabel:   detected.TypeLabel,
			IsWeb:       detected.IsWeb,
		})
	}

	a.sites = scannedSites

	// 1. Write updated Caddyfile
	_ = a.ReloadCaddy()

	// 2. Update system hosts file asynchronously in background (prevents ipconfig /flushdns UI freezes)
	webDomains := make([]string, 0)
	for _, s := range a.sites {
		if s.IsWeb && s.Domain != "" {
			webDomains = append(webDomains, s.Domain)
		}
	}
	go func(domains []string) {
		_ = hosts.UpdateSystemHosts(domains)
	}(webDomains)

	return a.sites
}

// ReloadCaddy writes the updated Caddyfile configuration and reloads Caddy if running
func (a *App) ReloadCaddy() error {
	caddyDir := filepath.Join("data", "caddy")
	_ = os.MkdirAll(caddyDir, 0755)
	caddyfilePath := filepath.Join(caddyDir, "Caddyfile")
	caddyContent := caddy.GenerateCaddyfile(a.sites, 2019)
	if err := os.WriteFile(caddyfilePath, []byte(caddyContent), 0644); err != nil {
		return err
	}

	// Also write to data/Caddyfile for standard root access
	_ = os.WriteFile(filepath.Join("data", "Caddyfile"), []byte(caddyContent), 0644)

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
func (a *App) SyncHosts() (string, error) {
	webDomains := make([]string, 0)
	for _, s := range a.sites {
		if s.IsWeb && s.Domain != "" {
			webDomains = append(webDomains, s.Domain)
		}
	}
	if err := hosts.UpdateSystemHosts(webDomains); err != nil {
		return "", err
	}
	if len(webDomains) == 0 {
		return "تم تنظيف ملف hosts بنجاح (لا توجد مواقع نشطة حالياً).", nil
	}

	return fmt.Sprintf("تمت مزامنة %d نطاق بنجاح مع ملف hosts المحلي وتحديث كاش DNS.", len(webDomains)), nil
}

