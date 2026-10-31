//go:build !windows

package winutil

import "os/exec"

func HideWindow(cmd *exec.Cmd) *exec.Cmd {
	return cmd
}

func Command(name string, arg ...string) *exec.Cmd {
	return exec.Command(name, arg...)
}

func KillProcessTree(pid int) error {
	return nil
}

func KillProcessesOnPort(port int) error {
	return nil
}
