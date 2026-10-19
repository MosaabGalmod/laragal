package main

import (
	_ "embed"

	"github.com/getlantern/systray"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed build/appicon.ico
var trayIconBytes []byte

func (a *App) initTray() {
	go systray.Run(func() {
		if len(trayIconBytes) > 0 {
			systray.SetIcon(trayIconBytes)
		}
		systray.SetTitle("LaraGal")
		systray.SetTooltip("LaraGal — بيئة التطوير المحلية (تعمل في الخلفية)")

		mShow := systray.AddMenuItem("إظهار LaraGal", "فتح واجهة التطبيق الرئيسية")
		mFolder := systray.AddMenuItem("فتح مجلد المشاريع", "استعراض مجلد المشاريع الحالي في ويندوز")
		systray.AddSeparator()
		mQuit := systray.AddMenuItem("إيقاف الخدمات والخروج التام", "إيقاف خوادم التطوير وإغلاق البرنامج نهائياً")

		go func() {
			for {
				select {
				case <-mShow.ClickedCh:
					if a.ctx != nil {
						runtime.WindowShow(a.ctx)
						runtime.WindowUnminimise(a.ctx)
					}
				case <-mFolder.ClickedCh:
					_ = a.OpenFolder(a.cfg.ProjectsPath)
				case <-mQuit.ClickedCh:
					systray.Quit()
					a.QuitApp()
					return
				}
			}
		}()
	}, func() {
		// onExit cleanup
	})
}
