package process_test

import (
	"testing"
	"time"

	"github.com/laragal/laragal/internal/process"
)

func TestSupervisorLifecycle(t *testing.T) {
	sup := process.NewSupervisor()

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
