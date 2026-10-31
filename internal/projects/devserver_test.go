package projects_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/laragal/laragal/internal/projects"
)

func TestDetectDevServer_NoPackageJSON(t *testing.T) {
	tmpDir := t.TempDir()
	info := projects.DetectDevServer(tmpDir)
	if info.HasPackageJSON {
		t.Errorf("expected HasPackageJSON=false for empty dir")
	}
}

func TestDetectDevServer_WithVite(t *testing.T) {
	tmpDir := t.TempDir()
	pkgJSON := `{
		"name": "my-vite-app",
		"scripts": {
			"dev": "vite",
			"build": "vite build"
		},
		"devDependencies": {
			"vite": "^5.0.0"
		}
	}`
	err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(pkgJSON), 0644)
	if err != nil {
		t.Fatalf("failed to write package.json: %v", err)
	}

	info := projects.DetectDevServer(tmpDir)
	if !info.HasPackageJSON {
		t.Errorf("expected HasPackageJSON=true")
	}
	if info.ScriptCommand != "dev" {
		t.Errorf("expected ScriptCommand=dev, got %s", info.ScriptCommand)
	}
	if info.Framework != "Vite" {
		t.Errorf("expected Framework=Vite, got %s", info.Framework)
	}
	if info.DefaultPort != 5173 {
		t.Errorf("expected DefaultPort=5173, got %d", info.DefaultPort)
	}
}

func TestDetectDevServer_PnpmAndNext(t *testing.T) {
	tmpDir := t.TempDir()
	pkgJSON := `{
		"name": "my-next-app",
		"scripts": {
			"dev": "next dev",
			"build": "next build"
		},
		"dependencies": {
			"next": "14.0.0"
		}
	}`
	_ = os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(pkgJSON), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "pnpm-lock.yaml"), []byte(""), 0644)

	info := projects.DetectDevServer(tmpDir)
	if info.PackageManager != "pnpm" {
		t.Errorf("expected PackageManager=pnpm, got %s", info.PackageManager)
	}
	if info.Framework != "Next.js" {
		t.Errorf("expected Framework=Next.js, got %s", info.Framework)
	}
	if info.DefaultPort != 3000 {
		t.Errorf("expected DefaultPort=3000, got %d", info.DefaultPort)
	}
}

func TestDetectNodeEnvironment(t *testing.T) {
	env := projects.DetectNodeEnvironment()
	// Node should be available on the test machine
	if env.Available {
		if env.NodeVersion == "" {
			t.Errorf("expected non-empty NodeVersion when Available=true")
		}
	}
}
