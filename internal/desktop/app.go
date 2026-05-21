package desktop

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/services/dock"
	"golang.design/x/hotkey"

	"siwap/internal/config"
	"siwap/internal/domain"
	"siwap/internal/harness"
	"siwap/internal/project"
	"siwap/internal/session"
	"siwap/internal/terminal"
	"siwap/internal/worktree"
)

// Version 表示当前应用版本，由构建脚本注入
var Version = "dev"

const resetSidebarWindowFlag = "--reset-sidebar-window"

// App 聚合桌面窗口、配置、项目、终端和会话服务
type App struct {
	ctx            context.Context
	cancel         context.CancelFunc
	focusSerial    atomic.Int64
	config         *config.Store
	harnesses      *harness.Service
	projects       *project.Service
	terminals      *terminal.Service
	sessions       *session.Service
	worktrees      *worktree.Service
	settingsTab    string
	windowMu       sync.Mutex
	desktop        *application.App
	dockService    *dock.DockService
	mainWindow     application.Window
	settingsWindow application.Window
	statusItem     *application.SystemTray
	hotkeyMu       sync.Mutex
	hotkey         *hotkey.Hotkey
	hotkeyStop     chan struct{}
	hotkeyShortcut string
	windowActionMu sync.Mutex
	edgeMu         sync.Mutex
	edgeStop       chan struct{}
}

// NewApp 创建桌面应用服务实例
func NewApp() *App {
	store := config.NewStore()
	return &App{
		config:      store,
		harnesses:   harness.NewService(store),
		projects:    project.NewService(store),
		terminals:   terminal.NewService(),
		sessions:    session.NewService(),
		worktrees:   worktree.NewService(),
		settingsTab: "general",
	}
}

// attachDesktop 绑定 Wails 应用和主窗口实例
func (a *App) attachDesktop(desktop *application.App, mainWindow application.Window) {
	a.windowMu.Lock()
	defer a.windowMu.Unlock()
	a.desktop = desktop
	a.mainWindow = mainWindow
}

// attachDockService 绑定 Dock 服务，用于在 macOS 上切换 Dock 图标显示状态
func (a *App) attachDockService(dockService *dock.DockService) {
	a.windowMu.Lock()
	defer a.windowMu.Unlock()
	a.dockService = dockService
}

// ServiceStartup 在 Wails 服务启动时初始化运行上下文
func (a *App) ServiceStartup(ctx context.Context, _ application.ServiceOptions) error {
	a.ctx, a.cancel = context.WithCancel(ctx)
	return nil
}

// ServiceShutdown 在 Wails 服务关闭时释放资源并保存配置
func (a *App) ServiceShutdown() error {
	if a.cancel != nil {
		a.cancel()
	}
	a.stopEdgeRevealWatcher()
	a.unregisterShortcut()
	_ = a.config.Flush()
	return nil
}

// GetBootstrap 返回前端启动所需的完整数据
func (a *App) GetBootstrap() domain.Bootstrap {
	return domain.Bootstrap{
		Version:          Version,
		Platform:         runtime.GOOS,
		Summary:          a.config.Summary,
		ConfigPath:       a.config.ConfigPath(),
		Preferences:      a.config.Preferences(),
		Harnesses:        a.harnesses.List(),
		Projects:         a.projects.List(),
		TerminalProfiles: a.config.ListTerminalProfiles(),
		Adapters:         a.currentAdapters(),
		Sessions:         a.listSessions(),
		Worktrees:        a.listAllWorktrees(),
	}
}

// GetPreferences 返回当前偏好设置
func (a *App) GetPreferences() domain.Preferences { return a.config.Preferences() }

// GetWindowRole 返回当前窗口角色
func (a *App) GetWindowRole() string {
	return "main"
}

// emit 向前端发送事件
func (a *App) emit(name string, data interface{}) {
	if a.desktop == nil {
		return
	}
	a.desktop.Event.Emit(name, data)
}
