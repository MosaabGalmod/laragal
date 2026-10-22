package php_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/laragal/laragal/internal/services/php"
)

func TestEnsurePHPIni_FromDevelopmentTemplate(t *testing.T) {
	tempDir := t.TempDir()

	devContent := `; Sample php.ini-development
memory_limit = 128M
;extension_dir = "ext"
;extension=curl
;extension=pdo_mysql
;extension=mbstring
`
	devFile := filepath.Join(tempDir, "php.ini-development")
	if err := os.WriteFile(devFile, []byte(devContent), 0644); err != nil {
		t.Fatal(err)
	}

	if err := php.EnsurePHPIni(tempDir); err != nil {
		t.Fatalf("EnsurePHPIni failed: %v", err)
	}

	iniPath := filepath.Join(tempDir, "php.ini")
	data, err := os.ReadFile(iniPath)
	if err != nil {
		t.Fatalf("failed to read created php.ini: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "extension_dir = \"ext\"") {
		t.Errorf("expected extension_dir to be enabled")
	}
	if !strings.Contains(content, "extension=curl") {
		t.Errorf("expected extension=curl to be enabled")
	}
	if !strings.Contains(content, "extension=pdo_mysql") {
		t.Errorf("expected extension=pdo_mysql to be enabled")
	}
	if !strings.Contains(content, "memory_limit = 512M") {
		t.Errorf("expected memory_limit to be increased to 512M")
	}
}

func TestEnsurePHPIni_FallbackWhenNoTemplate(t *testing.T) {
	tempDir := t.TempDir()

	if err := php.EnsurePHPIni(tempDir); err != nil {
		t.Fatalf("EnsurePHPIni failed on empty dir: %v", err)
	}

	iniPath := filepath.Join(tempDir, "php.ini")
	data, err := os.ReadFile(iniPath)
	if err != nil {
		t.Fatalf("failed to read created fallback php.ini: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "extension=pdo_mysql") {
		t.Errorf("expected extension=pdo_mysql in fallback php.ini")
	}
}
