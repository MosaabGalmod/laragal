package projects

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectProject(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Laravel project (artisan in root)
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

	// 3. Composer PHP project (e.g. Symfony or custom app)
	composerDir := filepath.Join(tempDir, "symfony-app")
	_ = os.MkdirAll(filepath.Join(composerDir, "public"), 0755)
	_ = os.WriteFile(filepath.Join(composerDir, "composer.json"), []byte(`{"require":{"symfony/framework-bundle":"^6.0"}}`), 0644)

	resComposer := DetectProject(composerDir, "8.3", ".test")
	if resComposer.Type != TypePHP || !resComposer.IsWeb || resComposer.PHPPort != 9083 {
		t.Fatalf("Expected PHP project for composer.json, got: %+v", resComposer)
	}
	if filepath.Base(resComposer.DocRoot) != "public" {
		t.Fatalf("Expected docroot public for composer app, got: %s", resComposer.DocRoot)
	}

	// 4. Static HTML project
	staticDir := filepath.Join(tempDir, "my-static")
	_ = os.MkdirAll(staticDir, 0755)
	_ = os.WriteFile(filepath.Join(staticDir, "index.html"), []byte("<h1>Hello</h1>"), 0644)

	resStatic := DetectProject(staticDir, "8.3", ".test")
	if resStatic.Type != TypeStatic || !resStatic.IsWeb || resStatic.PHPPort != 0 {
		t.Fatalf("Expected Static project with port 0, got: %+v", resStatic)
	}

	// 5. Node.js Vite project
	nodeDir := filepath.Join(tempDir, "vue-app")
	_ = os.MkdirAll(filepath.Join(nodeDir, "dist"), 0755)
	_ = os.WriteFile(filepath.Join(nodeDir, "package.json"), []byte(`{"name":"vue-app","dependencies":{"vue":"^3.0","vite":"^5.0"}}`), 0644)

	resNode := DetectProject(nodeDir, "8.3", ".test")
	if resNode.Type != TypeNode || !resNode.IsWeb || resNode.PHPPort != 0 {
		t.Fatalf("Expected Node.js project with port 0, got: %+v", resNode)
	}

	// 6. Non-web directory (e.g. skills, markdown prompts, tools)
	nonWebDir := filepath.Join(tempDir, "anti-ai-slop")
	_ = os.MkdirAll(nonWebDir, 0755)
	_ = os.WriteFile(filepath.Join(nonWebDir, "SKILL.md"), []byte("# Anti AI Slop"), 0644)
	_ = os.WriteFile(filepath.Join(nonWebDir, "notes.txt"), []byte("notes"), 0644)

	resNonWeb := DetectProject(nonWebDir, "8.3", ".test")
	if resNonWeb.Type != TypeNonWeb || resNonWeb.IsWeb || resNonWeb.Domain != "" {
		t.Fatalf("Expected Non-Web project, got: %+v", resNonWeb)
	}

	// 7. Multi-tier SaaS project with "backend" folder containing Laravel
	saasDir := filepath.Join(tempDir, "biometric-attendance-saas")
	_ = os.MkdirAll(filepath.Join(saasDir, "backend", "public"), 0755)
	_ = os.WriteFile(filepath.Join(saasDir, "backend", "artisan"), []byte("<?php"), 0644)
	_ = os.WriteFile(filepath.Join(saasDir, "backend", "public", "index.php"), []byte("<?php"), 0644)
	_ = os.MkdirAll(filepath.Join(saasDir, "mobile"), 0755)

	resSaas := DetectProject(saasDir, "8.3", ".test")
	if resSaas.Type != TypeLaravel || !resSaas.IsWeb || resSaas.Domain != "biometric-attendance-saas.test" {
		t.Fatalf("Expected multi-tier SaaS to be detected as web Laravel, got: %+v", resSaas)
	}
	expectedDocRoot := filepath.Join(saasDir, "backend", "public")
	if resSaas.DocRoot != expectedDocRoot {
		t.Fatalf("Expected docroot %s, got: %s", expectedDocRoot, resSaas.DocRoot)
	}

	// 8. Monorepo with "frontend" web app
	fullstackDir := filepath.Join(tempDir, "my-fullstack")
	_ = os.MkdirAll(filepath.Join(fullstackDir, "frontend", "dist"), 0755)
	_ = os.WriteFile(filepath.Join(fullstackDir, "frontend", "package.json"), []byte(`{"dependencies":{"vite":"^5.0"}}`), 0644)
	_ = os.WriteFile(filepath.Join(fullstackDir, "frontend", "dist", "index.html"), []byte("<h1>App</h1>"), 0644)

	resFullstack := DetectProject(fullstackDir, "8.3", ".test")
	if !resFullstack.IsWeb || resFullstack.Domain != "my-fullstack.test" || filepath.Base(resFullstack.DocRoot) != "dist" {
		t.Fatalf("Expected frontend subfolder to be detected as web with dist root, got: %+v", resFullstack)
	}

	// 9. Go / Web project like jarasdesk with web/templates and web/static
	jarasDir := filepath.Join(tempDir, "jarasdesk")
	_ = os.MkdirAll(filepath.Join(jarasDir, "web", "templates"), 0755)
	_ = os.MkdirAll(filepath.Join(jarasDir, "web", "static", "css"), 0755)
	_ = os.WriteFile(filepath.Join(jarasDir, "web", "templates", "dashboard.html"), []byte("<h1>Dashboard</h1>"), 0644)
	_ = os.WriteFile(filepath.Join(jarasDir, "go.mod"), []byte("module jarasdesk\n\ngo 1.22"), 0644)

	resJaras := DetectProject(jarasDir, "8.3", ".test")
	if !resJaras.IsWeb || resJaras.Domain != "jarasdesk.test" {
		t.Fatalf("Expected jarasdesk with web/templates to be detected as web, got: %+v", resJaras)
	}

	// 10. Actual project on disk test
	if _, err := os.Stat(`D:\projects\jarasdesk`); err == nil {
		resActual := DetectProject(`D:\projects\jarasdesk`, "8.3", ".test")
		if !resActual.IsWeb || resActual.Domain != "jarasdesk.test" {
			t.Fatalf("Actual D:/projects/jarasdesk not detected as web: %+v", resActual)
		}
	}
}
