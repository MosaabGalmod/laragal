//go:build !windows

package winutil

import "os/exec"

// HideWindow is a no-op on non-Windows platforms
func HideWindow(cmd *exec.Cmd) *exec.Cmd {
	return cmd
}

// Command creates a standard exec.Cmd on non-Windows platforms
func Command(name string, arg ...string) *exec.Cmd {
	return exec.Command(name, arg...)
}
