package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/laragal/laragal/internal/certs"
	"github.com/laragal/laragal/internal/config"
	"github.com/laragal/laragal/internal/downloader"
	"github.com/laragal/laragal/internal/hosts"
	"github.com/laragal/laragal/internal/privilege"
	"github.com/laragal/laragal/internal/process"
	"github.com/laragal/laragal/internal/projects"
	"github.com/laragal/laragal/internal/services/caddy"
	"github.com/laragal/laragal/internal/services/php"
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

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) shutdown(ctx context.Context) {
	_ = a.supervisor.StopAll()
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
		Domain:   domain,
		RootPath: rootPath,
		PHPPort:  phpPort,
	}

	a.sites = append(a.sites, newSite)

	domains := make([]string, 0, len(a.sites))
	for _, s := range a.sites {
		domains = append(domains, s.Domain)
	}
	_ = hosts.UpdateSystemHosts(domains)

	return nil
}

func (a *App) InstallSSLTrust() (string, error) {
	rootCertPath := filepath.Join("data", "caddy", "pki", "authorities", "local", "root.crt")
	if _, err := os.Stat(rootCertPath); os.IsNotExist(err) {
		return "لم يتم العثور على شهادة الجذر لـ Caddy بعد. يرجى تشغيل خادم Caddy أولاً.", nil
	}

	fingerprint, err := certs.ReadCertFingerprint(rootCertPath)
	if err != nil {
		return "", err
	}

	trusted, _ := certs.IsCertTrustedInWindowsStore(fingerprint)
	if trusted {
		return "شهادة الجذر المحلية موثوقة بالفعل في نظام ويندوز.", nil
	}

	cmdArgs := certs.BuildTrustCertCommand(rootCertPath)
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		// Attempt elevated execution via PowerShell Start-Process -Verb RunAs
		elevatedCmd := privilege.BuildElevatedCertCommand(rootCertPath)
		elCmd := exec.Command(elevatedCmd[0], elevatedCmd[1:]...)
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
	resp, err := exec.Command("powershell", "-NoProfile", "-Command", fmt.Sprintf("Invoke-WebRequest -Uri '%s' -OutFile '%s'", url, zipPath)).CombinedOutput()
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
	return exec.Command("explorer", filepath.FromSlash(folderPath)).Start()
}

func (a *App) OpenBrowser(targetURL string) error {
	return exec.Command("rundll32", "url.dll,FileProtocolHandler", targetURL).Start()
}

func (a *App) OpenVSCode(folderPath string) error {
	return exec.Command("cmd", "/c", "code", filepath.FromSlash(folderPath)).Start()
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

// SelectProjectsDirectory displays a native directory chooser dialog and updates configuration
func (a *App) SelectProjectsDirectory() (string, error) {
	if a.ctx == nil {
		return a.cfg.ProjectsPath, nil
	}

	selectedDir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		DefaultDirectory: a.cfg.ProjectsPath,
		Title:            "اختر مجلد المشاريع / Select Projects Directory",
	})
	if err != nil {
		return "", err
	}
	if selectedDir == "" {
		return a.cfg.ProjectsPath, nil
	}

	a.cfg.ProjectsPath = filepath.Clean(selectedDir)
	_ = a.cfg.Save(a.cfgPath)
	return a.cfg.ProjectsPath, nil
}

