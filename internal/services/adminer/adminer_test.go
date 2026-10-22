package adminer

import (
	"os"
	"strings"
	"testing"
)

func TestEnsureInstalled(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "laragal-adminer-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	installedPath, err := EnsureInstalled(tempDir)
	if err != nil {
		t.Fatalf("EnsureInstalled failed: %v", err)
	}

	if _, err := os.Stat(installedPath); err != nil {
		t.Errorf("installed file does not exist: %v", err)
	}

	data, err := os.ReadFile(installedPath)
	if err != nil || len(data) == 0 {
		t.Errorf("installed file is empty or cannot be read: %v", err)
	}

	// Idempotency check: calling again should not fail
	installedPath2, err := EnsureInstalled(tempDir)
	if err != nil || installedPath2 != installedPath {
		t.Errorf("second call failed or returned different path: %v", err)
	}
}

func TestBuildURL(t *testing.T) {
	u1 := BuildURL("adminer.test", 3307, "")
	if !strings.Contains(u1, "localhost:3307") || !strings.Contains(u1, "username=root") {
		t.Errorf("expected localhost:3307 and username=root, got: %s", u1)
	}

	u2 := BuildURL("adminer.test", 3307, "my_laravel_app")
	if !strings.Contains(u2, "db=my_laravel_app") {
		t.Errorf("expected db=my_laravel_app in url, got: %s", u2)
	}
}
