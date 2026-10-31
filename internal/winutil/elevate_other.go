//go:build !windows

package winutil

// IsAdmin returns true on non-Windows platforms.
func IsAdmin() bool {
	return true
}

// RelaunchAsAdmin is a stub on non-Windows platforms.
func RelaunchAsAdmin() error {
	return nil
}
