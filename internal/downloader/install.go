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
		if !fileExists(filepath.Join(stagingDir, "redis-server.exe")) {
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
	parentDir := filepath.Dir(destDir)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	safeName := sanitizeArtifactName(name)
	stagingDir := filepath.Join(parentDir, fmt.Sprintf(".staging-%s-%x", safeName, rand.Int63()))

	success := false
	defer func() {
		if !success {
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
		if err := os.Rename(stagingDir, destDir); err != nil {
			return fmt.Errorf("failed to move staging directory into place: %w", err)
		}
		success = true
		return nil
	}

	// 3. Upgrade / Reinstall existing runtime: preserve user configs and data
	movedDataDir := false

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
			stagingDataDir := filepath.Join(stagingDir, "data")
			if dirExists(stagingDataDir) {
				_ = os.RemoveAll(stagingDataDir)
			}
			if err := os.Rename(oldDataDir, stagingDataDir); err != nil {
				return fmt.Errorf("refusing upgrade: failed to safely move existing MariaDB data directory: %w", err)
			}
			movedDataDir = true
		}
	}

	// C. Preserve Redis config
	if safeName == "redis" {
		oldRedisConf := filepath.Join(destDir, "redis.conf")
		if fileExists(oldRedisConf) {
			_ = copyFile(oldRedisConf, filepath.Join(stagingDir, "redis.conf"))
		}
	}

	// 4. Atomic swap with rollback
	oldDir := filepath.Join(parentDir, fmt.Sprintf(".old-%s-%d", safeName, time.Now().UnixNano()))

	if err := os.Rename(destDir, oldDir); err != nil {
		if movedDataDir {
			_ = os.Rename(filepath.Join(stagingDir, "data"), filepath.Join(destDir, "data"))
		}
		return fmt.Errorf("failed to rename existing runtime to backup folder: %w", err)
	}

	if err := os.Rename(stagingDir, destDir); err != nil {
		// Roll back
		_ = os.RemoveAll(destDir)
		restoreErr := os.Rename(oldDir, destDir)
		if movedDataDir {
			if _, statErr := os.Stat(filepath.Join(stagingDir, "data")); statErr == nil {
				_ = os.Rename(filepath.Join(stagingDir, "data"), filepath.Join(destDir, "data"))
			}
		}
		return fmt.Errorf("failed to move staging runtime into place (rollback status: %v): %w", restoreErr, err)
	}

	// 5. Delete backup folder
	if strings.HasPrefix(filepath.Base(oldDir), ".old-") {
		_ = os.RemoveAll(oldDir)
	}

	success = true
	return nil
}
