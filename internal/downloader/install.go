package downloader

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/laragal/laragal/internal/services/php"
)

// RenameFunc defines the signature for renaming files and directories, allowing dependency injection in tests.
type RenameFunc func(oldpath, newpath string) error

func sanitizeArtifactName(name string) string {
	clean := strings.ToLower(strings.TrimSpace(name))
	clean = strings.ReplaceAll(clean, "/", "_")
	clean = strings.ReplaceAll(clean, "\\", "_")
	clean = strings.ReplaceAll(clean, "..", "")
	return clean
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}

// verifyStagingArtifact checks that the extracted staging directory contains the expected main executable.
func verifyStagingArtifact(name, stagingDir string) error {
	norm := sanitizeArtifactName(name)

	switch {
	case strings.HasPrefix(norm, "php"):
		phpCgi := filepath.Join(stagingDir, "php-cgi.exe")
		phpExe := filepath.Join(stagingDir, "php.exe")
		cgiExists := fileExists(phpCgi)
		exeExists := fileExists(phpExe)
		if !cgiExists && !exeExists {
			return fmt.Errorf("staging verification failed for %s: missing expected main executable (php-cgi.exe / php.exe)", name)
		}
	case norm == "caddy":
		if !fileExists(filepath.Join(stagingDir, "caddy.exe")) {
			return fmt.Errorf("staging verification failed for caddy: missing expected caddy.exe")
		}
	case norm == "mysql" || norm == "mariadb":
		mysqldInBin := fileExists(filepath.Join(stagingDir, "bin", "mysqld.exe"))
		mysqldInRoot := fileExists(filepath.Join(stagingDir, "mysqld.exe"))
		if !mysqldInBin && !mysqldInRoot {
			return fmt.Errorf("staging verification failed for %s: missing expected mysqld.exe (in bin/ or root)", name)
		}
	case norm == "mailpit":
		if !fileExists(filepath.Join(stagingDir, "mailpit.exe")) {
			return fmt.Errorf("staging verification failed for mailpit: missing expected mailpit.exe")
		}
	case norm == "redis":
		exeInRoot := fileExists(filepath.Join(stagingDir, "redis-server.exe"))
		exeInSub := false
		if !exeInRoot {
			if matches, _ := filepath.Glob(filepath.Join(stagingDir, "*", "redis-server.exe")); len(matches) > 0 {
				exeInSub = true
			}
		}
		if !exeInRoot && !exeInSub {
			return fmt.Errorf("staging verification failed for redis: missing expected redis-server.exe")
		}
	case norm == "phpmyadmin":
		if !fileExists(filepath.Join(stagingDir, "index.php")) {
			return fmt.Errorf("staging verification failed for phpmyadmin: missing expected index.php")
		}
	}

	return nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// InstallArtifactZip safely installs or upgrades a runtime artifact from a zip archive.
// It extracts into an isolated staging directory, verifies the main executable, preserves user
// config files (php.ini, my.ini, redis.conf) and ensures MariaDB data is never touched or lost,
// performs an atomic replacement, and rolls back to the original version on any failure.
func InstallArtifactZip(name string, zipPath string, destDir string) error {
	return InstallArtifactZipWithRename(name, zipPath, destDir, os.Rename)
}

// InstallArtifactZipWithRename installs or upgrades an artifact using a configurable rename function.
// If destDir/data exists for MariaDB, it is moved to a sibling directory outside staging first.
// On any failure, data remains preserved in the sibling folder and the error reports its path.
func InstallArtifactZipWithRename(name string, zipPath string, destDir string, rename RenameFunc) (retErr error) {
	if rename == nil {
		rename = os.Rename
	}

	parentDir := filepath.Dir(destDir)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	safeName := sanitizeArtifactName(name)
	stagingDir := filepath.Join(parentDir, fmt.Sprintf(".staging-%s-%x", safeName, rand.Int63()))

	success := false
	defer func() {
		if !success {
			// Rule: user data must NEVER be deleted by cleanup.
			stagingData := filepath.Join(stagingDir, "data")
			if dirExists(stagingData) {
				msg := fmt.Sprintf("database files preserved at %s", stagingData)
				if retErr == nil {
					retErr = fmt.Errorf("installation failed; %s", msg)
				} else if !strings.Contains(retErr.Error(), stagingData) {
					retErr = fmt.Errorf("%w; %s", retErr, msg)
				}
				return
			}
			_ = os.RemoveAll(stagingDir)
		}
	}()

	// 1. Extract zip into fresh staging directory
	if err := ExtractZip(zipPath, stagingDir); err != nil {
		return fmt.Errorf("failed to extract artifact zip to staging directory: %w", err)
	}

	// 2. Verify staging contains expected executable
	if err := verifyStagingArtifact(name, stagingDir); err != nil {
		return err
	}

	destExists := dirExists(destDir)

	if !destExists {
		// New installation
		if strings.HasPrefix(safeName, "php") {
			_ = php.EnsurePHPIni(stagingDir, name)
		}
		if err := rename(stagingDir, destDir); err != nil {
			return fmt.Errorf("failed to move staging directory into place: %w", err)
		}
		success = true
		return nil
	}

	// 3. Upgrade / Reinstall existing runtime: preserve user configs and data
	var savedDataDir string

	// A. Preserve PHP config
	if strings.HasPrefix(safeName, "php") {
		oldIni := filepath.Join(destDir, "php.ini")
		if fileExists(oldIni) {
			_ = copyFile(oldIni, filepath.Join(stagingDir, "php.ini"))
		}
		_ = php.EnsurePHPIni(stagingDir, name)
	}

	// B. Preserve MariaDB / MySQL config and data directory
	if safeName == "mysql" || safeName == "mariadb" {
		oldMyIni := filepath.Join(destDir, "my.ini")
		if fileExists(oldMyIni) {
			_ = copyFile(oldMyIni, filepath.Join(stagingDir, "my.ini"))
		}
		oldBinMyIni := filepath.Join(destDir, "bin", "my.ini")
		if fileExists(oldBinMyIni) {
			_ = copyFile(oldBinMyIni, filepath.Join(stagingDir, "bin", "my.ini"))
		}

		oldDataDir := filepath.Join(destDir, "data")
		if dirExists(oldDataDir) {
			// Move data directory to a sibling folder OUTSIDE staging FIRST
			siblingData := filepath.Join(parentDir, fmt.Sprintf(".mysql-data-%s-%d", safeName, time.Now().UnixNano()))
			if err := rename(oldDataDir, siblingData); err != nil {
				return fmt.Errorf("refusing upgrade: failed to safely preserve existing MariaDB data directory: %w", err)
			}
			savedDataDir = siblingData

			// Ensure staging does not have a conflicting empty data folder from the zip
			stagingData := filepath.Join(stagingDir, "data")
			if dirExists(stagingData) {
				_ = os.RemoveAll(stagingData)
			}
		}
	}

	// C. Preserve Redis config (supports both redis.conf and legacy redis.windows.conf)
	if safeName == "redis" {
		for _, confName := range []string{"redis.conf", "redis.windows.conf"} {
			oldConf := filepath.Join(destDir, confName)
			if fileExists(oldConf) {
				_ = copyFile(oldConf, filepath.Join(stagingDir, confName))
			}
			if matches, err := filepath.Glob(filepath.Join(destDir, "*", confName)); err == nil && len(matches) > 0 {
				_ = copyFile(matches[0], filepath.Join(stagingDir, confName))
			}
		}
	}

	// 4. Atomic swap with rollback
	oldDir := filepath.Join(parentDir, fmt.Sprintf(".old-%s-%d", safeName, time.Now().UnixNano()))

	// Step 4a: rename destDir -> oldDir
	if err := rename(destDir, oldDir); err != nil {
		if savedDataDir != "" {
			oldDataDir := filepath.Join(destDir, "data")
			if rErr := rename(savedDataDir, oldDataDir); rErr != nil {
				return fmt.Errorf("failed to rename existing runtime to backup folder (%w); rollback of data directory failed (database files preserved at %s): %v", err, savedDataDir, rErr)
			}
		}
		return fmt.Errorf("failed to rename existing runtime to backup folder: %w", err)
	}

	// Step 4b: rename stagingDir -> destDir
	if err := rename(stagingDir, destDir); err != nil {
		var rollbackErrs []string
		_ = os.RemoveAll(destDir)
		if rErr := rename(oldDir, destDir); rErr != nil {
			rollbackErrs = append(rollbackErrs, fmt.Sprintf("failed to restore old folder from %s: %v", oldDir, rErr))
		} else if savedDataDir != "" {
			oldDataDir := filepath.Join(destDir, "data")
			if dErr := rename(savedDataDir, oldDataDir); dErr != nil {
				rollbackErrs = append(rollbackErrs, fmt.Sprintf("failed to restore data directory to %s (database files preserved at %s): %v", oldDataDir, savedDataDir, dErr))
			}
		}
		if len(rollbackErrs) > 0 {
			if savedDataDir != "" && dirExists(savedDataDir) {
				rollbackErrs = append(rollbackErrs, fmt.Sprintf("database files preserved at %s", savedDataDir))
			}
			return fmt.Errorf("failed to move staging runtime into place (%w); rollback errors: %s", err, strings.Join(rollbackErrs, "; "))
		}
		return fmt.Errorf("failed to move staging runtime into place (rolled back to original): %w", err)
	}

	// Step 4c: move preserved data directory into the new destDir
	if savedDataDir != "" {
		newDestData := filepath.Join(destDir, "data")
		if err := rename(savedDataDir, newDestData); err != nil {
			// On any failure, data stays in savedDataDir and error names it!
			return fmt.Errorf("runtime upgrade placed new binary but failed to move database files into new runtime; database files preserved at %s: %w", savedDataDir, err)
		}
	}

	// 5. Delete backup folder
	if strings.HasPrefix(filepath.Base(oldDir), ".old-") {
		_ = os.RemoveAll(oldDir)
	}

	success = true
	return nil
}
