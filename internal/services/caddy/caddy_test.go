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
	if strings.Contains(caddyfile, "nonweb.test") {
		t.Errorf("non-web site should not be in Caddyfile")
	}
	if !strings.Contains(caddyfile, "tls internal") {
		t.Errorf("missing tls internal directive")
	}
	if !strings.Contains(caddyfile, "admin localhost:2019") {
		t.Errorf("missing admin block")
	}
}
