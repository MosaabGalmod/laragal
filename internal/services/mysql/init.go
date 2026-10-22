package mysql

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/laragal/laragal/internal/winutil"
)

// IsDataDirInitialized checks whether dataDir already contains database engine files
func IsDataDirInitialized(dataDir string) bool {
	markers := []string{
		filepath.Join(dataDir, "mysql"),
		filepath.Join(dataDir, "ibdata1"),
		filepath.Join(dataDir, "aria_log_control"),
		filepath.Join(dataDir, "auto.cnf"),
	}
	for _, m := range markers {
		if _, err := os.Stat(m); err == nil {
			return true
		}
	}
	return false
}

// FindInstallerCandidate checks for mariadb-install-db.exe or mysql_install_db.exe in the binary directory
func FindInstallerCandidate(binDir string) string {
	candidates := []string{
		filepath.Join(binDir, "mariadb-install-db.exe"),
		filepath.Join(binDir, "mysql_install_db.exe"),
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return ""
}

// EnsureDataDir initializes the MySQL/MariaDB data directory if it is not already initialized
func EnsureDataDir(mysqldExe string, dataDir string) error {
	absDataDir, err := filepath.Abs(dataDir)
	if err != nil {
		absDataDir = dataDir
	}

	if IsDataDirInitialized(absDataDir) {
		return nil
	}

	if err := os.MkdirAll(absDataDir, 0755); err != nil {
		return fmt.Errorf("failed to create mysql data directory: %w", err)
	}

	binDir := filepath.Dir(mysqldExe)
	installer := FindInstallerCandidate(binDir)

	var cmdArgs []string
	if installer != "" {
		cmdArgs = []string{installer, fmt.Sprintf("--datadir=%s", absDataDir)}
	} else {
		cmdArgs = []string{mysqldExe, "--initialize-insecure", fmt.Sprintf("--datadir=%s", absDataDir), "--console"}
	}

	cmd := winutil.Command(cmdArgs[0], cmdArgs[1:]...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to initialize mysql data directory (%v): %s", err, string(out))
	}

	return nil
}
