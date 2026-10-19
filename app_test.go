package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/laragal/laragal/internal/config"
)

func TestProjectsDirectoryLogic(t *testing.T) {
	tempDir := t.TempDir()
	cfgPath := filepath.Join(tempDir, "laragal.json")

	cfg := config.NewDefaultConfig()
	cfg.ProjectsPath = "C:\\laragal\\sites"
	if err := cfg.Save(cfgPath); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	app := &App{
		cfg:     cfg,
		cfgPath: cfgPath,
	}

	// Verify initial get
	if got := app.GetProjectsDirectory(); got != "C:\\laragal\\sites" {
		t.Errorf("expected C:\\laragal\\sites, got %s", got)
	}

	// Update directory
	newDir := filepath.Join(tempDir, "my-custom-projects")
	_ = os.MkdirAll(newDir, 0755)
	app.cfg.ProjectsPath = newDir
	if err := app.cfg.Save(app.cfgPath); err != nil {
		t.Fatalf("failed to save updated config: %v", err)
	}

	// Reload config from disk
	loadedCfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("failed to reload config: %v", err)
	}

	if loadedCfg.ProjectsPath != newDir {
		t.Errorf("expected loaded path %s, got %s", newDir, loadedCfg.ProjectsPath)
	}
}
