package desktop

import (
	"github.com/wailsapp/wails/v3/pkg/application"

	"siwap/internal/domain"
)

// ShowWindow 显示主窗口并切到前台
func (a *App) ShowWindow() domain.ActionResult {
	a.windowActionMu.Lock()
	defer a.windowActionMu.Unlock()
	return a.showWindowLocked()
}

// showWindowLocked 在持锁状态下显示主窗口
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
	a.deferShowDockIcon()
	return domain.ActionResult{OK: true, Status: "shown", Message: "Window shown."}
}

// HideWindow 隐藏主窗口
func (a *App) HideWindow() domain.ActionResult {
	a.windowActionMu.Lock()
	defer a.windowActionMu.Unlock()
	return a.hideWindowLocked()
}

// toggleWindow 在状态栏图标点击时显示或隐藏主窗口
func (a *App) toggleWindow() domain.ActionResult {
	a.windowActionMu.Lock()
	defer a.windowActionMu.Unlock()
	if a.mainWindow == nil {
		return domain.ActionResult{OK: false, Status: "missing", Message: "Wails window is not ready."}
	}
	if a.mainWindow.IsVisible() {
		return a.hideWindowLocked()
	}
	return a.showWindowLocked()
}

// hideWindowLocked 在持锁状态下隐藏主窗口
func (a *App) hideWindowLocked() domain.ActionResult {
	if a.mainWindow == nil {
		return domain.ActionResult{OK: false, Status: "missing", Message: "Wails window is not ready."}
	}
	a.rememberMainWindowSize()
	a.mainWindow.Hide()
	a.applyDockPreferenceForWindowState(a.config.Preferences().ShowDockIcon, false)
	return domain.ActionResult{OK: true, Status: "hidden", Message: "Window hidden."}
}

// Quit 退出桌面应用
func (a *App) Quit() domain.ActionResult {
	_ = a.config.Flush()
	if a.desktop != nil {
		a.desktop.Quit()
	}
	return domain.ActionResult{OK: true, Status: "quit", Message: "Quit requested."}
}

// ToggleAlwaysOnTop 切换主窗口置顶状态
func (a *App) ToggleAlwaysOnTop() (domain.WindowState, error) {
	prefs := a.config.Preferences()
	prefs.AlwaysOnTop = !prefs.AlwaysOnTop
	updated, err := a.config.UpdatePreferences(prefs)
	if a.mainWindow != nil {
		a.mainWindow.SetAlwaysOnTop(updated.AlwaysOnTop)
	}
	return domain.WindowState{Width: updated.WindowWidth, Height: updated.WindowHeight, AlwaysOnTop: updated.AlwaysOnTop, Mode: "normal"}, err
}

// ResetSidebarWindow 重置侧边栏窗口尺寸和位置
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

// handleSecondInstanceLaunch 处理第二个应用实例的启动请求
func (a *App) handleSecondInstanceLaunch(data application.SecondInstanceData) {
	for _, arg := range data.Args {
		if arg == resetSidebarWindowFlag {
			_ = a.ResetSidebarWindow()
			return
		}
	}
	_ = a.ShowWindow()
}

// alignSidebarWindow 将侧边栏窗口对齐到屏幕左侧
func (a *App) alignSidebarWindow() {
	if a.mainWindow == nil {
		return
	}
	x, y, width, height := a.sidebarWindowGeometry()
	a.mainWindow.SetPosition(x, y)
	a.mainWindow.SetSize(width, height)
}

// rememberMainWindowSize 保存主窗口尺寸和位置
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

// sidebarWindowGeometry 计算侧边栏窗口尺寸和坐标
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

// applyTheme 应用当前主题到桌面窗口和菜单
func (a *App) applyTheme(appearance string) {
	colour := windowBackgroundColour(appearance, a.systemDarkMode())
	if a.mainWindow != nil {
		a.mainWindow.SetBackgroundColour(colour)
	}
	if a.settingsWindow != nil {
		a.settingsWindow.SetBackgroundColour(colour)
	}
}

// menuText 根据语言偏好返回菜单文案
func (a *App) menuText(zh string, en string) string {
	if a.config.Preferences().Language == "en" {
		return en
	}
	return zh
}

// systemDarkMode 判断系统当前是否为深色模式
func (a *App) systemDarkMode() bool {
	if a.desktop == nil || a.desktop.Env == nil {
		return false
	}
	return a.desktop.Env.IsDarkMode()
}

// positivePanelWidth 返回有效的面板宽度
func positivePanelWidth(value int) int {
	if value <= 0 {
		return 320
	}
	return clampInt(value, 260, 520)
}

// clampInt 将整数限制在指定范围内
func clampInt(value int, min int, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
