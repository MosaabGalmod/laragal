package projects

import (
	"fmt"
	"strings"
)

type EnvConfig struct {
	AppURL   string `json:"app_url"`
	DBName   string `json:"db_name"`
	DBUser   string `json:"db_user"`
	DBPass   string `json:"db_pass"`
	MailHost string `json:"mail_host"`
	MailPort string `json:"mail_port"`
}

func InjectEnvVariables(content string, cfg EnvConfig) string {
	keysToUpdate := map[string]string{
		"APP_URL":       cfg.AppURL,
		"DB_CONNECTION": "mysql",
		"DB_HOST":       "127.0.0.1",
		"DB_PORT":       "3306",
		"DB_DATABASE":   cfg.DBName,
		"DB_USERNAME":   cfg.DBUser,
		"DB_PASSWORD":   cfg.DBPass,
		"MAIL_MAILER":   "smtp",
		"MAIL_HOST":     cfg.MailHost,
		"MAIL_PORT":     cfg.MailPort,
	}

	lines := strings.Split(content, "\n")
	foundKeys := make(map[string]bool)

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") || !strings.Contains(trimmed, "=") {
			continue
		}
		parts := strings.SplitN(trimmed, "=", 2)
		key := strings.TrimSpace(parts[0])
		if newVal, exists := keysToUpdate[key]; exists {
			lines[i] = fmt.Sprintf("%s=%s", key, newVal)
			foundKeys[key] = true
		}
	}

	// Append any missing keys that weren't originally in the file
	for k, v := range keysToUpdate {
		if !foundKeys[k] {
			lines = append(lines, fmt.Sprintf("%s=%s", k, v))
		}
	}

	return strings.Join(lines, "\n")
}
