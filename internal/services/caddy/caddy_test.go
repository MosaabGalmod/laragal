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
