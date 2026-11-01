package config_test

import (
	"path/filepath"
	"testing"

	"github.com/laragal/laragal/internal/config"
)

func TestDefaultConfigAndSave(t *testing.T) {
	tempDir := t.TempDir()
	cfgPath := filepath.Join(tempDir, "laragal.json")

	cfg := config.NewDefaultConfig()
	if cfg.DomainSuffix != ".test" {
		t.Fatalf("expected .test, got %s", cfg.DomainSuffix)
	}
	if !cfg.StartOnBoot {
		t.Fatalf("expected StartOnBoot to default to true")
	}
	if len(cfg.HiddenRuntimes) != 0 {
		t.Fatalf("expected HiddenRuntimes to default to empty slice, got %v", cfg.HiddenRuntimes)
	}

	cfg.StartOnBoot = false
	cfg.HiddenRuntimes = []string{"php-8.2", "mariadb_series"}
	err := cfg.Save(cfgPath)
	if err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	loaded, err := config.LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if loaded.DomainSuffix != ".test" {
		t.Fatalf("expected .test, got %s", loaded.DomainSuffix)
	}
	if loaded.StartOnBoot {
		t.Fatalf("expected StartOnBoot to be false after reload")
	}
	if len(loaded.HiddenRuntimes) != 2 || loaded.HiddenRuntimes[0] != "php-8.2" || loaded.HiddenRuntimes[1] != "mariadb_series" {
		t.Fatalf("expected HiddenRuntimes to persist, got %v", loaded.HiddenRuntimes)
	}
}
