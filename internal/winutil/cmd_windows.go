//go:build windows

package winutil

import (
	"os/exec"
	"syscall"
)

const CreateNoWindow = 0x08000000

// HideWindow configures an exec.Cmd so it never spawns or flashes a console window on Windows
func HideWindow(cmd *exec.Cmd) *exec.Cmd {
	if cmd == nil {
		return nil
	}
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.HideWindow = true
	cmd.SysProcAttr.CreationFlags |= CreateNoWindow
	return cmd
}

// Command creates an exec.Cmd pre-configured to run with no console window
func Command(name string, arg ...string) *exec.Cmd {
	cmd := exec.Command(name, arg...)
	return HideWindow(cmd)
}
