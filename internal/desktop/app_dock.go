package desktop

import (
	"runtime"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/icons"
)

// dockActivationPolicy 返回启动主窗口时需要的普通应用策略
func dockActivationPolicy() application.ActivationPolicy {
	return application.ActivationPolicyRegular
}

// applyDockPreference 根据主窗口状态和偏好刷新 Dock 与状态栏图标
func (a *App) applyDockPreference(showDockIcon bool) {
	a.applyDockPreferenceForWindowState(showDockIcon, a.mainWindowIsVisible())
}

// applyDockPreferenceForWindowState 使用明确的主窗口可见状态刷新 Dock 与状态栏图标
func (a *App) applyDockPreferenceForWindowState(showDockIcon bool, mainWindowVisible bool) {
	if runtime.GOOS != "darwin" {
		return
	}
	settingsWasVisible := a.settingsWindowIsVisible()
	a.ensureStatusItem()
	if showDockIcon || mainWindowVisible || settingsWasVisible {
		a.deferShowDockIcon()
		a.restoreSettingsWindowAfterDockChange(settingsWasVisible)
		return
	}
	a.deferHideDockIcon()
	a.restoreSettingsWindowAfterDockChange(settingsWasVisible)
}

// showDockForForegroundWindow 显示前台窗口所需的 Dock、菜单栏和状态栏入口
func (a *App) showDockForForegroundWindow() {
	if runtime.GOOS != "darwin" {
		return
	}
	a.ensureStatusItem()
	a.deferShowDockIcon()
}

// mainWindowIsVisible 判断主窗口是否可见
func (a *App) mainWindowIsVisible() bool {
	return a.mainWindow != nil && a.mainWindow.IsVisible()
}

// deferShowDockIcon 异步显示 Dock 图标，避免在托盘回调栈里直接切换激活策略
func (a *App) deferShowDockIcon() {
	if runtime.GOOS != "darwin" || a.dockService == nil {
		return
	}
	go func() {
		time.Sleep(50 * time.Millisecond)
		a.dockService.ShowAppIcon()
	}()
}

// deferHideDockIcon 异步隐藏 Dock 图标，避免在托盘回调栈里直接切换激活策略
func (a *App) deferHideDockIcon() {
	if runtime.GOOS != "darwin" || a.dockService == nil {
		return
	}
	go func() {
		time.Sleep(50 * time.Millisecond)
		a.dockService.HideAppIcon()
	}()
}

// restoreSettingsWindowAfterDockChange 在切换 activation policy 后恢复设置窗口焦点
// macOS 可能会在 Dock/Accessory 切换期间关闭或隐藏辅助窗口，这里做一次延迟兜底
func (a *App) restoreSettingsWindowAfterDockChange(wasVisible bool) {
	if !wasVisible {
		return
	}
	go func() {
		time.Sleep(120 * time.Millisecond)
		a.windowMu.Lock()
		settingsWindow := a.settingsWindow
		target := settingsTarget{Section: a.settingsTab}
		a.windowMu.Unlock()
		if settingsWindow == nil {
			a.openOrFocusSettingsWindow(target)
			return
		}
		settingsWindow.Show()
		settingsWindow.Restore()
		settingsWindow.Focus()
	}()
}

// ensureStatusItem 确保 macOS 顶部状态栏图标存在
func (a *App) ensureStatusItem() {
	if runtime.GOOS != "darwin" || a.desktop == nil {
		return
	}
	if a.statusItem != nil {
		return
	}
	item := a.desktop.SystemTray.New()
	item.SetTemplateIcon(icons.SystrayMacTemplate)
	item.SetTooltip("Siwap")
	item.OnClick(func() {
		_ = a.ShowWindow()
	})
	item.SetMenu(a.statusItemMenu())
	a.statusItem = item
}

// statusItemMenu 构造状态栏图标右键菜单
func (a *App) statusItemMenu() *application.Menu {
	menu := a.desktop.Menu.New()
	menu.Add(a.menuText("显示主窗口", "Show Main Window")).OnClick(func(*application.Context) {
		_ = a.ShowWindow()
	})
	menu.Add(a.menuText("隐藏主窗口", "Hide Main Window")).OnClick(func(*application.Context) {
		_ = a.HideWindow()
	})
	menu.AddSeparator()
	menu.Add(a.menuText("设置...", "Settings...")).OnClick(func(*application.Context) {
		_ = a.RunAction("settings")
	})
	menu.AddSeparator()
	menu.Add(a.menuText("退出", "Quit")).OnClick(func(*application.Context) {
		_ = a.Quit()
	})
	return menu
}
