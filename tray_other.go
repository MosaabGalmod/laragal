//go:build !windows

package main

func (a *App) initTray() {
	// No-op for non-Windows platforms
}
