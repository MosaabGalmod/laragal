//go:build !windows

package winutil

// IsStartOnBoot is a stub for non-Windows platforms.
func IsStartOnBoot(appName string) bool {
	return false
}

// SetStartOnBoot is a stub for non-Windows platforms.
func SetStartOnBoot(appName string, exePath string, enabled bool) error {
	return nil
}
