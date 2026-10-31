package mysql

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/laragal/laragal/internal/winutil"
)

// IsDataDirInitialized checks whether dataDir already contains database engine files
func IsDataDirInitialized(dataDir string) bool {
	markers := []string{
		filepath.Join(dataDir, "mysql"),
		filepath.Join(dataDir, "ibdata1"),
		filepath.Join(dataDir, "aria_log_control"),
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
	baseDir := filepath.Dir(binDir)

	// Clean out any leftover non-db files (e.g. empty my.ini)
	if entries, err := os.ReadDir(absDataDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() {
				_ = os.Remove(filepath.Join(absDataDir, e.Name()))
			}
		}
	}

	// Bootstrap MariaDB using mysqld --bootstrap with system tables
	var sqlParts []string
	sqlParts = append(sqlParts, "CREATE DATABASE IF NOT EXISTS mysql;\nUSE mysql;\n")

	shareDir := filepath.Join(baseDir, "share")
	for _, tableFile := range []string{"mysql_system_tables.sql", "mysql_system_tables_data.sql", "mysql_performance_tables.sql"} {
		p := filepath.Join(shareDir, tableFile)
		if data, err := os.ReadFile(p); err == nil {
			sqlParts = append(sqlParts, string(data))
		}
	}

	// Ensure root has full access without password and create default dev databases
	sqlParts = append(sqlParts, `
CREATE DATABASE IF NOT EXISTS `+"`laragal`"+`;
GRANT ALL PRIVILEGES ON *.* TO 'root'@'127.0.0.1' WITH GRANT OPTION;
GRANT ALL PRIVILEGES ON *.* TO 'root'@'localhost' WITH GRANT OPTION;
FLUSH PRIVILEGES;
`)

	fullSQL := strings.Join(sqlParts, "\n")

	cmd := winutil.Command(mysqldExe,
		fmt.Sprintf("--basedir=%s", baseDir),
		fmt.Sprintf("--datadir=%s", absDataDir),
		"--bootstrap",
		"--console",
	)
	cmd.Stdin = strings.NewReader(fullSQL)
	out, err := cmd.CombinedOutput()
	if err != nil {
		// Fallback: try installer binary if present
		installer := FindInstallerCandidate(binDir)
		if installer != "" {
			altCmd := winutil.Command(installer, fmt.Sprintf("--datadir=%s", absDataDir))
			if _, altErr := altCmd.CombinedOutput(); altErr == nil {
				return nil
			}
		}
		return fmt.Errorf("failed to initialize mysql data directory (%v): %s", err, string(out))
	}

	return nil
}
