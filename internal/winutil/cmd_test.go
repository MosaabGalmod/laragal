package winutil

import (
	"os/exec"
	"runtime"
	"testing"
)

func TestHideWindow(t *testing.T) {
	cmd := Command("echo", "test")
	if runtime.GOOS == "windows" {
		if cmd.SysProcAttr == nil {
			t.Fatal("expected SysProcAttr to be set on Windows")
		}
		if !cmd.SysProcAttr.HideWindow {
			t.Errorf("expected HideWindow to be true")
		}
		if cmd.SysProcAttr.CreationFlags&CreateNoWindow == 0 {
			t.Errorf("expected CreationFlags to contain CREATE_NO_WINDOW")
		}
	}
}

func TestHideWindowExisting(t *testing.T) {
	cmd := exec.Command("echo", "test")
	HideWindow(cmd)
	if runtime.GOOS == "windows" {
		if !cmd.SysProcAttr.HideWindow {
			t.Errorf("expected HideWindow to be true")
		}
	}
}
