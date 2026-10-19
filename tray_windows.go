//go:build windows

package main

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	goruntime "runtime"
	"syscall"
	"unsafe"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed build/appicon.ico
var trayIconBytes []byte

type guid struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}

// notifyIconDataW is the exact 976-byte NOTIFYICONDATAW layout for Windows 64-bit
type notifyIconDataW struct {
	CbSize           uint32
	_                uint32 // 4 bytes padding to align HWnd to 8 bytes
	HWnd             uintptr
	UID              uint32
	UFlags           uint32
	UCallbackMessage uint32
	_                uint32 // 4 bytes padding to align HIcon to 8 bytes
	HIcon            uintptr
	SzTip            [128]uint16
	DwState          uint32
	DwStateMask      uint32
	SzInfo           [256]uint16
	UTimeoutVersion  uint32 // union
	SzInfoTitle      [64]uint16
	DwInfoFlags      uint32
	GuidItem         guid
	HBalloonIcon     uintptr
}

type wndClassExW struct {
	Size, Style                        uint32
	WndProc                            uintptr
	ClsExtra, WndExtra                 int32
	Instance, Icon, Cursor, Background uintptr
	MenuName, ClassName                *uint16
	IconSm                             uintptr
}

type point struct {
	X, Y int32
}

type winMsg struct {
	HWnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      point
}

var (
	u32 = syscall.NewLazyDLL("user32.dll")
	s32 = syscall.NewLazyDLL("shell32.dll")
	k32 = syscall.NewLazyDLL("kernel32.dll")

	pGetModuleHandle       = k32.NewProc("GetModuleHandleW")
	pLoadIcon              = u32.NewProc("LoadIconW")
	pLoadImage             = u32.NewProc("LoadImageW")
	pLoadCursor            = u32.NewProc("LoadCursorW")
	pRegisterClassEx       = u32.NewProc("RegisterClassExW")
	pCreateWindowEx        = u32.NewProc("CreateWindowExW")
	pDefWindowProc         = u32.NewProc("DefWindowProcW")
	pDestroyWindow         = u32.NewProc("DestroyWindow")
	pPostQuitMessage       = u32.NewProc("PostQuitMessage")
	pGetMessage            = u32.NewProc("GetMessageW")
	pTranslateMessage      = u32.NewProc("TranslateMessage")
	pDispatchMessage       = u32.NewProc("DispatchMessageW")
	pCreatePopupMenu       = u32.NewProc("CreatePopupMenu")
	pAppendMenu            = u32.NewProc("AppendMenuW")
	pTrackPopupMenuEx      = u32.NewProc("TrackPopupMenuEx")
	pDestroyMenu           = u32.NewProc("DestroyMenu")
	pSetForegroundWindow   = u32.NewProc("SetForegroundWindow")
	pGetCursorPos          = u32.NewProc("GetCursorPos")
	pRegisterWindowMessage = u32.NewProc("RegisterWindowMessageW")
	pShellNotifyIcon       = s32.NewProc("Shell_NotifyIconW")
)

const (
	wmAppTray = 0x8001
	nimAdd    = 0x00000000
	nimModify = 0x00000001
	nimDelete = 0x00000002

	nifMessage = 0x00000001
	nifIcon    = 0x00000002
	nifTip     = 0x00000004

	wmLbuttonUp     = 0x0202
	wmLbuttonDblClk = 0x0203
	wmRbuttonUp     = 0x0205
	wmDestroy       = 0x0002

	mfString    = 0x00000000
	mfSeparator = 0x00000800

	tpmRightAlign = 0x0008
	tpmLayoutRtl  = 0x8000
	tpmReturnCmd  = 0x0100
)

var (
	globalTrayApp *App
	globalNid     notifyIconDataW
	wmTaskbarCreated uint32
)

func trayWndProc(hWnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case wmAppTray:
		switch lParam {
		case wmLbuttonUp, wmLbuttonDblClk:
			if globalTrayApp != nil && globalTrayApp.ctx != nil {
				runtime.WindowShow(globalTrayApp.ctx)
				runtime.WindowUnminimise(globalTrayApp.ctx)
			}
			return 0

		case wmRbuttonUp:
			hMenu, _, _ := pCreatePopupMenu.Call()
			if hMenu != 0 {
				showText, _ := syscall.UTF16PtrFromString("إظهار LaraGal")
				folderText, _ := syscall.UTF16PtrFromString("فتح مجلد المشاريع")
				quitText, _ := syscall.UTF16PtrFromString("إيقاف الخدمات والخروج التام")

				pAppendMenu.Call(hMenu, mfString, 1001, uintptr(unsafe.Pointer(showText)))
				pAppendMenu.Call(hMenu, mfString, 1002, uintptr(unsafe.Pointer(folderText)))
				pAppendMenu.Call(hMenu, mfSeparator, 0, 0)
				pAppendMenu.Call(hMenu, mfString, 1003, uintptr(unsafe.Pointer(quitText)))

				var pt point
				pGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
				pSetForegroundWindow.Call(hWnd)

				cmd, _, _ := pTrackPopupMenuEx.Call(
					hMenu,
					tpmReturnCmd|tpmRightAlign|tpmLayoutRtl,
					uintptr(pt.X),
					uintptr(pt.Y),
					hWnd,
					0,
				)
				pDestroyMenu.Call(hMenu)

				switch cmd {
				case 1001:
					if globalTrayApp != nil && globalTrayApp.ctx != nil {
						runtime.WindowShow(globalTrayApp.ctx)
						runtime.WindowUnminimise(globalTrayApp.ctx)
					}
				case 1002:
					if globalTrayApp != nil {
						_ = globalTrayApp.OpenFolder(globalTrayApp.cfg.ProjectsPath)
					}
				case 1003:
					if globalTrayApp != nil {
						globalTrayApp.QuitApp()
					}
				}
			}
			return 0
		}

	case wmDestroy:
		pShellNotifyIcon.Call(nimDelete, uintptr(unsafe.Pointer(&globalNid)))
		pPostQuitMessage.Call(0)
		return 0

	default:
		if wmTaskbarCreated != 0 && msg == wmTaskbarCreated {
			// Explorer was restarted, re-register tray icon
			pShellNotifyIcon.Call(nimAdd, uintptr(unsafe.Pointer(&globalNid)))
			return 0
		}
	}

	ret, _, _ := pDefWindowProc.Call(hWnd, uintptr(msg), wParam, lParam)
	return ret
}

func (a *App) initTray() {
	globalTrayApp = a

	go func() {
		goruntime.LockOSThread()
		defer goruntime.UnlockOSThread()

		hInst, _, _ := pGetModuleHandle.Call(0)

		// Register TaskbarCreated message
		tbCreatedStr, _ := syscall.UTF16PtrFromString("TaskbarCreated")
		tbRes, _, _ := pRegisterWindowMessage.Call(uintptr(unsafe.Pointer(tbCreatedStr)))
		wmTaskbarCreated = uint32(tbRes)

		// Load Icon from embedded bytes
		var hIcon uintptr
		if len(trayIconBytes) > 0 {
			tempIconPath := filepath.Join(os.TempDir(), "laragal_tray.ico")
			if err := os.WriteFile(tempIconPath, trayIconBytes, 0644); err == nil {
				pathPtr, _ := syscall.UTF16PtrFromString(tempIconPath)
				// IMAGE_ICON = 1, LR_LOADFROMFILE = 0x0010, LR_DEFAULTSIZE = 0x0040
				hLoaded, _, _ := pLoadImage.Call(0, uintptr(unsafe.Pointer(pathPtr)), 1, 0, 0, 0x0010|0x0040)
				if hLoaded != 0 {
					hIcon = hLoaded
				}
			}
		}
		if hIcon == 0 {
			// Fallback to standard application icon
			hIcon, _, _ = pLoadIcon.Call(0, 32512)
		}

		hCursor, _, _ := pLoadCursor.Call(0, 32512)
		classNamePtr, _ := syscall.UTF16PtrFromString(fmt.Sprintf("LaraGalTrayWnd_%d", os.Getpid()))

		wcex := wndClassExW{
			Size:       uint32(unsafe.Sizeof(wndClassExW{})),
			Style:      3,
			WndProc:    syscall.NewCallback(trayWndProc),
			Instance:   hInst,
			Icon:       hIcon,
			Cursor:     hCursor,
			Background: 6,
			ClassName:  classNamePtr,
			IconSm:     hIcon,
		}

		pRegisterClassEx.Call(uintptr(unsafe.Pointer(&wcex)))

		hWnd, _, _ := pCreateWindowEx.Call(
			0,
			uintptr(unsafe.Pointer(classNamePtr)),
			uintptr(unsafe.Pointer(classNamePtr)),
			0x00CF0000, // WS_OVERLAPPEDWINDOW
			0x80000000, 0x80000000, 0x80000000, 0x80000000,
			0, 0, hInst, 0,
		)

		if hWnd == 0 {
			return
		}

		globalNid = notifyIconDataW{
			CbSize:           uint32(unsafe.Sizeof(globalNid)),
			HWnd:             hWnd,
			UID:              1001,
			UFlags:           nifMessage | nifIcon | nifTip,
			UCallbackMessage: wmAppTray,
			HIcon:            hIcon,
		}
		copy(globalNid.SzTip[:], syscall.StringToUTF16("LaraGal — بيئة التطوير المحلية (تعمل في الخلفية)"))

		// Add tray icon to Windows notification area
		pShellNotifyIcon.Call(nimAdd, uintptr(unsafe.Pointer(&globalNid)))

		// Standard Win32 Message Loop
		var msg winMsg
		for {
			res, _, _ := pGetMessage.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
			if res == 0 || int32(res) == -1 {
				break
			}
			pTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
			pDispatchMessage.Call(uintptr(unsafe.Pointer(&msg)))
		}
	}()
}
