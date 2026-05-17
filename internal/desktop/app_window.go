package desktop

import (
	"github.com/wailsapp/wails/v3/pkg/application"

	"siwap/internal/domain"
)

func (a *App) ShowWindow() domain.ActionResult {
	a.windowActionMu.Lock()
	defer a.windowActionMu.Unlock()
	return a.showWindowLocked()
}

func (a *App) showWindowLocked() domain.ActionResult {
	if a.mainWindow == nil {
		return domain.ActionResult{OK: false, Status: "missing", Message: "Wails window is not ready."}
	}
	x, y, width, height := a.sidebarWindowGeometry()
	a.mainWindow.SetSize(width, height)
	a.mainWindow.SetPosition(x, y)
	a.mainWindow.Show()
	a.mainWindow.Restore()
	a.mainWindow.Focus()
	return domain.ActionResult{OK: true, Status: "shown", Message: "Window shown."}
}

func (a *App) HideWindow() domain.ActionResult {
	a.windowActionMu.Lock()
	defer a.windowActionMu.Unlock()
	return a.hideWindowLocked()
}

func (a *App) hideWindowLocked() domain.ActionResult {
	if a.mainWindow == nil {
		return domain.ActionResult{OK: false, Status: "missing", Message: "Wails window is not ready."}
	}
	a.rememberMainWindowSize()
	a.mainWindow.Hide()
	return domain.ActionResult{OK: true, Status: "hidden", Message: "Window hidden."}
}

func (a *App) Quit() domain.ActionResult {
	_ = a.config.Flush()
	if a.desktop != nil {
		a.desktop.Quit()
	}
	return domain.ActionResult{OK: true, Status: "quit", Message: "Quit requested."}
}

func (a *App) ToggleAlwaysOnTop() (domain.WindowState, error) {
	prefs := a.config.Preferences()
	prefs.AlwaysOnTop = !prefs.AlwaysOnTop
	updated, err := a.config.UpdatePreferences(prefs)
	if a.mainWindow != nil {
		a.mainWindow.SetAlwaysOnTop(updated.AlwaysOnTop)
	}
	return domain.WindowState{Width: updated.WindowWidth, Height: updated.WindowHeight, AlwaysOnTop: updated.AlwaysOnTop, Mode: "normal"}, err
}

func (a *App) ResetSidebarWindow() domain.WindowState {
	a.windowActionMu.Lock()
	defer a.windowActionMu.Unlock()
	prefs := a.config.Preferences()
	prefs.WindowX = 0
	prefs.WindowY = 0
	prefs.WindowPositionSaved = false
	_, _ = a.config.UpdatePreferences(prefs)
	a.alignSidebarWindow()
	if a.mainWindow != nil {
		a.mainWindow.Show()
	}
	prefs = a.config.Preferences()
	return domain.WindowState{Width: prefs.WindowWidth, Height: prefs.WindowHeight, AlwaysOnTop: prefs.AlwaysOnTop, Mode: "sidebar"}
}

func (a *App) handleSecondInstanceLaunch(data application.SecondInstanceData) {
	for _, arg := range data.Args {
		if arg == resetSidebarWindowFlag {
			_ = a.ResetSidebarWindow()
			return
		}
	}
	_ = a.ShowWindow()
}

func (a *App) alignSidebarWindow() {
	if a.mainWindow == nil {
		return
	}
	x, y, width, height := a.sidebarWindowGeometry()
	a.mainWindow.SetPosition(x, y)
	a.mainWindow.SetSize(width, height)
}

func (a *App) rememberMainWindowSize() {
	if a.mainWindow == nil {
		return
	}
	width, height := a.mainWindow.Size()
	x, y := a.mainWindow.Position()
	if width <= 0 || height <= 0 {
		return
	}
	prefs := a.config.Preferences()
	if prefs.PanelWidth == width && prefs.WindowWidth == width && prefs.WindowHeight == height && prefs.WindowX == x && prefs.WindowY == y {
		return
	}
	prefs.PanelWidth = width
	prefs.WindowWidth = width
	prefs.WindowHeight = height
	prefs.WindowX = x
	prefs.WindowY = y
	prefs.WindowPositionSaved = true
	updated, err := a.config.UpdatePreferences(prefs)
	if err == nil {
		a.emit("preferences:updated", updated)
	}
}

func (a *App) sidebarWindowGeometry() (x int, y int, width int, height int) {
	prefs := a.config.Preferences()
	if prefs.WindowPositionSaved {
		width = prefs.WindowWidth
		height = prefs.WindowHeight
		if width <= 0 {
			width = 320
		}
		if height <= 0 {
			height = 900
		}
		return prefs.WindowX, prefs.WindowY, width, height
	}
	if a.desktop != nil {
		screens := a.desktop.Screen.GetAll()
		if len(screens) > 0 {
			screen := screens[0]
			if primary := a.desktop.Screen.GetPrimary(); primary != nil {
				screen = primary
			}
			for _, item := range screens {
				if item.IsPrimary {
					screen = item
				}
			}
			screenWidth := screen.Size.Width
			screenHeight := screen.Size.Height
			if screenWidth == 0 {
				screenWidth = screen.Bounds.Width
			}
			if screenHeight == 0 {
				screenHeight = screen.Bounds.Height
			}
			x = screen.X
			y = screen.Y
			if prefs.PanelWidth <= 0 && screenWidth > 0 {
				width = clampInt(screenWidth/6, 260, 420)
			}
			if screenHeight > 0 {
				height = screenHeight
			}
		}
	}
	return x, y, width, height
}

func (a *App) applyTheme(appearance string) {
	colour := windowBackgroundColour(appearance, a.systemDarkMode())
	if a.mainWindow != nil {
		a.mainWindow.SetBackgroundColour(colour)
	}
	if a.settingsWindow != nil {
		a.settingsWindow.SetBackgroundColour(colour)
	}
}

func (a *App) menuText(zh string, en string) string {
	if a.config.Preferences().Language == "en" {
		return en
	}
	return zh
}

func (a *App) systemDarkMode() bool {
	if a.desktop == nil || a.desktop.Env == nil {
		return false
	}
	return a.desktop.Env.IsDarkMode()
}

func positivePanelWidth(value int) int {
	if value <= 0 {
		return 320
	}
	return clampInt(value, 260, 520)
}

func clampInt(value int, min int, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
