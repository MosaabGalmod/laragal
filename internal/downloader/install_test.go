package downloader_test

import (
	"archive/zip"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/laragal/laragal/internal/downloader"
)

func createTestZip(t *testing.T, zipPath string, files map[string]string) {
	t.Helper()
	_ = os.MkdirAll(filepath.Dir(zipPath), 0755)
	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	w := zip.NewWriter(f)
	defer w.Close()

	for name, content := range files {
		fw, err := w.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := fw.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestInstallArtifactZip_LeftoverFilesRemoved(t *testing.T) {
	tempBase := t.TempDir()
	destDir := filepath.Join(tempBase, "bin", "php-8.5")
	if err := os.MkdirAll(filepath.Join(destDir, "ext"), 0755); err != nil {
		t.Fatal(err)
	}

	// Old files from PHP 8.4
	_ = os.WriteFile(filepath.Join(destDir, "php.exe"), []byte("old-php-8.4"), 0755)
	_ = os.WriteFile(filepath.Join(destDir, "php-cgi.exe"), []byte("old-php-cgi-8.4"), 0755)
	_ = os.WriteFile(filepath.Join(destDir, "php8ts.dll"), []byte("old-dll"), 0644)
	_ = os.WriteFile(filepath.Join(destDir, "php8apache2_4.dll"), []byte("old-apache-dll"), 0644)
	_ = os.WriteFile(filepath.Join(destDir, "ext", "php_opcache.dll"), []byte("old-opcache-dll"), 0644)
	_ = os.WriteFile(filepath.Join(destDir, "stale_file.txt"), []byte("stale"), 0644)

	// New PHP 8.5 zip package without opcache dll (opcache is core in 8.5)
	zipPath := filepath.Join(tempBase, "php-8.5.zip")
	createTestZip(t, zipPath, map[string]string{
		"php.exe":             "new-php-8.5",
		"php-cgi.exe":         "new-php-cgi-8.5",
		"ext/php_curl.dll":    "new-curl-dll",
		"php.ini-development": "[PHP]\nmemory_limit = 128M\n",
	})

	err := downloader.InstallArtifactZip("php-8.5", zipPath, destDir)
	if err != nil {
		t.Fatalf("InstallArtifactZip failed: %v", err)
	}

	// Verify new files exist
	newExe, err := os.ReadFile(filepath.Join(destDir, "php.exe"))
	if err != nil || string(newExe) != "new-php-8.5" {
		t.Errorf("expected new php.exe, got: %s (err: %v)", string(newExe), err)
	}
	if _, err := os.Stat(filepath.Join(destDir, "ext", "php_curl.dll")); err != nil {
		t.Errorf("expected ext/php_curl.dll to exist")
	}

	// Verify leftover files from old version are GONE
	if _, err := os.Stat(filepath.Join(destDir, "php8ts.dll")); err == nil {
		t.Errorf("expected leftover php8ts.dll to be removed after upgrade")
	}
	if _, err := os.Stat(filepath.Join(destDir, "php8apache2_4.dll")); err == nil {
		t.Errorf("expected leftover php8apache2_4.dll to be removed after upgrade")
	}
	if _, err := os.Stat(filepath.Join(destDir, "ext", "php_opcache.dll")); err == nil {
		t.Errorf("expected leftover ext/php_opcache.dll to be removed after upgrade")
	}
	if _, err := os.Stat(filepath.Join(destDir, "stale_file.txt")); err == nil {
		t.Errorf("expected leftover stale_file.txt to be removed after upgrade")
	}
}

func TestInstallArtifactZip_PhpIniPreserved(t *testing.T) {
	tempBase := t.TempDir()
	destDir := filepath.Join(tempBase, "bin", "php-8.5")
	if err := os.MkdirAll(destDir, 0755); err != nil {
		t.Fatal(err)
	}

	_ = os.WriteFile(filepath.Join(destDir, "php.exe"), []byte("old-php"), 0755)
	_ = os.WriteFile(filepath.Join(destDir, "php-cgi.exe"), []byte("old-php-cgi"), 0755)
	oldIni := `[PHP]
memory_limit = 1024M
zend_extension=opcache
[opcache]
opcache.enable = 1
opcache.memory_consumption = 256
`
	_ = os.WriteFile(filepath.Join(destDir, "php.ini"), []byte(oldIni), 0644)

	zipPath := filepath.Join(tempBase, "php-8.5.zip")
	createTestZip(t, zipPath, map[string]string{
		"php.exe":             "new-php-8.5",
		"php-cgi.exe":         "new-php-cgi-8.5",
		"php.ini-development": "[PHP]\nmemory_limit = 128M\n",
	})

	err := downloader.InstallArtifactZip("php-8.5", zipPath, destDir)
	if err != nil {
		t.Fatalf("InstallArtifactZip failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(destDir, "php.ini"))
	if err != nil {
		t.Fatalf("failed to read preserved php.ini: %v", err)
	}
	content := string(data)

	// User setting memory_limit = 1024M must be preserved
	if !strings.Contains(content, "memory_limit = 1024M") {
		t.Errorf("expected custom memory_limit = 1024M to be preserved, got:\n%s", content)
	}
	// For PHP 8.5, zend_extension=opcache must be commented out
	if !strings.Contains(content, ";zend_extension=opcache") {
		t.Errorf("expected ;zend_extension=opcache to be commented out in PHP 8.5")
	}
	// opcache settings must be preserved
	if !strings.Contains(content, "opcache.memory_consumption = 256") {
		t.Errorf("expected custom opcache.memory_consumption = 256 to be preserved")
	}
}

func TestInstallArtifactZip_FailureMidwayRestoresOldFolder(t *testing.T) {
	tempBase := t.TempDir()
	destDir := filepath.Join(tempBase, "bin", "caddy")
	if err := os.MkdirAll(destDir, 0755); err != nil {
		t.Fatal(err)
	}

	_ = os.WriteFile(filepath.Join(destDir, "caddy.exe"), []byte("original-caddy-bin"), 0755)
	_ = os.WriteFile(filepath.Join(destDir, "my_custom_config.json"), []byte(`{"preserved":true}`), 0644)

	// Corrupted zip that does NOT contain caddy.exe (verification must fail)
	badZip := filepath.Join(tempBase, "caddy-bad.zip")
	createTestZip(t, badZip, map[string]string{
		"wrong-executable.exe": "some-data",
	})

	err := downloader.InstallArtifactZip("caddy", badZip, destDir)
	if err == nil {
		t.Fatal("expected InstallArtifactZip to fail verification on missing caddy.exe, got nil")
	}

	// Verify original destDir was restored and intact
	origBin, err := os.ReadFile(filepath.Join(destDir, "caddy.exe"))
	if err != nil || string(origBin) != "original-caddy-bin" {
		t.Errorf("original caddy.exe not restored properly, got: %s (err: %v)", string(origBin), err)
	}
	cfgData, err := os.ReadFile(filepath.Join(destDir, "my_custom_config.json"))
	if err != nil || string(cfgData) != `{"preserved":true}` {
		t.Errorf("original custom config not restored properly, got: %s (err: %v)", string(cfgData), err)
	}

	// Verify no staging or .old folders remain
	parentDir := filepath.Dir(destDir)
	entries, _ := os.ReadDir(parentDir)
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".staging-") || strings.HasPrefix(entry.Name(), ".old-") {
			t.Errorf("found leftover staging/old dir: %s", entry.Name())
		}
	}
}

func TestInstallArtifactZip_MariaDBDataDirUntouched_InsideBin(t *testing.T) {
	tempBase := t.TempDir()
	destDir := filepath.Join(tempBase, "bin", "mysql")
	dataDir := filepath.Join(destDir, "data", "shop_db")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		t.Fatal(err)
	}
	_ = os.MkdirAll(filepath.Join(destDir, "bin"), 0755)

	_ = os.WriteFile(filepath.Join(destDir, "bin", "mysqld.exe"), []byte("old-mysqld"), 0755)
	_ = os.WriteFile(filepath.Join(destDir, "my.ini"), []byte("port=3306\n"), 0644)
	_ = os.WriteFile(filepath.Join(dataDir, "users.ibd"), []byte("user-database-content-12345"), 0644)

	// New MariaDB zip
	zipPath := filepath.Join(tempBase, "mariadb.zip")
	createTestZip(t, zipPath, map[string]string{
		"bin/mysqld.exe": "new-mysqld-binary",
		"bin/mysql.exe":  "new-client-binary",
	})

	err := downloader.InstallArtifactZip("mysql", zipPath, destDir)
	if err != nil {
		t.Fatalf("InstallArtifactZip failed: %v", err)
	}

	// Check new mysqld.exe
	newMysqld, err := os.ReadFile(filepath.Join(destDir, "bin", "mysqld.exe"))
	if err != nil || string(newMysqld) != "new-mysqld-binary" {
		t.Errorf("expected new mysqld binary, got: %s", string(newMysqld))
	}

	// Check my.ini preserved
	iniData, err := os.ReadFile(filepath.Join(destDir, "my.ini"))
	if err != nil || !strings.Contains(string(iniData), "port=3306") {
		t.Errorf("my.ini was lost or not preserved")
	}

	// Check data directory was NEVER touched or deleted
	dbFile, err := os.ReadFile(filepath.Join(destDir, "data", "shop_db", "users.ibd"))
	if err != nil || string(dbFile) != "user-database-content-12345" {
		t.Fatalf("CRITICAL: MariaDB data directory was corrupted or deleted during upgrade! got: %s, err: %v", string(dbFile), err)
	}
}

func TestInstallArtifactZip_MariaDBDataDirUntouched_OutsideBin(t *testing.T) {
	tempBase := t.TempDir()
	destDir := filepath.Join(tempBase, "bin", "mysql")
	externalDataDir := filepath.Join(tempBase, "data", "mysql", "production_db")
	if err := os.MkdirAll(externalDataDir, 0755); err != nil {
		t.Fatal(err)
	}
	_ = os.MkdirAll(filepath.Join(destDir, "bin"), 0755)

	_ = os.WriteFile(filepath.Join(destDir, "bin", "mysqld.exe"), []byte("old-mysqld"), 0755)
	_ = os.WriteFile(filepath.Join(externalDataDir, "orders.ibd"), []byte("production-orders-9999"), 0644)

	zipPath := filepath.Join(tempBase, "mariadb.zip")
	createTestZip(t, zipPath, map[string]string{
		"bin/mysqld.exe": "new-mysqld-binary",
	})

	err := downloader.InstallArtifactZip("mysql", zipPath, destDir)
	if err != nil {
		t.Fatalf("InstallArtifactZip failed: %v", err)
	}

	// External data directory must be completely untouched
	dbFile, err := os.ReadFile(filepath.Join(externalDataDir, "orders.ibd"))
	if err != nil || string(dbFile) != "production-orders-9999" {
		t.Fatalf("CRITICAL: External MariaDB data directory was affected during upgrade! got: %s, err: %v", string(dbFile), err)
	}
}

func TestInstallArtifactZip_VerificationRules(t *testing.T) {
	tempBase := t.TempDir()

	// 1. PHP missing php-cgi.exe and php.exe
	phpZip := filepath.Join(tempBase, "invalid-php.zip")
	createTestZip(t, phpZip, map[string]string{"random.dll": "abc"})
	destPhp := filepath.Join(tempBase, "bin", "php-8.4")
	if err := downloader.InstallArtifactZip("php-8.4", phpZip, destPhp); err == nil {
		t.Errorf("expected error for PHP missing php-cgi.exe")
	}

	// 2. MySQL missing mysqld.exe
	mysqlZip := filepath.Join(tempBase, "invalid-mysql.zip")
	createTestZip(t, mysqlZip, map[string]string{"bin/mysql.exe": "client-only"})
	destMysql := filepath.Join(tempBase, "bin", "mysql")
	if err := downloader.InstallArtifactZip("mysql", mysqlZip, destMysql); err == nil {
		t.Errorf("expected error for MySQL missing mysqld.exe")
	}

	// 3. Redis missing redis-server.exe
	redisZip := filepath.Join(tempBase, "invalid-redis.zip")
	createTestZip(t, redisZip, map[string]string{"redis-cli.exe": "cli-only"})
	destRedis := filepath.Join(tempBase, "bin", "redis")
	if err := downloader.InstallArtifactZip("redis", redisZip, destRedis); err == nil {
		t.Errorf("expected error for Redis missing redis-server.exe")
	}

	// 4. Mailpit missing mailpit.exe
	mailpitZip := filepath.Join(tempBase, "invalid-mailpit.zip")
	createTestZip(t, mailpitZip, map[string]string{"other.exe": "other"})
	destMailpit := filepath.Join(tempBase, "bin", "mailpit")
	if err := downloader.InstallArtifactZip("mailpit", mailpitZip, destMailpit); err == nil {
		t.Errorf("expected error for Mailpit missing mailpit.exe")
	}
}

// Injected failure tests to guarantee user data files are NEVER deleted on failures

func TestInstallArtifactZip_MariaDBDataPreserved_WhenDestDirRenameFails(t *testing.T) {
	tempBase := t.TempDir()
	destDir := filepath.Join(tempBase, "bin", "mysql")
	dataDir := filepath.Join(destDir, "data", "shop_db")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		t.Fatal(err)
	}
	_ = os.MkdirAll(filepath.Join(destDir, "bin"), 0755)

	_ = os.WriteFile(filepath.Join(destDir, "bin", "mysqld.exe"), []byte("old-mysqld"), 0755)
	_ = os.WriteFile(filepath.Join(dataDir, "users.ibd"), []byte("user-data-preserve-test-1"), 0644)

	zipPath := filepath.Join(tempBase, "mariadb.zip")
	createTestZip(t, zipPath, map[string]string{
		"bin/mysqld.exe": "new-mysqld-binary",
	})

	// Inject rename that fails when renaming destDir -> oldDir
	failingRename := func(oldpath, newpath string) error {
		if strings.Contains(newpath, ".old-") {
			return errors.New("simulated failure renaming destDir to .old")
		}
		return os.Rename(oldpath, newpath)
	}

	err := downloader.InstallArtifactZipWithRename("mysql", zipPath, destDir, failingRename)
	if err == nil {
		t.Fatal("expected InstallArtifactZipWithRename to fail")
	}

	// Database file must still exist and be intact (restored to destDir/data)
	dbFile, readErr := os.ReadFile(filepath.Join(destDir, "data", "shop_db", "users.ibd"))
	if readErr != nil || string(dbFile) != "user-data-preserve-test-1" {
		t.Fatalf("CRITICAL: user data was lost on destDir rename failure! got: %s (err: %v)", string(dbFile), readErr)
	}
}

func TestInstallArtifactZip_MariaDBDataPreserved_WhenRollbackRenameFails(t *testing.T) {
	tempBase := t.TempDir()
	destDir := filepath.Join(tempBase, "bin", "mysql")
	dataDir := filepath.Join(destDir, "data", "shop_db")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		t.Fatal(err)
	}
	_ = os.MkdirAll(filepath.Join(destDir, "bin"), 0755)

	_ = os.WriteFile(filepath.Join(destDir, "bin", "mysqld.exe"), []byte("old-mysqld"), 0755)
	_ = os.WriteFile(filepath.Join(dataDir, "users.ibd"), []byte("user-data-preserve-test-2"), 0644)

	zipPath := filepath.Join(tempBase, "mariadb.zip")
	createTestZip(t, zipPath, map[string]string{
		"bin/mysqld.exe": "new-mysqld-binary",
	})

	// Inject rename that fails on destDir -> oldDir AND also fails on rollback rename of data
	var capturedSiblingDir string
	failingRename := func(oldpath, newpath string) error {
		if strings.Contains(newpath, ".mysql-data-") {
			capturedSiblingDir = newpath
			return os.Rename(oldpath, newpath)
		}
		if strings.Contains(newpath, ".old-") {
			return errors.New("simulated failure renaming destDir to .old")
		}
		if strings.Contains(oldpath, ".mysql-data-") && strings.HasSuffix(newpath, "data") {
			return errors.New("simulated failure on rollback rename of data directory")
		}
		return os.Rename(oldpath, newpath)
	}

	err := downloader.InstallArtifactZipWithRename("mysql", zipPath, destDir, failingRename)
	if err == nil {
		t.Fatal("expected InstallArtifactZipWithRename to fail")
	}

	// The error message MUST name the sibling folder containing preserved data
	if !strings.Contains(err.Error(), "database files preserved at") {
		t.Errorf("expected error to mention 'database files preserved at', got: %v", err)
	}

	// The database file MUST still exist in the sibling directory
	if capturedSiblingDir == "" {
		t.Fatal("capturedSiblingDir is empty")
	}
	dbFile, readErr := os.ReadFile(filepath.Join(capturedSiblingDir, "shop_db", "users.ibd"))
	if readErr != nil || string(dbFile) != "user-data-preserve-test-2" {
		t.Fatalf("CRITICAL: user data was deleted when rollback rename failed! got: %s (err: %v)", string(dbFile), readErr)
	}
}

func TestInstallArtifactZip_MariaDBDataPreserved_WhenFinalDataMoveFails(t *testing.T) {
	tempBase := t.TempDir()
	destDir := filepath.Join(tempBase, "bin", "mysql")
	dataDir := filepath.Join(destDir, "data", "shop_db")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		t.Fatal(err)
	}
	_ = os.MkdirAll(filepath.Join(destDir, "bin"), 0755)

	_ = os.WriteFile(filepath.Join(destDir, "bin", "mysqld.exe"), []byte("old-mysqld"), 0755)
	_ = os.WriteFile(filepath.Join(dataDir, "users.ibd"), []byte("user-data-preserve-test-3"), 0644)

	zipPath := filepath.Join(tempBase, "mariadb.zip")
	createTestZip(t, zipPath, map[string]string{
		"bin/mysqld.exe": "new-mysqld-binary",
	})

	// Inject rename that succeeds on swap but fails when moving savedDataDir into new destDir
	var capturedSiblingDir string
	failingRename := func(oldpath, newpath string) error {
		if strings.Contains(newpath, ".mysql-data-") {
			capturedSiblingDir = newpath
			return os.Rename(oldpath, newpath)
		}
		if strings.Contains(oldpath, ".mysql-data-") && strings.HasSuffix(newpath, "data") {
			return errors.New("simulated failure moving data directory into newly upgraded runtime")
		}
		return os.Rename(oldpath, newpath)
	}

	err := downloader.InstallArtifactZipWithRename("mysql", zipPath, destDir, failingRename)
	if err == nil {
		t.Fatal("expected InstallArtifactZipWithRename to fail")
	}

	// The error message MUST name the sibling folder containing preserved data
	if !strings.Contains(err.Error(), "database files preserved at") {
		t.Errorf("expected error to mention 'database files preserved at', got: %v", err)
	}

	// Data must still exist in the sibling directory
	dbFile, readErr := os.ReadFile(filepath.Join(capturedSiblingDir, "shop_db", "users.ibd"))
	if readErr != nil || string(dbFile) != "user-data-preserve-test-3" {
		t.Fatalf("CRITICAL: user data was deleted when final data move failed! got: %s (err: %v)", string(dbFile), readErr)
	}
}

func TestInstallArtifactZip_DeferredCleanup_NeverDeletesDataInStaging(t *testing.T) {
	tempBase := t.TempDir()
	destDir := filepath.Join(tempBase, "bin", "mysql")
	if err := os.MkdirAll(destDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Zip containing data directory
	zipPath := filepath.Join(tempBase, "mariadb-with-data.zip")
	createTestZip(t, zipPath, map[string]string{
		"bin/mysqld.exe":   "mysqld-bin",
		"data/crucial.db": "precious-db-data",
	})

	// Inject rename that fails during installation after zip extraction
	failingRename := func(oldpath, newpath string) error {
		return errors.New("simulated error during rename")
	}

	err := downloader.InstallArtifactZipWithRename("mysql", zipPath, destDir, failingRename)
	if err == nil {
		t.Fatal("expected InstallArtifactZipWithRename to fail")
	}

	// If staging has data/, the error MUST state "database files preserved at <path>"
	if !strings.Contains(err.Error(), "database files preserved at") {
		t.Errorf("expected error to state 'database files preserved at', got: %v", err)
	}

	// Verify the staging directory with data was NOT deleted by deferred cleanup
	parentDir := filepath.Dir(destDir)
	entries, _ := os.ReadDir(parentDir)
	foundStagingWithData := false
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".staging-") {
			stagingData := filepath.Join(parentDir, entry.Name(), "data")
			if _, statErr := os.Stat(stagingData); statErr == nil {
				foundStagingWithData = true
				content, _ := os.ReadFile(filepath.Join(stagingData, "crucial.db"))
				if string(content) != "precious-db-data" {
					t.Errorf("crucial.db corrupted, got: %s", string(content))
				}
			}
		}
	}

	if !foundStagingWithData {
		t.Error("CRITICAL: staging directory containing data was deleted by deferred cleanup!")
	}
}
