//go:build windows

package winutil

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

const runRegistryKey = `Software\Microsoft\Windows\CurrentVersion\Run`

// IsStartOnBoot checks whether the given application is configured to start with Windows.
func IsStartOnBoot(appName string) bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, runRegistryKey, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer k.Close()

	val, _, err := k.GetStringValue(appName)
	return err == nil && val != ""
}

// SetStartOnBoot enables or disables auto-starting the application with Windows.
func SetStartOnBoot(appName string, exePath string, enabled bool) error {
	if enabled {
		if exePath == "" {
			var err error
			exePath, err = os.Executable()
			if err != nil {
				return fmt.Errorf("failed to determine executable path: %w", err)
			}
		}
		absPath, err := filepath.Abs(exePath)
		if err == nil {
			exePath = absPath
		}

		k, err := registry.OpenKey(registry.CURRENT_USER, runRegistryKey, registry.SET_VALUE)
		if err != nil {
			return fmt.Errorf("failed to open run registry key: %w", err)
		}
		defer k.Close()

		command := fmt.Sprintf("\"%s\"", exePath)
		if err := k.SetStringValue(appName, command); err != nil {
			return fmt.Errorf("failed to set run registry value: %w", err)
		}
		return nil
	}

	// Disable auto-start
	k, err := registry.OpenKey(registry.CURRENT_USER, runRegistryKey, registry.SET_VALUE)
	if err != nil {
		return nil
	}
	defer k.Close()

	err = k.DeleteValue(appName)
	if err != nil && err != registry.ErrNotExist {
		return fmt.Errorf("failed to remove run registry value: %w", err)
	}
	return nil
}
