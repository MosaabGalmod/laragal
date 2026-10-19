package caddy

import (
	"fmt"
	"strings"
)

type SiteDef struct {
	Domain   string `json:"domain"`
	RootPath string `json:"root_path"`
	PHPPort  int    `json:"php_port"`
}

func GenerateCaddyfile(sites []SiteDef, adminPort int) string {
	var sb strings.Builder

	// Global configuration block
	sb.WriteString("{\n")
	sb.WriteString(fmt.Sprintf("    admin localhost:%d\n", adminPort))
	sb.WriteString("    local_certs\n")
	sb.WriteString("    auto_https disable_redirects\n")
	sb.WriteString("}\n\n")

	// Render each configured site
	for _, site := range sites {
		// Normalize path to forward slashes for Caddyfile
		normalizedRoot := strings.ReplaceAll(site.RootPath, "\\", "/")

		sb.WriteString(fmt.Sprintf("%s {\n", site.Domain))
		sb.WriteString(fmt.Sprintf("    root * \"%s\"\n", normalizedRoot))
		sb.WriteString("    encode gzip zstd\n")
		sb.WriteString(fmt.Sprintf("    php_fastcgi 127.0.0.1:%d\n", site.PHPPort))
		sb.WriteString("    file_server\n")
		sb.WriteString("    tls internal\n")
		sb.WriteString("}\n\n")
	}

	return sb.String()
}
