package winutil

import (
	"os"
	"strings"
	"testing"
)

func TestAutostartToggle(t *testing.T) {
	appName := "LaraGal_Test_Startup"

	// Ensure cleanup
	defer func() {
		_ = SetStartOnBoot(appName, "", false)
	}()

	// 1. Initial should be false or clean
	_ = SetStartOnBoot(appName, "", false)
	if IsStartOnBoot(appName) {
		t.Errorf("expected %s to not be registered on boot initially", appName)
	}

	// 2. Enable with dummy path
	dummyExe, _ := os.Executable()
	if err := SetStartOnBoot(appName, dummyExe, true); err != nil {
		if strings.Contains(err.Error(), "Access is denied") {
			t.Skip("skipping registry test in restricted sandbox environment")
		}
		t.Fatalf("failed to enable start on boot: %v", err)
	}

	// 3. Verify it is set
	if !IsStartOnBoot(appName) {
		t.Errorf("expected %s to be registered on boot after enabling", appName)
	}

	// 4. Disable
	if err := SetStartOnBoot(appName, "", false); err != nil {
		t.Fatalf("failed to disable start on boot: %v", err)
	}

	// 5. Verify it is disabled
	if IsStartOnBoot(appName) {
		t.Errorf("expected %s to be unregistered after disabling", appName)
	}
}
