package main

import (
	"context"
	"embed"
	"os"
	"path/filepath"

	"github.com/laragal/laragal/internal/winutil"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Ensure process working directory always matches executable directory
	// (critical when launched via Windows Startup, shortcuts, or system tasks)
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		_ = os.Chdir(exeDir)
	}

	// Auto-elevate to Administrator on Windows if launched without elevation.
	// This ensures seamless operation without requiring manual "Run as administrator":
	// 1. Modifying C:\Windows\System32\drivers\etc\hosts for *.test domains
	// 2. High-performance reverse proxy binding on ports 80 and 443
	// 3. Installing and trusting local SSL certificates
	if !winutil.IsAdmin() {
		if err := winutil.RelaunchAsAdmin(); err == nil {
			os.Exit(0)
		}
	}

	app := NewApp()

	err := wails.Run(&options.App{
		Title:             "LaraGal — بيئة الاستضافة والتطوير المحلية الشاملة لتطبيقات الويب",
		Width:             1280,
		Height:            820,
		MinWidth:          960,
		MinHeight:         640,
		Frameless:         true,
		HideWindowOnClose: true,
		StartHidden:       false,
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId: "laragal-instance-lock-5e4a8f29-1b3a-4d72-8c9b",
			OnSecondInstanceLaunch: func(data options.SecondInstanceData) {
				if app.ctx != nil {
					runtime.WindowShow(app.ctx)
					runtime.WindowUnminimise(app.ctx)
				}
			},
		},
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 248, G: 250, B: 252, A: 1},
		OnStartup:        app.startup,
		OnDomReady: func(ctx context.Context) {
			runtime.WindowShow(ctx)
			runtime.WindowUnminimise(ctx)
		},
		OnShutdown: app.shutdown,
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			WebviewIsTransparent:             false,
			WindowIsTranslucent:              false,
			BackdropType:                     windows.Auto,
			DisableFramelessWindowDecorations: false,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
