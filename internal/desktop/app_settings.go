package desktop

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"

	"siwap/internal/domain"
)

// GetInitialSettingsSection 返回设置窗口初始打开的分区
func (a *App) GetInitialSettingsSection() string {
	return a.settingsTab
}

// OpenSettingsDialog 打开设置窗口
func (a *App) OpenSettingsDialog() (domain.ActionResult, error) {
	return a.OpenSettingsSection("general")
}

// OpenSettingsSection 打开设置窗口并定位到指定分区
func (a *App) OpenSettingsSection(section string) (domain.ActionResult, error) {
	if a.desktop == nil {
		return domain.ActionResult{OK: false, Status: "missing", Message: "Wails application is not ready."}, fmt.Errorf("Wails application is not ready")
	}
	target := parseSettingsTarget(section)
	a.settingsTab = target.Section
	a.openOrFocusSettingsWindow(target)
	return domain.ActionResult{OK: true, Status: "settings", Message: "Settings window opened."}, nil
}

type settingsTarget struct {
	Section string
	Action  string
}

// openOrFocusSettingsWindow 打开或聚焦设置窗口
func (a *App) openOrFocusSettingsWindow(target settingsTarget) {
	a.windowMu.Lock()
	defer a.windowMu.Unlock()
	route := settingsWindowRoute(target)
	if a.settingsWindow != nil {
		a.settingsWindow.SetURL(route)
		a.settingsWindow.Show()
		a.settingsWindow.Restore()
		a.settingsWindow.Focus()
		a.settingsWindow.EmitEvent("ui:open-settings", settingsTargetPayload(target))
		return
	}
	settingsWindow := a.desktop.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:                       "settings",
		Title:                      "Siwap",
		Width:                      860,
		Height:                     640,
		MinWidth:                   720,
		MinHeight:                  520,
		InitialPosition:            application.WindowCentered,
		UseApplicationMenu:         true,
		BackgroundColour:           windowBackgroundColour(a.config.Preferences().Appearance, a.systemDarkMode()),
		Mac:                        macWindowChrome(a.config.Preferences().Appearance),
		Windows:                    windowsWindowChrome(a.config.Preferences().Appearance),
		DefaultContextMenuDisabled: true,
		URL:                        route,
	})
	settingsWindow.RegisterHook(events.Common.WindowClosing, func(event *application.WindowEvent) {
		event.Cancel()
		settingsWindow.Hide()
	})
	a.settingsWindow = settingsWindow
	settingsWindow.Show()
	settingsWindow.Focus()
}

// settingsWindowRoute 生成设置窗口路由
func settingsWindowRoute(target settingsTarget) string {
	next := url.URL{Path: "/"}
	query := next.Query()
	query.Set("window", "settings")
	query.Set("section", target.Section)
	if target.Action != "" {
		query.Set("action", target.Action)
	}
	next.RawQuery = query.Encode()
	return next.String()
}

// parseSettingsTarget 解析设置窗口目标参数
func parseSettingsTarget(value string) settingsTarget {
	section, action, _ := strings.Cut(value, ":")
	target := settingsTarget{Section: sanitizeSettingsSection(section)}
	if target.Section == "worktrees" && action == "create" {
		target.Action = "create"
	}
	return target
}

// settingsTargetPayload 生成设置目标的前端事件载荷
func settingsTargetPayload(target settingsTarget) string {
	if target.Action == "" {
		return target.Section
	}
	return target.Section + ":" + target.Action
}

// CloseSettingsWindow 关闭设置窗口
func (a *App) CloseSettingsWindow() domain.ActionResult {
	if a.settingsWindow != nil {
		a.settingsWindow.Hide()
	}
	return domain.ActionResult{OK: true, Status: "settings", Message: "Settings window closed."}
}

// RunAction 执行前端请求的窗口级动作
func (a *App) RunAction(action string) domain.ActionResult {
	switch action {
	case "settings":
		result, err := a.OpenSettingsDialog()
		if err != nil {
			return domain.ActionResult{OK: false, Status: "failed", Message: err.Error()}
		}
		return result
	case "hide":
		return a.HideWindow()
	case "show":
		return a.ShowWindow()
	case "toggleAlwaysOnTop":
		_, err := a.ToggleAlwaysOnTop()
		if err != nil {
			return domain.ActionResult{OK: false, Status: "failed", Message: err.Error()}
		}
		return domain.ActionResult{OK: true, Status: "always-on-top", Message: "Always-on-top toggled."}
	default:
		return domain.ActionResult{OK: false, Status: "unknown", Message: "Unknown action: " + action}
	}
}

// sanitizeSettingsSection 清理并校验设置分区名称
func sanitizeSettingsSection(section string) string {
	switch section {
	case "general", "projects", "worktrees", "terminal", "ai":
		return section
	default:
		return "general"
	}
}
