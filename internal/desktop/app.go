package desktop

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/wailsapp/wails/v3/pkg/application"
	"golang.design/x/hotkey"

	"siwap/internal/config"
	"siwap/internal/domain"
	"siwap/internal/harness"
	"siwap/internal/project"
	"siwap/internal/session"
	"siwap/internal/terminal"
	"siwap/internal/worktree"
)

var Version = "dev"

const resetSidebarWindowFlag = "--reset-sidebar-window"

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
	mainWindow     application.Window
	settingsWindow application.Window
	hotkeyMu       sync.Mutex
	hotkey         *hotkey.Hotkey
	hotkeyStop     chan struct{}
	hotkeyShortcut string
	windowActionMu sync.Mutex
	edgeMu         sync.Mutex
	edgeStop       chan struct{}
}

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

func (a *App) attachDesktop(desktop *application.App, mainWindow application.Window) {
	a.windowMu.Lock()
	defer a.windowMu.Unlock()
	a.desktop = desktop
	a.mainWindow = mainWindow
}

func (a *App) ServiceStartup(ctx context.Context, _ application.ServiceOptions) error {
	a.ctx, a.cancel = context.WithCancel(ctx)
	return nil
}

func (a *App) ServiceShutdown() error {
	if a.cancel != nil {
		a.cancel()
	}
	a.stopEdgeRevealWatcher()
	a.unregisterShortcut()
	_ = a.config.Flush()
	return nil
}

func (a *App) GetBootstrap() domain.Bootstrap {
	return domain.Bootstrap{
		Version:          Version,
		Summary:          a.config.Summary,
		ConfigPath:       a.config.ConfigPath(),
		Preferences:      a.config.Preferences(),
		Harnesses:        a.harnesses.List(),
		Projects:         a.projects.List(),
		TerminalProfiles: a.config.ListTerminalProfiles(),
		Adapters:         a.currentAdapters(),
		Sessions:         a.sessions.List(),
	}
}

func (a *App) GetPreferences() domain.Preferences { return a.config.Preferences() }

func (a *App) GetWindowRole() string {
	return "main"
}

func (a *App) emit(name string, data interface{}) {
	if a.desktop == nil {
		return
	}
	a.desktop.Event.Emit(name, data)
}
