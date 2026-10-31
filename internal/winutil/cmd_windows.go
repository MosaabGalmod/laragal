//go:build windows

package winutil

import (
	"fmt"
	"os/exec"
	"strconv"
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

// KillProcessTree forcefully terminates a process and all of its child processes using taskkill /F /T /PID <pid>
func KillProcessTree(pid int) error {
	if pid <= 0 {
		return nil
	}
	cmd := Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(pid))
	return cmd.Run()
}

// KillProcessesOnPort terminates any process and its child processes currently listening on the given TCP port
func KillProcessesOnPort(port int) error {
	if port <= 0 {
		return nil
	}
	cmd := Command("netstat", "-ano", "-p", "TCP")
	out, err := cmd.Output()
	if err == nil {
		pids := ParseListeningPIDs(string(out), port)
		for _, pid := range pids {
			_ = KillProcessTree(pid)
		}
		return nil
	}

	// Fallback to powershell only if netstat command itself failed
	psScript := fmt.Sprintf("Get-NetTCPConnection -LocalPort %d -ErrorAction SilentlyContinue | Select-Object -ExpandProperty OwningProcess -Unique | ForEach-Object { if ($_ -and $_ -gt 0) { taskkill /F /T /PID $_ 2>$null } }", port)
	fallbackCmd := Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psScript)
	return fallbackCmd.Run()
}
