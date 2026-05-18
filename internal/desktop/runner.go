package desktop

import (
	"embed"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

// Run 创建并运行 Wails 桌面应用
func Run(assets embed.FS) error {
	service := NewApp()
	prefs := service.config.Preferences()
	windowWidth := positivePanelWidth(prefs.PanelWidth)

	desktop := application.New(application.Options{
		Name:        "Siwap",
		Description: "Siwap desktop workspace switcher",
		Services: []application.Service{
			application.NewService(service),
		},
		Assets: application.AssetOptions{
			Handler: application.BundledAssetFileServer(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: false,
		},
		SingleInstance: &application.SingleInstanceOptions{
			UniqueID:               "siwap-main",
			OnSecondInstanceLaunch: service.handleSecondInstanceLaunch,
		},
		ShouldQuit: func() bool {
			_ = service.config.Flush()
			return true
		},
	})

	desktop.Menu.Set(appMenu(service))

	windowOpts := application.WebviewWindowOptions{
		Name:                       "main",
		Title:                      "Siwap",
		Width:                      windowWidth,
		Height:                     prefs.WindowHeight,
		MinWidth:                   260,
		MinHeight:                  620,
		AlwaysOnTop:                prefs.AlwaysOnTop,
		UseApplicationMenu:         true,
		BackgroundColour:           windowBackgroundColour(prefs.Appearance, desktop.Env.IsDarkMode()),
		Mac:                        macWindowChrome(prefs.Appearance),
		Windows:                    windowsWindowChrome(prefs.Appearance),
		DefaultContextMenuDisabled: true,
		URL:                        "/",
	}
	windowOpts.InitialPosition = application.WindowXY
	if prefs.WindowPositionSaved {
		windowOpts.X = prefs.WindowX
		windowOpts.Y = prefs.WindowY
	} else {
		windowOpts.X = 0
		windowOpts.Y = 0
	}
	mainWindow := desktop.Window.NewWithOptions(windowOpts)
	service.attachDesktop(desktop, mainWindow)

	mainWindow.RegisterHook(events.Common.WindowRuntimeReady, func(*application.WindowEvent) {
		service.applyTheme(service.config.Preferences().Appearance)
		service.registerShortcut(service.config.Preferences().GlobalShortcut)
		service.startEdgeRevealWatcher()
	})
	desktop.Event.OnApplicationEvent(events.Common.ThemeChanged, func(*application.ApplicationEvent) {
		service.applyTheme(service.config.Preferences().Appearance)
	})
	mainWindow.RegisterHook(events.Common.WindowClosing, func(event *application.WindowEvent) {
		event.Cancel()
		_ = service.HideWindow()
	})
	mainWindow.RegisterHook(events.Common.WindowDidResize, func(*application.WindowEvent) {
		service.rememberMainWindowSize()
	})
	mainWindow.RegisterHook(events.Common.WindowDidMove, func(*application.WindowEvent) {
		service.rememberMainWindowSize()
	})

	return desktop.Run()
}

// appMenu 构造桌面应用菜单
func appMenu(app *App) *application.Menu {
	root := application.NewMenu()
	appMenu := root.AddSubmenu("Siwap")
	appMenu.Add(app.menuText("显示主窗口", "Show Main Window")).SetAccelerator("Ctrl+Cmd+s").OnClick(func(*application.Context) { _ = app.ShowWindow() })
	appMenu.Add(app.menuText("隐藏主窗口", "Hide Main Window")).SetAccelerator("CmdOrCtrl+h").OnClick(func(*application.Context) { _ = app.HideWindow() })
	appMenu.Add(app.menuText("设置...", "Settings...")).SetAccelerator("CmdOrCtrl+,").OnClick(func(*application.Context) { _ = app.RunAction("settings") })
	appMenu.AddSeparator()
	appMenu.Add(app.menuText("退出", "Quit")).SetAccelerator("CmdOrCtrl+q").OnClick(func(*application.Context) { _ = app.Quit() })

	sessions := root.AddSubmenu(app.menuText("会话", "Sessions"))
	sessions.Add(app.menuText("清理会话", "Clear Sessions")).SetAccelerator("CmdOrCtrl+Shift+delete").OnClick(func(*application.Context) { _ = app.ClearSessions() })

	windowMenu := root.AddSubmenu(app.menuText("窗口", "Window"))
	windowMenu.Add(app.menuText("置顶开关", "Toggle Always On Top")).SetAccelerator("CmdOrCtrl+Shift+t").OnClick(func(*application.Context) { _, _ = app.ToggleAlwaysOnTop() })
	windowMenu.Add(app.menuText("重置位置", "Reset Position")).OnClick(func(*application.Context) { _ = app.ResetSidebarWindow() })
	return root
}
