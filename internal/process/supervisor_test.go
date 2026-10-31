package process_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/laragal/laragal/internal/process"
)

func TestSupervisorLifecycle(t *testing.T) {
	sup := process.NewSupervisor()
	sup.LogsDir = t.TempDir()

	// Test starting a mock service (powershell sleep on Windows)
	err := sup.StartService("test_service", "powershell", []string{"-NoProfile", "-Command", "Start-Sleep -Seconds 3"}, 0)
	if err != nil {
		t.Fatalf("failed to start service: %v", err)
	}

	status, found := sup.GetServiceStatus("test_service")
	if !found || !status.Running {
		t.Fatalf("service should be running, got %+v", status)
	}

	err = sup.StopService("test_service")
	if err != nil {
		t.Fatalf("failed to stop service: %v", err)
	}

	time.Sleep(100 * time.Millisecond)
	status, _ = sup.GetServiceStatus("test_service")
	if status.Running {
		t.Fatalf("service should not be running after stop")
	}
}

func TestSupervisorOnExitAndLogging(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("skipping Windows-specific test")
	}

	tempDir := t.TempDir()
	sup := process.NewSupervisor()
	sup.LogsDir = tempDir

	exitCalled := make(chan string, 1)
	sup.OnExit = func(name string) {
		select {
		case exitCalled <- name:
		default:
		}
	}

	// Start a short-lived process (cmd /c exit 3)
	err := sup.StartService("short_service", "cmd.exe", []string{"/c", "echo hello-from-log & exit 3"}, 0)
	if err != nil {
		t.Fatalf("failed to start service: %v", err)
	}

	// Assert OnExit is called with the name within ~2s
	select {
	case name := <-exitCalled:
		if name != "short_service" {
			t.Errorf("expected OnExit for short_service, got %s", name)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for OnExit callback")
	}

	// Assert GetServiceStatus reports Running=false
	status, found := sup.GetServiceStatus("short_service")
	if !found {
		t.Fatal("expected short_service to be found in status")
	}
	if status.Running {
		t.Errorf("expected Running=false, got Running=true")
	}

	// Assert the log file is created under temp dir
	logPath := filepath.Join(tempDir, "short_service.log")
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("expected log file at %s, error: %v", logPath, err)
	}
	if !strings.Contains(string(content), "hello-from-log") {
		t.Errorf("expected log content to contain 'hello-from-log', got: %s", string(content))
	}
}

func TestSupervisorWorkingDirInheritance(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("skipping Windows-specific test")
	}

	tempDir := t.TempDir()
	sup := process.NewSupervisor()
	sup.LogsDir = tempDir

	exitCalled := make(chan string, 1)
	sup.OnExit = func(name string) {
		select {
		case exitCalled <- name:
		default:
		}
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}

	err = sup.StartService("cwdcheck", "cmd", []string{"/c", "cd"}, 0)
	if err != nil {
		t.Fatalf("failed to start service: %v", err)
	}

	select {
	case name := <-exitCalled:
		if name != "cwdcheck" {
			t.Errorf("expected OnExit for cwdcheck, got %s", name)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for OnExit callback")
	}

	logPath := filepath.Join(tempDir, "cwdcheck.log")
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("expected log file at %s, error: %v", logPath, err)
	}

	actualCwd := strings.TrimSpace(string(content))
	if !strings.EqualFold(actualCwd, cwd) {
		t.Fatalf("expected working directory %q, got %q", cwd, actualCwd)
	}
}

