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
	if !resJaras.IsWeb || resJaras.Domain != "jarasdesk.test" || resJaras.Type != TypeGo {
		t.Fatalf("Expected jarasdesk with web/templates to be detected as TypeGo, got: %+v", resJaras)
	}

	// 10. Actual project on disk test
	if _, err := os.Stat(`D:\projects\jarasdesk`); err == nil {
		resActual := DetectProject(`D:\projects\jarasdesk`, "8.3", ".test")
		if !resActual.IsWeb || resActual.Domain != "jarasdesk.test" {
			t.Fatalf("Actual D:/projects/jarasdesk not detected as web: %+v", resActual)
		}
	}

	// 11. Rust project (Cargo.toml + src/main.rs)
	rustDir := filepath.Join(tempDir, "rust-service")
	_ = os.MkdirAll(filepath.Join(rustDir, "src"), 0755)
	_ = os.WriteFile(filepath.Join(rustDir, "Cargo.toml"), []byte("[package]\nname = \"rust-service\"\nversion = \"0.1.0\"\n"), 0644)
	_ = os.WriteFile(filepath.Join(rustDir, "src", "main.rs"), []byte("fn main() {}"), 0644)

	resRust := DetectProject(rustDir, "8.3", ".test")
	if resRust.Type != TypeRust || !resRust.IsWeb || resRust.ProxyPort != 8080 {
		t.Fatalf("Expected Rust project with TypeRust and port 8080, got: %+v", resRust)
	}

	// 12. .NET / C# project (*.csproj + wwwroot)
	dotnetDir := filepath.Join(tempDir, "dotnet-api")
	_ = os.MkdirAll(filepath.Join(dotnetDir, "wwwroot"), 0755)
	_ = os.WriteFile(filepath.Join(dotnetDir, "app.csproj"), []byte("<Project Sdk=\"Microsoft.NET.Sdk.Web\"></Project>"), 0644)

	resDotNet := DetectProject(dotnetDir, "8.3", ".test")
	if resDotNet.Type != TypeDotNet || !resDotNet.IsWeb || resDotNet.ProxyPort != 5000 {
		t.Fatalf("Expected .NET project with TypeDotNet and port 5000, got: %+v", resDotNet)
	}
	if filepath.Base(resDotNet.DocRoot) != "wwwroot" {
		t.Fatalf("Expected .NET docroot to be wwwroot, got: %s", resDotNet.DocRoot)
	}

	// 13. Python Web project (manage.py or pyproject.toml with FastAPI/Django)
	pyDir := filepath.Join(tempDir, "fastapi-app")
	_ = os.MkdirAll(filepath.Join(pyDir, "static"), 0755)
	_ = os.WriteFile(filepath.Join(pyDir, "requirements.txt"), []byte("fastapi\nuvicorn\n"), 0644)
	_ = os.WriteFile(filepath.Join(pyDir, "main.py"), []byte("from fastapi import FastAPI\napp = FastAPI()"), 0644)

	resPy := DetectProject(pyDir, "8.3", ".test")
	if resPy.Type != TypePython || !resPy.IsWeb || resPy.ProxyPort != 8000 {
		t.Fatalf("Expected Python project with TypePython and port 8000, got: %+v", resPy)
	}

	// 14. Ruby on Rails project (Gemfile + config.ru)
	rubyDir := filepath.Join(tempDir, "rails-app")
	_ = os.MkdirAll(filepath.Join(rubyDir, "public"), 0755)
	_ = os.WriteFile(filepath.Join(rubyDir, "Gemfile"), []byte("source 'https://rubygems.org'\ngem 'rails'"), 0644)

	resRuby := DetectProject(rubyDir, "8.3", ".test")
	if resRuby.Type != TypeRuby || !resRuby.IsWeb || resRuby.ProxyPort != 3000 {
		t.Fatalf("Expected Ruby project with TypeRuby and port 3000, got: %+v", resRuby)
	}

	// 15. Java Spring Boot project (pom.xml)
	javaDir := filepath.Join(tempDir, "spring-app")
	_ = os.MkdirAll(filepath.Join(javaDir, "src", "main", "resources", "static"), 0755)
	_ = os.WriteFile(filepath.Join(javaDir, "pom.xml"), []byte("<project><modelVersion>4.0.0</modelVersion><artifactId>spring-app</artifactId></project>"), 0644)

	resJava := DetectProject(javaDir, "8.3", ".test")
	if resJava.Type != TypeJava || !resJava.IsWeb || resJava.ProxyPort != 8080 {
		t.Fatalf("Expected Java project with TypeJava and port 8080, got: %+v", resJava)
	}

	// 16. Flutter Web project (pubspec.yaml + build/web)
	flutterDir := filepath.Join(tempDir, "flutter-web-app")
	_ = os.MkdirAll(filepath.Join(flutterDir, "build", "web"), 0755)
	_ = os.WriteFile(filepath.Join(flutterDir, "pubspec.yaml"), []byte("name: flutter_app\nflutter:\n  uses-material-design: true\n"), 0644)
	_ = os.WriteFile(filepath.Join(flutterDir, "build", "web", "index.html"), []byte("<!DOCTYPE html><html><body>Flutter</body></html>"), 0644)

	resFlutter := DetectProject(flutterDir, "8.3", ".test")
	if resFlutter.Type != TypeFlutter || !resFlutter.IsWeb || filepath.Base(resFlutter.DocRoot) != "web" {
		t.Fatalf("Expected Flutter project with TypeFlutter and docroot build/web, got: %+v", resFlutter)
	}

	// 17. PHP project with composer.json requiring PHP 8.4
	php84Dir := filepath.Join(tempDir, "laravel-84")
	_ = os.MkdirAll(filepath.Join(php84Dir, "public"), 0755)
	_ = os.WriteFile(filepath.Join(php84Dir, "artisan"), []byte("<?php"), 0644)
	_ = os.WriteFile(filepath.Join(php84Dir, "composer.json"), []byte(`{"require":{"php":"^8.4","laravel/framework":"^11.0"}}`), 0644)

	res84 := DetectProject(php84Dir, "8.3", ".test")
	if res84.Type != TypeLaravel || res84.PHPVersion != "8.4" || res84.PHPPort != 9084 {
		t.Fatalf("Expected Laravel with PHP 8.4 and port 9084, got: %+v", res84)
	}

	// 18. PHP project with .php-version file
	phpVerDir := filepath.Join(tempDir, "custom-php-ver")
	_ = os.MkdirAll(filepath.Join(phpVerDir, "public"), 0755)
	_ = os.WriteFile(filepath.Join(phpVerDir, "public", "index.php"), []byte("<?php"), 0644)
	_ = os.WriteFile(filepath.Join(phpVerDir, ".php-version"), []byte("8.2\n"), 0644)

	resVer := DetectProject(phpVerDir, "8.3", ".test")
	if resVer.PHPVersion != "8.2" || resVer.PHPPort != 9082 {
		t.Fatalf("Expected PHP 8.2 and port 9082 from .php-version, got: %+v", resVer)
	}

	// 19. PHP project with .laragal file
	laragalDir := filepath.Join(tempDir, "laragal-config-php")
	_ = os.MkdirAll(filepath.Join(laragalDir, "public"), 0755)
	_ = os.WriteFile(filepath.Join(laragalDir, "public", "index.php"), []byte("<?php"), 0644)
	_ = os.WriteFile(filepath.Join(laragalDir, ".laragal"), []byte(`{"php":"8.1"}`), 0644)

	resLaragal := DetectProject(laragalDir, "8.3", ".test")
	if resLaragal.PHPVersion != "8.1" || resLaragal.PHPPort != 9081 {
		t.Fatalf("Expected PHP 8.1 and port 9081 from .laragal, got: %+v", resLaragal)
	}
}

func TestParsePHPConstraint(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"^8.4", "8.4"},
		{"^8.4.0", "8.4"},
		{">=8.4", "8.4"},
		{"^8.3", "8.3"},
		{"8.3.*", "8.3"},
		{"^8.2", "8.2"},
		{"~8.2.0", "8.2"},
		{"^8.1", "8.1"},
		{"8.1.*", "8.1"},
		{"^8.0", "8.1"},
		{"^7.4", "8.1"},
		{"*", ""},
	}

	for _, c := range cases {
		got := ParsePHPConstraint(c.input)
		if got != c.expected {
			t.Errorf("ParsePHPConstraint(%q) = %q; expected %q", c.input, got, c.expected)
		}
	}
}
