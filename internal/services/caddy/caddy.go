package caddy

import (
	"fmt"
	"strings"
)

type SiteDef struct {
	Name             string `json:"name"`
	Domain           string `json:"domain"`
	Path             string `json:"path"`
	RootPath         string `json:"root_path"`
	PHPPort          int    `json:"php_port"`
	PHPVersion       string `json:"php_version"`
	ProxyPort        int    `json:"proxy_port,omitempty"`
	ProjectType      string `json:"project_type"`
	TypeLabel        string `json:"type_label"`
	IsWeb            bool   `json:"is_web"`
	DBType           string `json:"db_type,omitempty"`
	DBName           string `json:"db_name,omitempty"`
	DBPath           string `json:"db_path,omitempty"`
	DBStatus         string `json:"db_status,omitempty"`
	DBTablesCount    int    `json:"db_tables_count,omitempty"`
	DBNeedsMigration bool   `json:"db_needs_migration,omitempty"`
}

type PhpMyAdminConfig struct {
	Enabled  bool   `json:"enabled"`
	Domain   string `json:"domain"`
	RootPath string `json:"root_path"`
	PHPPort  int    `json:"php_port"`
	Port     int    `json:"port"`
}

type AdminerConfig struct {
	Enabled  bool   `json:"enabled"`
	Domain   string `json:"domain"`
	RootPath string `json:"root_path"`
	PHPPort  int    `json:"php_port"`
	Port     int    `json:"port"`
}

func GenerateCaddyfile(sites []SiteDef, adminPort int, adminerCfg ...AdminerConfig) string {
	var aCfg AdminerConfig
	if len(adminerCfg) > 0 {
		aCfg = adminerCfg[0]
	}
	return GenerateCaddyfileWithTools(sites, adminPort, aCfg, PhpMyAdminConfig{})
}

func GenerateCaddyfileWithTools(sites []SiteDef, adminPort int, adminerCfg AdminerConfig, pmaCfg PhpMyAdminConfig) string {
	var sb strings.Builder

	// Global configuration block
	sb.WriteString("{\n")
	sb.WriteString(fmt.Sprintf("    admin localhost:%d\n", adminPort))
	sb.WriteString("    local_certs\n")
	sb.WriteString("    skip_install_trust\n")
	sb.WriteString("    auto_https disable_redirects\n")
	sb.WriteString("}\n\n")

	// Render each configured site
	for _, site := range sites {
		if !site.IsWeb || strings.TrimSpace(site.Domain) == "" {
			continue
		}

		// Normalize path to forward slashes for Caddyfile
		normalizedRoot := strings.ReplaceAll(site.RootPath, "\\", "/")

		sb.WriteString(fmt.Sprintf("%s {\n", site.Domain))
		sb.WriteString(fmt.Sprintf("    root * \"%s\"\n", normalizedRoot))
		sb.WriteString("    encode gzip zstd\n")
		if site.PHPPort > 0 {
			sb.WriteString(fmt.Sprintf("    php_fastcgi 127.0.0.1:%d {\n", site.PHPPort))
			sb.WriteString("        capture_stderr\n")
			sb.WriteString("    }\n")
			sb.WriteString("    file_server\n")
		} else if site.ProjectType == "go" || site.ProjectType == "rust" || site.ProjectType == "dotnet" || site.ProjectType == "python" || site.ProjectType == "ruby" || site.ProjectType == "java" || site.ProjectType == "node" || site.ProjectType == "proxy" || site.ProxyPort > 0 {
			proxyPort := site.ProxyPort
			if proxyPort <= 0 {
				if site.ProjectType == "node" {
					proxyPort = 3000
				} else if site.ProjectType == "dotnet" {
					proxyPort = 5000
				} else if site.ProjectType == "python" {
					proxyPort = 8000
				} else if site.ProjectType == "ruby" {
					proxyPort = 3000
				} else {
					proxyPort = 8080
				}
			}
			// Handle static files directly through Caddy
			sb.WriteString("    handle /static/* {\n")
			sb.WriteString(fmt.Sprintf("        root * \"%s\"\n", normalizedRoot))
			sb.WriteString("        file_server\n")
			sb.WriteString("    }\n")
			sb.WriteString("    handle /wwwroot/* {\n")
			sb.WriteString(fmt.Sprintf("        root * \"%s\"\n", normalizedRoot))
			sb.WriteString("        file_server\n")
			sb.WriteString("    }\n")
			sb.WriteString("    handle /assets/* {\n")
			sb.WriteString(fmt.Sprintf("        root * \"%s\"\n", normalizedRoot))
			sb.WriteString("        file_server\n")
			sb.WriteString("    }\n")
			// Reverse proxy all other requests to the backend
			sb.WriteString("    handle {\n")
			sb.WriteString(fmt.Sprintf("        reverse_proxy 127.0.0.1:%d {\n", proxyPort))
			sb.WriteString("            header_up Host {host}\n")
			sb.WriteString("            header_up X-Real-IP {remote_host}\n")
			sb.WriteString("            header_up X-Forwarded-For {remote_host}\n")
			sb.WriteString("            header_up X-Forwarded-Proto {scheme}\n")
			sb.WriteString("        }\n")
			sb.WriteString("    }\n")
			typeLabel := site.TypeLabel
			if typeLabel == "" {
				typeLabel = "الخادم"
			}
			sb.WriteString("    handle_errors {\n")
			sb.WriteString("        header Content-Type \"text/html; charset=utf-8\"\n")
			sb.WriteString(fmt.Sprintf("        respond \"<div dir='rtl' style='font-family:system-ui,sans-serif;max-width:540px;margin:80px auto;padding:32px;background:#f8fafc;border:1px solid #e2e8f0;border-radius:16px;text-align:center;box-shadow:0 4px 12px rgba(0,0,0,0.05);'><h2 style='color:#0f172a;margin-bottom:12px;'>LaraGal — خادم التطبيق غير متصل</h2><p style='color:#475569;line-height:1.6;font-size:14px;'>تطبيق %s غير مشغل حالياً على المنفذ <b>%d</b>.<br>يمكنك تشغيل التطبيق بنقرة واحدة من لوحة تحكم <b>LaraGal</b> أو تشغيل أداة التطوير الخاصة به.</p></div>\" 502\n", typeLabel, proxyPort))
			sb.WriteString("    }\n")
		} else {
			sb.WriteString("    try_files {path} {path}/ /index.html /index.htm\n")
			sb.WriteString("    file_server browse\n")
		}
		sb.WriteString("    tls internal\n")
		sb.WriteString("}\n\n")
	}

	// Render Adminer database manager virtual host if enabled
	if adminerCfg.Enabled {
		cfg := adminerCfg
		normalizedRoot := strings.ReplaceAll(cfg.RootPath, "\\", "/")
		adminerDomain := cfg.Domain
		if adminerDomain == "" {
			adminerDomain = "adminer.test"
		}
		adminerPort := cfg.Port
		if adminerPort <= 0 {
			adminerPort = 3307
		}

		sb.WriteString(fmt.Sprintf("%s, http://localhost:%d {\n", adminerDomain, adminerPort))
		sb.WriteString(fmt.Sprintf("    root * \"%s\"\n", normalizedRoot))
		sb.WriteString("    encode gzip zstd\n")
		if cfg.PHPPort > 0 {
			sb.WriteString(fmt.Sprintf("    php_fastcgi 127.0.0.1:%d {\n", cfg.PHPPort))
			sb.WriteString("        capture_stderr\n")
			sb.WriteString("    }\n")
		}
		sb.WriteString("    file_server\n")
		sb.WriteString("    tls internal\n")
		sb.WriteString("}\n\n")
	}

		// Render phpMyAdmin database manager virtual host if enabled
	if pmaCfg.Enabled {
		normalizedRoot := strings.ReplaceAll(pmaCfg.RootPath, "\\", "/")
		pmaDomain := pmaCfg.Domain
		if pmaDomain == "" {
			pmaDomain = "phpmyadmin.test"
		}
		pmaPort := pmaCfg.Port
		if pmaPort <= 0 {
			pmaPort = 3308
		}

		sb.WriteString(fmt.Sprintf("%s, http://localhost:%d {\n", pmaDomain, pmaPort))
		sb.WriteString(fmt.Sprintf("    root * \"%s\"\n", normalizedRoot))
		sb.WriteString("    encode gzip zstd\n")
		if pmaCfg.PHPPort > 0 {
			sb.WriteString(fmt.Sprintf("    php_fastcgi 127.0.0.1:%d {\n", pmaCfg.PHPPort))
			sb.WriteString("        capture_stderr\n")
			sb.WriteString("    }\n")
		}
		sb.WriteString("    file_server\n")
		sb.WriteString("    tls internal\n")
		sb.WriteString("}\n\n")
	}

	return sb.String()
}
