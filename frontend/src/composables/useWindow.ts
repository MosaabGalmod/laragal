export function useWindow() {
  function closeWindow() {
    const w = window as any
    if (w.go?.main?.App?.CloseWindow) {
      w.go.main.App.CloseWindow()
    } else if (w.runtime?.WindowHide) {
      w.runtime.WindowHide()
    } else if (w.runtime?.Quit) {
      w.runtime.Quit()
    }
  }

  function toggleMaximize() {
    const w = window as any
    if (w.go?.main?.App?.ToggleMaximizeWindow) {
      w.go.main.App.ToggleMaximizeWindow()
    } else if (w.runtime?.WindowToggleMaximise) {
      w.runtime.WindowToggleMaximise()
    }
  }

  function minimizeWindow() {
    const w = window as any
    if (w.go?.main?.App?.MinimizeWindow) {
      w.go.main.App.MinimizeWindow()
    } else if (w.runtime?.WindowMinimise) {
      w.runtime.WindowMinimise()
    }
  }

  return {
    closeWindow,
    toggleMaximize,
    minimizeWindow,
  }
}
