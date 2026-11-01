package caddy_test

import (
	"strings"
	"testing"

	"github.com/laragal/laragal/internal/services/caddy"
)

func TestGenerateCaddyfile(t *testing.T) {
	sites := []caddy.SiteDef{
		{
			Domain:   "laravel.test",
			RootPath: "D:/projects/laravel/public",
			PHPPort:  9083,
			IsWeb:    true,
		},
		{
			Domain:   "wordpress.test",
			RootPath: "D:/projects/wordpress",
			PHPPort:  9082,
			IsWeb:    true,
		},
		{
			Domain:   "static.test",
			RootPath: "D:/projects/static",
			PHPPort:  0,
			IsWeb:    true,
		},
		{
			Domain:      "jarasdesk.test",
			Path:        "D:/projects/jarasdesk",
			RootPath:    "D:/projects/jarasdesk/web",
			PHPPort:     0,
			ProxyPort:   8080,
			ProjectType: "go",
			IsWeb:       true,
		},
		{
			Domain:      "rustapp.test",
			Path:        "D:/projects/rustapp",
			RootPath:    "D:/projects/rustapp/static",
			PHPPort:     0,
			ProxyPort:   8080,
			ProjectType: "rust",
			TypeLabel:   "Rust / ويب",
			IsWeb:       true,
		},
		{
			Domain:      "dotnetapp.test",
			Path:        "D:/projects/dotnetapp",
			RootPath:    "D:/projects/dotnetapp/wwwroot",
			PHPPort:     0,
			ProxyPort:   5000,
			ProjectType: "dotnet",
			TypeLabel:   ".NET / C#",
			IsWeb:       true,
		},
		{
			Domain:      "fastapi.test",
			Path:        "D:/projects/fastapi",
			RootPath:    "D:/projects/fastapi/static",
			PHPPort:     0,
			ProxyPort:   8000,
			ProjectType: "python",
			TypeLabel:   "Python / Web",
			IsWeb:       true,
		},
		{
			Domain:      "flutterapp.test",
			Path:        "D:/projects/flutterapp",
			RootPath:    "D:/projects/flutterapp/build/web",
			PHPPort:     0,
			ProjectType: "flutter",
			TypeLabel:   "Flutter Web",
			IsWeb:       true,
		},
		{
			Domain:   "nonweb.test",
			RootPath: "D:/projects/nonweb",
			PHPPort:  0,
			IsWeb:    false,
		},
	}

	caddyfile := caddy.GenerateCaddyfile(sites, 2019)

	if !strings.Contains(caddyfile, "laravel.test {") {
		t.Errorf("missing laravel.test block")
	}
	if !strings.Contains(caddyfile, "php_fastcgi 127.0.0.1:9083") {
		t.Errorf("missing PHP 9083 proxy")
	}
	if !strings.Contains(caddyfile, "static.test {") {
		t.Errorf("missing static.test block")
	}
	if !strings.Contains(caddyfile, "reverse_proxy 127.0.0.1:8080") {
		t.Errorf("missing Go reverse proxy directive")
	}
	if !strings.Contains(caddyfile, "dotnetapp.test {") || !strings.Contains(caddyfile, "reverse_proxy 127.0.0.1:5000") {
		t.Errorf("missing .NET reverse proxy on 5000")
	}
	if !strings.Contains(caddyfile, "fastapi.test {") || !strings.Contains(caddyfile, "reverse_proxy 127.0.0.1:8000") {
		t.Errorf("missing Python reverse proxy on 8000")
	}
	if !strings.Contains(caddyfile, "flutterapp.test {") {
		t.Errorf("missing flutterapp.test block")
	}
	if strings.Contains(caddyfile, "nonweb.test") {
		t.Errorf("non-web site should not be in Caddyfile")
	}
	if !strings.Contains(caddyfile, "tls internal") {
		t.Errorf("missing tls internal directive")
	}
	if !strings.Contains(caddyfile, "admin localhost:2019") {
		t.Errorf("missing admin block")
	}
	if !strings.Contains(caddyfile, `header Content-Type "text/html; charset=utf-8"`) {
		t.Errorf("missing Content-Type header in handle_errors")
	}
}

func TestGenerateCaddyfileWithAdminer(t *testing.T) {
	sites := []caddy.SiteDef{
		{
			Domain:   "mysite.test",
			RootPath: "D:/projects/mysite",
			PHPPort:  9084,
			IsWeb:    true,
		},
	}

	adminerCfg := caddy.AdminerConfig{
		Enabled:  true,
		Domain:   "adminer.test",
		RootPath: "D:/projects/laragal/data/adminer",
		PHPPort:  9084,
		Port:     3307,
	}

	caddyfile := caddy.GenerateCaddyfile(sites, 2019, adminerCfg)

	if !strings.Contains(caddyfile, "adminer.test, http://localhost:3307 {") {
		t.Errorf("missing adminer block with dual host and port")
	}
	if !strings.Contains(caddyfile, `root * "D:/projects/laragal/data/adminer"`) {
		t.Errorf("missing adminer root path")
	}
	if !strings.Contains(caddyfile, "php_fastcgi 127.0.0.1:9084") {
		t.Errorf("missing php_fastcgi proxy in adminer block")
	}
}
func TestGenerateCaddyfileWithPhpMyAdmin(t *testing.T) {
	sites := []caddy.SiteDef{
		{
			Domain:   "mysite.test",
			RootPath: "D:/projects/mysite",
			PHPPort:  9084,
			IsWeb:    true,
		},
	}

	adminerCfg := caddy.AdminerConfig{
		Enabled:  true,
		Domain:   "adminer.test",
		RootPath: "D:/projects/laragal/data/adminer",
		PHPPort:  9084,
		Port:     3307,
	}

	pmaCfg := caddy.PhpMyAdminConfig{
		Enabled:  true,
		Domain:   "phpmyadmin.test",
		RootPath: "D:/projects/laragal/data/phpmyadmin",
		PHPPort:  9084,
		Port:     3308,
	}

	caddyfile := caddy.GenerateCaddyfileWithTools(sites, 2019, adminerCfg, pmaCfg)

	if !strings.Contains(caddyfile, "adminer.test, http://localhost:3307 {") {
		t.Errorf("missing adminer block")
	}
	if !strings.Contains(caddyfile, "phpmyadmin.test, http://localhost:3308 {") {
		t.Errorf("missing phpmyadmin block with dual host and port")
	}
	if !strings.Contains(caddyfile, `root * "D:/projects/laragal/data/phpmyadmin"`) {
		t.Errorf("missing phpmyadmin root path")
	}
}

func TestCaddyfileSitesLoopbackBinding(t *testing.T) {
	sites := []caddy.SiteDef{
		{
			Domain:   "laravel.test",
			RootPath: "D:/projects/laravel/public",
			PHPPort:  9083,
			IsWeb:    true,
		},
		{
			Domain:      "nodeapp.test",
			RootPath:    "D:/projects/nodeapp",
			ProjectType: "node",
			ProxyPort:   3000,
			IsWeb:       true,
		},
		{
			Domain:   "static.test",
			RootPath: "D:/projects/static",
			IsWeb:    true,
		},
	}
	adminerCfg := caddy.AdminerConfig{
		Enabled:  true,
		Domain:   "adminer.test",
		RootPath: "D:/projects/laragal/data/adminer",
		PHPPort:  9084,
		Port:     3307,
	}
	pmaCfg := caddy.PhpMyAdminConfig{
		Enabled:  true,
		Domain:   "phpmyadmin.test",
		RootPath: "D:/projects/laragal/data/phpmyadmin",
		PHPPort:  9084,
		Port:     3308,
	}

	caddyfile := caddy.GenerateCaddyfileWithTools(sites, 2019, adminerCfg, pmaCfg)

	lines := strings.Split(caddyfile, "\n")
	siteBlockCount := 0
	bindCount := 0

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Top-level site block opener: non-indented, ends with '{', and is not the global '{' block
		if strings.HasSuffix(trimmed, "{") && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
			if trimmed == "{" {
				// Global options block
				continue
			}
			siteBlockCount++
			if i+1 >= len(lines) || strings.TrimSpace(lines[i+1]) != "bind 127.0.0.1 [::1]" {
				t.Errorf("site block %q at line %d does not have 'bind 127.0.0.1 [::1]' as first directive; next line: %q", line, i+1, lines[i+1])
			}
		}
		if trimmed == "bind 127.0.0.1 [::1]" {
			bindCount++
		}
	}

	expectedSites := len(sites) + 2 // 3 user sites + adminer + phpmyadmin
	if siteBlockCount != expectedSites {
		t.Fatalf("expected %d site blocks, found %d", expectedSites, siteBlockCount)
	}
	if bindCount != siteBlockCount {
		t.Errorf("bind directive count (%d) does not match site block count (%d)", bindCount, siteBlockCount)
	}
}

func TestCaddyfile502BilingualAndEscaping(t *testing.T) {
	sites := []caddy.SiteDef{
		{
			Domain:      "crafted.test",
			Path:        "D:/projects/crafted",
			RootPath:    "D:/projects/crafted",
			ProxyPort:   4567,
			ProjectType: "custom",
			TypeLabel:   `Crafted "Evil" <script>alert(1)</script> \test`,
			IsWeb:       true,
		},
	}

	caddyfile := caddy.GenerateCaddyfile(sites, 2019, caddy.AdminerConfig{})

	// 1. Both languages must be present with proper dir and lang attributes
	hasArBlock := strings.Contains(caddyfile, `dir="rtl" lang="ar"`) || strings.Contains(caddyfile, `dir='rtl' lang='ar'`) || strings.Contains(caddyfile, `dir=\"rtl\" lang=\"ar\"`)
	if !hasArBlock {
		t.Errorf("missing Arabic block with dir=rtl lang=ar")
	}
	hasEnBlock := strings.Contains(caddyfile, `dir="ltr" lang="en"`) || strings.Contains(caddyfile, `dir='ltr' lang='en'`) || strings.Contains(caddyfile, `dir=\"ltr\" lang=\"en\"`)
	if !hasEnBlock {
		t.Errorf("missing English block with dir=ltr lang=en")
	}

	// 2. Dark mode support via @media in a <style> tag
	hasStyle := strings.Contains(caddyfile, `<style>`) || strings.Contains(caddyfile, `<style `)
	hasDarkMode := strings.Contains(caddyfile, `@media (prefers-color-scheme: dark)`) || strings.Contains(caddyfile, `@media(prefers-color-scheme:dark)`)
	if !hasStyle || !hasDarkMode {
		t.Errorf("missing dark mode @media in <style> tag")
	}

	// 3. XSS/Escape assertions: raw <script> or unescaped quotes must not appear
	if strings.Contains(caddyfile, `<script>`) {
		t.Errorf("raw <script> tag found in Caddyfile, expected HTML escaping")
	}
	if !strings.Contains(caddyfile, `&lt;script&gt;`) {
		t.Errorf("expected HTML escaped &lt;script&gt;, not found")
	}

	// 4. Verify respond directive structure
	lines := strings.Split(caddyfile, "\n")
	foundRespond := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "respond ") {
			foundRespond = true
			if !strings.HasSuffix(trimmed, " 502") {
				t.Errorf("respond line does not end with 502 status code: %s", trimmed)
			}
			inner := strings.TrimPrefix(trimmed, "respond \"")
			if !strings.HasSuffix(inner, "\" 502") {
				t.Errorf("respond directive not properly double-quoted: %s", trimmed)
			}
			body := strings.TrimSuffix(inner, "\" 502")
			unescapedQuote := false
			for i := 0; i < len(body); i++ {
				if body[i] == '"' && (i == 0 || body[i-1] != '\\') {
					unescapedQuote = true
					break
				}
			}
			if unescapedQuote {
				t.Errorf("body contains unescaped quote that breaks out of Caddyfile string: %s", body)
			}
		}
	}
	if !foundRespond {
		t.Errorf("no respond directive found in generated Caddyfile")
	}
}

