package mysql_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/laragal/laragal/internal/services/mysql"
)

func TestIsDataDirInitialized(t *testing.T) {
	tempDir := t.TempDir()

	if mysql.IsDataDirInitialized(tempDir) {
		t.Errorf("expected empty dir to not be initialized")
	}

	// Create marker file
	if err := os.WriteFile(filepath.Join(tempDir, "ibdata1"), []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}

	if !mysql.IsDataDirInitialized(tempDir) {
		t.Errorf("expected dir with ibdata1 to be recognized as initialized")
	}
}

func TestFindInstallerCandidate(t *testing.T) {
	tempDir := t.TempDir()

	if found := mysql.FindInstallerCandidate(tempDir); found != "" {
		t.Errorf("expected empty string when no installer exists, got: %s", found)
	}

	mariadbInstall := filepath.Join(tempDir, "mariadb-install-db.exe")
	if err := os.WriteFile(mariadbInstall, []byte("fake"), 0755); err != nil {
		t.Fatal(err)
	}

	if found := mysql.FindInstallerCandidate(tempDir); found != mariadbInstall {
		t.Errorf("expected %s, got: %s", mariadbInstall, found)
	}
}

func TestEnsureDataDir_WhenAlreadyInitialized(t *testing.T) {
	tempDir := t.TempDir()
	dataDir := filepath.Join(tempDir, "data")
	_ = os.MkdirAll(dataDir, 0755)
	_ = os.WriteFile(filepath.Join(dataDir, "ibdata1"), []byte("data"), 0644)

	// Since it's already initialized, EnsureDataDir should return nil without running mysqld
	err := mysql.EnsureDataDir("nonexistent-mysqld.exe", dataDir)
	if err != nil {
		t.Errorf("expected nil when data directory already initialized, got: %v", err)
	}
}
