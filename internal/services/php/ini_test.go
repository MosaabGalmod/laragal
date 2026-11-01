package php_test

import (
	"fmt"
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
	expectedExt := filepath.ToSlash(filepath.Join(tempDir, "ext"))
	if !strings.Contains(content, fmt.Sprintf("extension_dir = \"%s\"", expectedExt)) {
		t.Errorf("expected extension_dir to be %q, got:\n%s", expectedExt, content)
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

func TestEnsurePHPIni_OPcache_NewIni_PHP84(t *testing.T) {
	tempDir := t.TempDir()
	devContent := `[PHP]
memory_limit = 128M
;zend_extension=opcache
[opcache]
opcache.enable = 1
opcache.memory_consumption = 128
`
	devFile := filepath.Join(tempDir, "php.ini-development")
	if err := os.WriteFile(devFile, []byte(devContent), 0644); err != nil {
		t.Fatal(err)
	}

	if err := php.EnsurePHPIni(tempDir, "8.4"); err != nil {
		t.Fatalf("EnsurePHPIni failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(tempDir, "php.ini"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	// For 8.4, zend_extension=opcache should be active
	if !strings.Contains(content, "zend_extension=opcache") {
		t.Errorf("expected zend_extension=opcache to be active in PHP 8.4, got:\n%s", content)
	}
	if strings.Contains(content, ";zend_extension=opcache") {
		t.Errorf("expected zend_extension=opcache not to be commented out in PHP 8.4")
	}
	if !strings.Contains(content, "[opcache]") || !strings.Contains(content, "opcache.memory_consumption = 128") {
		t.Errorf("expected [opcache] settings to be preserved")
	}
}

func TestEnsurePHPIni_OPcache_NewIni_PHP85(t *testing.T) {
	tempDir := t.TempDir()
	devContent := `[PHP]
memory_limit = 128M
;zend_extension=opcache
[opcache]
opcache.enable = 1
opcache.memory_consumption = 128
`
	devFile := filepath.Join(tempDir, "php.ini-development")
	if err := os.WriteFile(devFile, []byte(devContent), 0644); err != nil {
		t.Fatal(err)
	}

	if err := php.EnsurePHPIni(tempDir, "8.5"); err != nil {
		t.Fatalf("EnsurePHPIni failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(tempDir, "php.ini"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	// For 8.5, zend_extension=opcache must be commented out or skipped
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "zend_extension") && strings.Contains(trimmed, "opcache") {
			t.Errorf("expected no active zend_extension=opcache line in PHP 8.5, found: %s", line)
		}
	}
	if !strings.Contains(content, ";zend_extension=opcache") {
		t.Errorf("expected ;zend_extension=opcache to be commented out in PHP 8.5")
	}
	if !strings.Contains(content, "[opcache]") || !strings.Contains(content, "opcache.memory_consumption = 128") {
		t.Errorf("expected [opcache] settings to still apply in PHP 8.5")
	}
}

func TestEnsurePHPIni_OPcache_UpgradePreservedIni_84to85(t *testing.T) {
	tempDir := t.TempDir()
	existingIni := `[PHP]
memory_limit = 512M
zend_extension=opcache
extension=pdo_mysql

[opcache]
opcache.enable = 1
opcache.memory_consumption = 256
`
	iniFile := filepath.Join(tempDir, "php.ini")
	if err := os.WriteFile(iniFile, []byte(existingIni), 0644); err != nil {
		t.Fatal(err)
	}

	// Upgrading to PHP 8.5 on existing preserved php.ini
	if err := php.EnsurePHPIni(tempDir, "8.5"); err != nil {
		t.Fatalf("EnsurePHPIni failed: %v", err)
	}

	data, err := os.ReadFile(iniFile)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	// Verify zend_extension=opcache is commented out
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "zend_extension") && strings.Contains(trimmed, "opcache") {
			t.Errorf("expected no active zend_extension=opcache line after upgrade to 8.5, found: %s", line)
		}
	}
	if !strings.Contains(content, ";zend_extension=opcache") {
		t.Errorf("expected ;zend_extension=opcache commented out after upgrade to 8.5")
	}
	// Settings in [opcache] must still apply
	if !strings.Contains(content, "[opcache]") || !strings.Contains(content, "opcache.memory_consumption = 256") {
		t.Errorf("expected [opcache] settings to be preserved after upgrade to 8.5")
	}
}

func TestEnsurePHPIni_OPcache_UpgradePreservedIni_83to84(t *testing.T) {
	tempDir := t.TempDir()
	existingIni := `[PHP]
memory_limit = 512M
zend_extension=opcache
extension=pdo_mysql

[opcache]
opcache.enable = 1
opcache.memory_consumption = 128
`
	iniFile := filepath.Join(tempDir, "php.ini")
	if err := os.WriteFile(iniFile, []byte(existingIni), 0644); err != nil {
		t.Fatal(err)
	}

	// Preserving php.ini on PHP 8.4 upgrade
	if err := php.EnsurePHPIni(tempDir, "8.4"); err != nil {
		t.Fatalf("EnsurePHPIni failed: %v", err)
	}

	data, err := os.ReadFile(iniFile)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	// For 8.4, zend_extension=opcache must remain active
	if !strings.Contains(content, "zend_extension=opcache") {
		t.Errorf("expected zend_extension=opcache to be kept active for PHP 8.4")
	}
	if strings.Contains(content, ";zend_extension=opcache") {
		t.Errorf("expected zend_extension=opcache not to be commented out for PHP 8.4")
	}
}

func TestEnsurePHPIni_OPcache_AutoDetectFromDir(t *testing.T) {
	baseDir := t.TempDir()
	php85Dir := filepath.Join(baseDir, "php-8.5")
	if err := os.MkdirAll(php85Dir, 0755); err != nil {
		t.Fatal(err)
	}
	existingIni := `[PHP]
zend_extension=opcache
[opcache]
opcache.enable = 1
`
	if err := os.WriteFile(filepath.Join(php85Dir, "php.ini"), []byte(existingIni), 0644); err != nil {
		t.Fatal(err)
	}

	// No versionHint passed; should auto-detect from dir name "php-8.5"
	if err := php.EnsurePHPIni(php85Dir); err != nil {
		t.Fatalf("EnsurePHPIni failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(php85Dir, "php.ini"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, ";zend_extension=opcache") {
		t.Errorf("expected ;zend_extension=opcache to be commented out via auto-detection")
	}
}


func TestEnsurePHPIni_OPcache_PHP86_NoZendExtension(t *testing.T) {
	tempDir := t.TempDir()
	devContent := `[PHP]
memory_limit = 128M
;zend_extension=opcache
[opcache]
opcache.enable = 1
`
	devFile := filepath.Join(tempDir, "php.ini-development")
	if err := os.WriteFile(devFile, []byte(devContent), 0644); err != nil {
		t.Fatal(err)
	}

	if err := php.EnsurePHPIni(tempDir, "8.6"); err != nil {
		t.Fatalf("EnsurePHPIni failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(tempDir, "php.ini"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	// For 8.6, zend_extension=opcache must NOT be active
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "zend_extension") && strings.Contains(trimmed, "opcache") {
			t.Errorf("expected no active zend_extension=opcache in PHP 8.6, found: %s", line)
		}
	}
	if !strings.Contains(content, ";zend_extension=opcache") {
		t.Errorf("expected ;zend_extension=opcache to be commented out in PHP 8.6")
	}
	if !strings.Contains(content, "[opcache]") {
		t.Errorf("expected [opcache] block to be present in PHP 8.6")
	}
}

func TestEnsurePHPIni_OPcache_PHP90_NoZendExtension(t *testing.T) {
	tempDir := t.TempDir()
	existingIni := `[PHP]
zend_extension=opcache
[opcache]
opcache.enable = 1
`
	if err := os.WriteFile(filepath.Join(tempDir, "php.ini"), []byte(existingIni), 0644); err != nil {
		t.Fatal(err)
	}

	if err := php.EnsurePHPIni(tempDir, "9.0"); err != nil {
		t.Fatalf("EnsurePHPIni failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(tempDir, "php.ini"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "zend_extension") && strings.Contains(trimmed, "opcache") {
			t.Errorf("expected no active zend_extension=opcache in PHP 9.0, found: %s", line)
		}
	}
	if !strings.Contains(content, ";zend_extension=opcache") {
		t.Errorf("expected ;zend_extension=opcache to be commented out in PHP 9.0")
	}
}

func TestEnsurePHPIni_HealsStaleStagingExtensionDirs(t *testing.T) {
	tempDir := t.TempDir()

	staleIni := `[PHP]
extension_dir = "D:/projects/laragal/bin/.staging-php-8.5-12345/ext"
extension_dir = "D:/projects/laragal/bin/.staging-php-8.5-67890/ext"
extension_dir = "D:/projects/laragal/bin/.staging-php-8.5-abcde/ext"
extension=curl
`
	iniPath := filepath.Join(tempDir, "php.ini")
	if err := os.WriteFile(iniPath, []byte(staleIni), 0644); err != nil {
		t.Fatal(err)
	}

	if err := php.EnsurePHPIni(tempDir, "8.5"); err != nil {
		t.Fatalf("EnsurePHPIni failed: %v", err)
	}

	data, err := os.ReadFile(iniPath)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	if strings.Contains(content, ".staging-") {
		t.Errorf("expected all .staging- references to be healed, got:\n%s", content)
	}

	expectedExt := filepath.ToSlash(filepath.Join(tempDir, "ext"))
	expectedLine := fmt.Sprintf("extension_dir = \"%s\"", expectedExt)

	extCount := 0
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "extension_dir =") || strings.HasPrefix(trimmed, ";extension_dir =") {
			extCount++
			if trimmed != expectedLine {
				t.Errorf("expected healed line %q, got %q", expectedLine, trimmed)
			}
		}
	}

	if extCount != 1 {
		t.Errorf("expected exactly 1 extension_dir line after healing, got %d", extCount)
	}
}

func TestEnsurePHPIni_PreservesValidCustomExtensionDir(t *testing.T) {
	tempDir := t.TempDir()

	customExtDir := t.TempDir() // Real existing directory on disk
	customExtSlash := filepath.ToSlash(customExtDir)

	iniWithCustom := fmt.Sprintf(`[PHP]
;extension_dir = "ext"
extension_dir = "%s"
extension=curl
`, customExtSlash)

	iniPath := filepath.Join(tempDir, "php.ini")
	if err := os.WriteFile(iniPath, []byte(iniWithCustom), 0644); err != nil {
		t.Fatal(err)
	}

	if err := php.EnsurePHPIni(tempDir, "8.5"); err != nil {
		t.Fatalf("EnsurePHPIni failed: %v", err)
	}

	data, err := os.ReadFile(iniPath)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	expectedLine := fmt.Sprintf("extension_dir = \"%s\"", customExtSlash)

	extCount := 0
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "extension_dir =") || strings.HasPrefix(trimmed, ";extension_dir =") {
			extCount++
			if trimmed != expectedLine {
				t.Errorf("expected valid custom extension_dir line %q to be preserved, got %q", expectedLine, trimmed)
			}
		}
	}

	if extCount != 1 {
		t.Errorf("expected exactly 1 extension_dir line, got %d", extCount)
	}
}
