package phpmyadmin_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/laragal/laragal/internal/services/phpmyadmin"
)

func TestBuildURL(t *testing.T) {
	urlNoDB := phpmyadmin.BuildURL("phpmyadmin.test", 3308, "")
	if !strings.HasPrefix(urlNoDB, "http://localhost:3308/") {
		t.Errorf("expected localhost:3308 baseURL, got: %s", urlNoDB)
	}

	urlWithDB := phpmyadmin.BuildURL("phpmyadmin.test", 3308, "my_database")
	if !strings.Contains(urlWithDB, "db=my_database") {
		t.Errorf("expected db query parameter, got: %s", urlWithDB)
	}
}

func TestEnsureConfig(t *testing.T) {
	tmpDir := t.TempDir()
	err := phpmyadmin.EnsureConfig(tmpDir)
	if err != nil {
		t.Fatalf("unexpected EnsureConfig error: %v", err)
	}

	cfgPath := filepath.Join(tmpDir, "data", "phpmyadmin", "config.inc.php")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("config.inc.php was not created: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "blowfish_secret") || !strings.Contains(content, "auth_type") {
		t.Errorf("config.inc.php missing critical directives: %s", content)
	}
}

func TestIsInstalled(t *testing.T) {
	tmpDir := t.TempDir()
	if phpmyadmin.IsInstalled(tmpDir) {
		t.Errorf("expected false for empty directory")
	}

	pmaDir := filepath.Join(tmpDir, "data", "phpmyadmin")
	_ = os.MkdirAll(pmaDir, 0755)
	_ = os.WriteFile(filepath.Join(pmaDir, "index.php"), []byte("<?php echo 'pma';"), 0644)

	if !phpmyadmin.IsInstalled(tmpDir) {
		t.Errorf("expected true when index.php is present")
	}
}
