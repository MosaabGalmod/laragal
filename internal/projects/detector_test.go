package projects

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectProject(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Laravel project
	laravelDir := filepath.Join(tempDir, "my-laravel")
	_ = os.MkdirAll(filepath.Join(laravelDir, "public"), 0755)
	_ = os.WriteFile(filepath.Join(laravelDir, "artisan"), []byte("<?php"), 0644)
	_ = os.WriteFile(filepath.Join(laravelDir, "public", "index.php"), []byte("<?php"), 0644)

	res := DetectProject(laravelDir, "8.3", ".test")
	if res.Type != TypeLaravel || !res.IsWeb || res.Domain != "my-laravel.test" {
		t.Fatalf("Expected Laravel project, got: %+v", res)
	}
	if filepath.Base(res.DocRoot) != "public" {
		t.Fatalf("Expected docroot to be public, got: %s", res.DocRoot)
	}

	// 2. Pure PHP project
	phpDir := filepath.Join(tempDir, "my-php")
	_ = os.MkdirAll(phpDir, 0755)
	_ = os.WriteFile(filepath.Join(phpDir, "index.php"), []byte("<?php echo 'hello';"), 0644)

	resPHP := DetectProject(phpDir, "8.3", ".test")
	if resPHP.Type != TypePHP || !resPHP.IsWeb || resPHP.PHPPort != 9083 {
		t.Fatalf("Expected PHP project, got: %+v", resPHP)
	}

	// 3. Static HTML project
	staticDir := filepath.Join(tempDir, "my-static")
	_ = os.MkdirAll(staticDir, 0755)
	_ = os.WriteFile(filepath.Join(staticDir, "index.html"), []byte("<h1>Hello</h1>"), 0644)

	resStatic := DetectProject(staticDir, "8.3", ".test")
	if resStatic.Type != TypeStatic || !resStatic.IsWeb || resStatic.PHPPort != 0 {
		t.Fatalf("Expected Static project with port 0, got: %+v", resStatic)
	}

	// 4. Non-web directory (e.g. skills, markdown prompts, tools)
	nonWebDir := filepath.Join(tempDir, "anti-ai-slop")
	_ = os.MkdirAll(nonWebDir, 0755)
	_ = os.WriteFile(filepath.Join(nonWebDir, "SKILL.md"), []byte("# Anti AI Slop"), 0644)
	_ = os.WriteFile(filepath.Join(nonWebDir, "notes.txt"), []byte("notes"), 0644)

	resNonWeb := DetectProject(nonWebDir, "8.3", ".test")
	if resNonWeb.Type != TypeNonWeb || resNonWeb.IsWeb || resNonWeb.Domain != "" {
		t.Fatalf("Expected Non-Web project, got: %+v", resNonWeb)
	}
}
