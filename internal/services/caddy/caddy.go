package caddy

import (
	"fmt"
	"strings"
)

type SiteDef struct {
	Name        string `json:"name"`
	Domain      string `json:"domain"`
	Path        string `json:"path"`
	RootPath    string `json:"root_path"`
	PHPPort     int    `json:"php_port"`
	PHPVersion  string `json:"php_version"`
	ProxyPort   int    `json:"proxy_port,omitempty"`
	ProjectType string `json:"project_type"`
	TypeLabel   string `json:"type_label"`
	IsWeb       bool   `json:"is_web"`
}

func GenerateCaddyfile(sites []SiteDef, adminPort int) string {
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

	return sb.String()
}
