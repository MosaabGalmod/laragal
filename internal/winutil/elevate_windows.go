//go:build windows

package winutil

import (
	"fmt"
	"os"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	shell32           = windows.NewLazySystemDLL("shell32.dll")
	procShellExecuteW = shell32.NewProc("ShellExecuteW")
)

// IsAdmin checks whether the current process is running with elevated Administrator privileges.
func IsAdmin() bool {
	var sid *windows.SID
	err := windows.AllocateAndInitializeSid(
		&windows.SECURITY_NT_AUTHORITY,
		2,
		windows.SECURITY_BUILTIN_DOMAIN_RID,
		windows.DOMAIN_ALIAS_RID_ADMINS,
		0, 0, 0, 0, 0, 0,
		&sid,
	)
	if err != nil {
		return false
	}
	defer func() {
		_ = windows.FreeSid(sid)
	}()

	token := windows.Token(0)
	member, err := token.IsMember(sid)
	if err != nil {
		return false
	}
	return member
}

// RelaunchAsAdmin relaunches the current executable with elevated Administrator privileges via Windows UAC.
func RelaunchAsAdmin() error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	verbPtr, _ := windows.UTF16PtrFromString("runas")
	exePtr, _ := windows.UTF16PtrFromString(exe)

	var args string
	if len(os.Args) > 1 {
		args = strings.Join(os.Args[1:], " ")
	}
	argsPtr, _ := windows.UTF16PtrFromString(args)

	cwd, _ := os.Getwd()
	cwdPtr, _ := windows.UTF16PtrFromString(cwd)

	// SW_SHOWNORMAL = 1
	ret, _, _ := procShellExecuteW.Call(
		0,
		uintptr(unsafe.Pointer(verbPtr)),
		uintptr(unsafe.Pointer(exePtr)),
		uintptr(unsafe.Pointer(argsPtr)),
		uintptr(unsafe.Pointer(cwdPtr)),
		1,
	)

	// ShellExecute returns HINSTANCE > 32 on success
	if ret <= 32 {
		return fmt.Errorf("ShellExecute failed with code: %d", ret)
	}

	return nil
}
