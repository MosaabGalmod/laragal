package mysql

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/laragal/laragal/internal/projects"
)

// SyncProjectDatabaseEnv creates or updates the project's .env file to match the target database
func SyncProjectDatabaseEnv(projectPath string, targetDB string) error {
	cleanPath := filepath.Clean(projectPath)
	envPath := filepath.Join(cleanPath, ".env")
	examplePath := filepath.Join(cleanPath, ".env.example")

	var content string

	if data, err := os.ReadFile(envPath); err == nil {
		content = string(data)
	} else if data, err := os.ReadFile(examplePath); err == nil {
		content = string(data)
	} else {
		content = defaultEnvTemplate
	}

	normDB := NormalizeTargetDB(targetDB)
	cfg := projects.EnvConfig{
		DBName:   normDB,
		DBUser:   "root",
		DBPass:   "",
		MailHost: "127.0.0.1",
		MailPort: "1025",
	}

	updated := projects.InjectEnvVariables(content, cfg)
	return os.WriteFile(envPath, []byte(updated), 0644)
}

const defaultEnvTemplate = `APP_NAME=Laravel
APP_ENV=local
APP_KEY=
APP_DEBUG=true
APP_URL=http://localhost

DB_CONNECTION=mysql
DB_HOST=127.0.0.1
DB_PORT=3306
DB_DATABASE=app_db
DB_USERNAME=root
DB_PASSWORD=

MAIL_MAILER=smtp
MAIL_HOST=127.0.0.1
MAIL_PORT=1025
`

// ImportDiscoveredDatabase coordinates importing schema or legacy data folder and syncing .env
func ImportDiscoveredDatabase(projectPath string, item DiscoveredDatabase) error {
	normDB := NormalizeTargetDB(item.TargetDB)
	if normDB == "" {
		normDB = NormalizeTargetDB(filepath.Base(projectPath))
	}

	// 1. Sync project .env file
	if err := SyncProjectDatabaseEnv(projectPath, normDB); err != nil {
		return fmt.Errorf("failed to sync project .env: %w", err)
	}

	// 2. If it's a legacy server directory, copy database directory safely
	if item.Type == "legacy_server_dir" && item.SourcePath != "" {
		targetDataDir := filepath.Join("data", "mysql", normDB)
		if err := copyDirectorySafely(item.SourcePath, targetDataDir); err != nil {
			return fmt.Errorf("failed to copy legacy database files: %w", err)
		}
	}

	return nil
}

func copyDirectorySafely(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !srcInfo.IsDir() {
		return fmt.Errorf("source is not a directory: %s", src)
	}

	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		s := filepath.Join(src, entry.Name())
		d := filepath.Join(dst, entry.Name())
		if entry.IsDir() {
			continue // MySQL database directories are flat table lists (.ibd, .frm, .opt)
		}
		data, err := os.ReadFile(s)
		if err != nil {
			return err
		}
		if err := os.WriteFile(d, data, 0644); err != nil {
			return err
		}
	}
	return nil
}
