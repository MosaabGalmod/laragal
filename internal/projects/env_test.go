package projects_test

import (
	"strings"
	"testing"

	"github.com/laragal/laragal/internal/projects"
)

func TestInjectEnvVariables(t *testing.T) {
	sampleEnv := `APP_NAME=Laravel
APP_URL=http://localhost
DB_CONNECTION=sqlite
DB_HOST=127.0.0.1
DB_DATABASE=laravel
`

	cfg := projects.EnvConfig{
		AppURL:   "https://mysite.test",
		DBName:   "mysite_db",
		DBUser:   "root",
		DBPass:   "",
		MailHost: "127.0.0.1",
		MailPort: "1025",
	}

	result := projects.InjectEnvVariables(sampleEnv, cfg)

	if !strings.Contains(result, "APP_URL=https://mysite.test") {
		t.Errorf("failed to inject APP_URL: %s", result)
	}
	if !strings.Contains(result, "DB_DATABASE=mysite_db") {
		t.Errorf("failed to inject DB_DATABASE: %s", result)
	}
	if !strings.Contains(result, "MAIL_HOST=127.0.0.1") {
		t.Errorf("failed to inject MAIL_HOST: %s", result)
	}
}
