package desktop

import (
	"fmt"
	"strings"

	"siwap/internal/domain"
	"siwap/internal/worktree"
)

// UpdatePreferences 保存用户偏好并刷新相关窗口状态
func (a *App) UpdatePreferences(prefs domain.Preferences) (domain.Preferences, error) {
	updated, err := a.config.UpdatePreferences(prefs)
	if a.mainWindow != nil {
		a.mainWindow.SetAlwaysOnTop(updated.AlwaysOnTop)
	}
	a.applyTheme(updated.Appearance)
	// Rebuilding the native app menu from a service callback can terminate the
	// Wails dev primary process on macOS. The UI language updates immediately;
	// menu labels pick up the saved language on the next app start.
	a.registerShortcut(updated.GlobalShortcut)
	a.emit("preferences:updated", updated)
	return updated, err
}

// ListHarnesses 返回所有助手配置
func (a *App) ListHarnesses() []domain.Harness { return a.harnesses.List() }

// UpdateHarness 更新助手配置并广播最新列表
func (a *App) UpdateHarness(next domain.Harness) (domain.Harness, error) {
	updated, err := a.harnesses.Update(next)
	a.emit("harnesses:updated", a.harnesses.List())
	return updated, err
}

// CreateHarness 创建助手配置并广播最新列表
func (a *App) CreateHarness(next domain.Harness) (domain.Harness, error) {
	next.Enabled = true
	created, err := a.harnesses.Create(next)
	a.emit("harnesses:updated", a.harnesses.List())
	return created, err
}

// RemoveHarness 删除助手配置并广播最新列表
func (a *App) RemoveHarness(id string) error {
	err := a.harnesses.Remove(id)
	a.emit("harnesses:updated", a.harnesses.List())
	return err
}

// ReorderHarnesses 重排助手列表并广播最新顺序
func (a *App) ReorderHarnesses(ids []string) ([]domain.Harness, error) {
	items, err := a.harnesses.Reorder(ids)
	a.emit("harnesses:updated", items)
	a.emit("preferences:updated", a.config.Preferences())
	return items, err
}

// ListProjects 返回所有项目配置
func (a *App) ListProjects() []domain.Project { return a.projects.List() }

// ChooseProjectDirectory 打开目录选择器并添加选中的项目
func (a *App) ChooseProjectDirectory() (domain.Project, error) {
	if a.desktop == nil {
		return domain.Project{}, fmt.Errorf("Wails application is not ready")
	}
	path, err := a.desktop.Dialog.OpenFile().
		SetTitle(a.menuText("选择项目文件夹", "Choose Project Folder")).
		CanChooseFiles(false).
		CanChooseDirectories(true).
		CanCreateDirectories(false).
		PromptForSingleSelection()
	if err != nil {
		return domain.Project{}, err
	}
	if strings.TrimSpace(path) == "" {
		return domain.Project{}, fmt.Errorf("no directory selected")
	}
	created, err := a.projects.Add(path, "")
	a.emit("projects:updated", a.projects.List())
	a.emit("preferences:updated", a.config.Preferences())
	return created, err
}

// RemoveProject 删除项目并刷新前端数据
func (a *App) RemoveProject(id string) error {
	err := a.projects.Remove(id)
	a.emit("projects:updated", a.projects.List())
	a.emit("preferences:updated", a.config.Preferences())
	return err
}

// SelectProject 切换当前项目并刷新前端数据
func (a *App) SelectProject(id string) (domain.Project, error) {
	selected, err := a.projects.Select(id)
	a.emit("projects:updated", a.projects.List())
	a.emit("preferences:updated", a.config.Preferences())
	return selected, err
}

// SetDefaultProject 设置默认项目并刷新前端数据
func (a *App) SetDefaultProject(id string) (domain.Project, error) {
	project, err := a.projects.SetDefault(id)
	a.emit("projects:updated", a.projects.List())
	a.emit("preferences:updated", a.config.Preferences())
	return project, err
}

// ReorderProjects 重排项目列表并刷新前端数据
func (a *App) ReorderProjects(ids []string) ([]domain.Project, error) {
	projects, err := a.projects.Reorder(ids)
	a.emit("projects:updated", projects)
	a.emit("preferences:updated", a.config.Preferences())
	return projects, err
}

// ListTerminalAdapters 返回当前可展示的终端适配器
func (a *App) ListTerminalAdapters() []domain.TerminalAdapter { return a.currentAdapters() }

// ListTerminalProfiles 返回自定义终端配置列表
func (a *App) ListTerminalProfiles() []domain.TerminalProfile { return a.config.ListTerminalProfiles() }

// ChooseTerminalExecutable 打开文件选择器并返回终端可执行文件路径
func (a *App) ChooseTerminalExecutable() (string, error) {
	if a.desktop == nil {
		return "", fmt.Errorf("Wails application is not ready")
	}
	return a.desktop.Dialog.OpenFile().
		SetTitle(a.menuText("选择终端应用或可执行文件", "Choose Terminal Application")).
		ShowHiddenFiles(true).
		ResolvesAliases(true).
		CanChooseFiles(true).
		CanChooseDirectories(false).
		CanCreateDirectories(false).
		PromptForSingleSelection()
}

// UpsertTerminalProfile 创建或更新自定义终端配置
func (a *App) UpsertTerminalProfile(profile domain.TerminalProfile) (domain.TerminalProfile, error) {
	updated, err := a.config.UpsertTerminalProfile(profile)
	a.emit("terminalProfiles:updated", a.config.ListTerminalProfiles())
	a.emit("adapters:updated", a.currentAdapters())
	return updated, err
}

// RemoveTerminalProfile 删除自定义终端配置
func (a *App) RemoveTerminalProfile(id string) error {
	err := a.config.RemoveTerminalProfile(id)
	a.emit("terminalProfiles:updated", a.config.ListTerminalProfiles())
	a.emit("adapters:updated", a.currentAdapters())
	return err
}

// ReorderTerminalAdapters 保存终端展示顺序
func (a *App) ReorderTerminalAdapters(ids []string) ([]domain.TerminalAdapter, error) {
	prefs := a.config.Preferences()
	prefs.TerminalOrder = append([]string(nil), ids...)
	_, err := a.config.UpdatePreferences(prefs)
	adapters := a.currentAdapters()
	a.emit("adapters:updated", adapters)
	a.emit("preferences:updated", a.config.Preferences())
	return adapters, err
}

// ListSessions 返回当前会话列表
func (a *App) ListSessions() []domain.Session { return a.sessions.List() }

// FocusSession 聚焦指定会话，必要时尝试重新打开终端
func (a *App) FocusSession(id string) domain.ActionResult {
	serial := a.focusSerial.Add(1)
	s, ok := a.sessions.Get(id)
	if !ok {
		return domain.ActionResult{OK: false, Status: "missing", Message: "Session not found."}
	}
	if s.Status == "failed" {
		return a.reopenSession(s)
	}
	result := a.terminals.Focus(s)
	if serial != a.focusSerial.Load() {
		return domain.ActionResult{OK: false, Status: "cancelled", Message: "A newer focus request superseded this one."}
	}
	if !result.OK && shouldReopenMissingTerminal(s, result.Status) {
		return a.reopenSession(s)
	}
	if result.OK {
		a.sessions.UpdateStatus(id, "focused", "")
	}
	a.emit("sessions:updated", a.sessions.List())
	return result
}

// shouldReopenMissingTerminal 判断会话丢失时是否允许自动重开终端
func shouldReopenMissingTerminal(s domain.Session, status string) bool {
	switch status {
	case "gone":
		return s.AdapterID == "ghostty" || s.AdapterID == "terminal-app" || s.AdapterID == "windows-terminal"
	case "missing":
		return (s.AdapterID == "ghostty" || s.AdapterID == "terminal-app") && s.Ref.WindowID != ""
	default:
		return false
	}
}

// CloseSession 关闭或移除指定会话
func (a *App) CloseSession(id string) domain.ActionResult {
	s, ok := a.sessions.Get(id)
	if !ok {
		return domain.ActionResult{OK: false, Status: "missing", Message: "Session not found."}
	}
	result := a.terminals.Close(s)
	a.sessions.Remove(id)
	a.emit("sessions:updated", a.sessions.List())
	return result
}

// ClearSessions 清空会话列表
func (a *App) ClearSessions() domain.ActionResult {
	items := a.sessions.List()
	for _, s := range items {
		_ = a.terminals.Close(s)
	}
	removed := a.sessions.Clear()
	a.emit("sessions:updated", a.sessions.List())
	return domain.ActionResult{OK: true, Status: "cleared", Message: fmt.Sprintf("Removed %d sessions.", len(removed))}
}

// ListWorktrees 返回项目的 Git worktree 列表
func (a *App) ListWorktrees(projectID string) []domain.Worktree {
	project, ok := a.resolveProject(projectID)
	if !ok {
		return []domain.Worktree{}
	}
	return a.worktrees.List(project)
}

// ListWorktreeBranches 返回可用于创建 worktree 的分支列表
func (a *App) ListWorktreeBranches(projectID string) []string {
	project, ok := a.resolveProject(projectID)
	if !ok {
		return []string{}
	}
	return a.worktrees.Branches(project.Path)
}

// CreateWorktree 创建新的 Git worktree
func (a *App) CreateWorktree(req worktree.CreateRequest) (domain.Worktree, error) {
	project, ok := a.resolveProject(req.ProjectID)
	if !ok && req.ProjectPath == "" {
		return domain.Worktree{}, fmt.Errorf("project not found")
	}
	if req.ProjectPath == "" {
		req.ProjectPath = project.Path
	}
	if req.ProjectID == "" {
		req.ProjectID = project.ID
	}
	if req.BaseDir == "" {
		req.BaseDir = a.worktreeBaseDir(project)
	}
	created, err := a.worktrees.Create(req)
	a.emit("worktrees:updated", a.worktrees.List(project))
	return created, err
}

// RemoveWorktree 删除指定 Git worktree
func (a *App) RemoveWorktree(projectID string, path string, force bool) domain.ActionResult {
	project, ok := a.resolveProject(projectID)
	if !ok {
		return domain.ActionResult{OK: false, Status: "missing", Message: "Project not found."}
	}
	result := a.worktrees.Remove(project.Path, path, force)
	a.emit("worktrees:updated", a.worktrees.List(project))
	return result
}
