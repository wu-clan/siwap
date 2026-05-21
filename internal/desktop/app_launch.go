package desktop

import (
	"strings"

	"siwap/internal/domain"
	"siwap/internal/harness"
	"siwap/internal/process"
	"siwap/internal/session"
	"siwap/internal/terminal"
)

// LaunchSession 根据前端传入的请求启动一个新的助手终端会话，并返回最新会话列表
func (a *App) LaunchSession(req session.LaunchRequest) domain.SessionActionResult {
	prepared := a.prepareLaunch(req)
	sessionEnv := terminal.SessionID()
	// 注入会话 ID，后续聚焦、关闭、重开终端时都通过它识别同一个会话
	env := map[string]string{
		domain.SessionEnvironmentKey: sessionEnv,
		"SIWAP_HARNESS_ID":           prepared.HarnessID,
		"SIWAP_PROJECT_ID":           prepared.ProjectID,
	}
	if prepared.WorktreePath != "" {
		env["SIWAP_WORKTREE_PATH"] = prepared.WorktreePath
	}

	result, err := a.terminals.LaunchWithProfiles(terminal.LaunchRequest{
		AdapterID:   prepared.AdapterID,
		Title:       terminalWindowTitle(sessionEnv),
		Command:     prepared.Command,
		WorkingDir:  prepared.WorkingDir,
		Environment: env,
		Background:  false,
	}, a.config.ListTerminalProfiles())
	if err != nil {
		created := a.sessions.MarkError(prepared, err, sessionEnv)
		result := a.sessionActionResultWithSession(
			domain.ActionResult{OK: false, Status: "failed", Message: err.Error()},
			created,
		)
		a.emit("sessions:updated", result.Sessions)
		return result
	}
	result.Ref.ProcessTreePIDs = process.TreePIDs(result.PID)
	created := a.sessions.Create(prepared, result, sessionEnv)
	action := domain.ActionResult{OK: true, Status: result.Status, Message: "Session launched."}
	out := a.sessionActionResultWithSession(action, created)
	a.emit("sessions:updated", out.Sessions)
	return out
}

// reopenSession 使用原始会话信息重新打开终端
func (a *App) reopenSession(existing domain.Session) domain.ActionResult {
	sessionEnv := terminal.SessionID()
	env := map[string]string{
		domain.SessionEnvironmentKey: sessionEnv,
		"SIWAP_HARNESS_ID":           existing.HarnessID,
		"SIWAP_PROJECT_ID":           existing.ProjectID,
	}
	if existing.WorktreePath != "" {
		env["SIWAP_WORKTREE_PATH"] = existing.WorktreePath
	}
	result, err := a.terminals.LaunchWithProfiles(terminal.LaunchRequest{
		AdapterID:   existing.AdapterID,
		Title:       terminalWindowTitle(sessionEnv),
		Command:     existing.Command,
		WorkingDir:  existing.WorkingDir,
		Environment: env,
		Background:  false,
	}, a.config.ListTerminalProfiles())
	if err != nil {
		a.sessions.UpdateStatus(existing.ID, "failed", err.Error())
		a.emit("sessions:updated", a.listSessions())
		return domain.ActionResult{OK: false, Status: "failed", Message: err.Error()}
	}
	result.Ref.ProcessTreePIDs = process.TreePIDs(result.PID)
	a.sessions.UpdateLaunch(existing.ID, result, sessionEnv)
	a.emit("sessions:updated", a.listSessions())
	return domain.ActionResult{OK: true, Status: "reopened", Message: "Session terminal reopened."}
}

// terminalWindowTitle 生成包含会话 ID 的终端窗口标题
func terminalWindowTitle(sessionEnv string) string {
	return "Siwap " + sessionEnv
}

// prepareLaunch 补齐启动请求中的项目、终端、工作目录和助手命令
func (a *App) prepareLaunch(req session.LaunchRequest) session.LaunchRequest {
	prefs := a.config.Preferences()
	requestedAdapter := strings.TrimSpace(req.AdapterID)
	// 终端为空、已禁用或不可启动时回退到 auto，避免用户配置变更后启动失败
	if requestedAdapter == "" {
		req.AdapterID = prefs.DefaultAdapterID
		if req.AdapterID == "" || stringSet(prefs.DisabledTerminalIDs)[req.AdapterID] || (req.AdapterID != "auto" && !a.adapterLaunchable(req.AdapterID)) {
			req.AdapterID = "auto"
		}
	} else {
		req.AdapterID = requestedAdapter
	}
	if req.AdapterID == "auto" {
		req.AdapterID = a.bestEnabledAdapterID()
	}
	if req.ProjectID == "" {
		if selected, ok := a.projects.Selected(); ok {
			req.ProjectID = selected.ID
		}
	}
	project, hasProject := a.resolveProject(req.ProjectID)
	if req.WorkingDir == "" && hasProject {
		req.WorkingDir = project.Path
	}
	if req.WorktreePath != "" {
		req.WorkingDir = req.WorktreePath
	}
	if req.HarnessID == "" {
		for _, h := range a.harnesses.List() {
			if h.Enabled {
				req.HarnessID = h.ID
				break
			}
		}
	}
	if req.Command == "" {
		if h, ok := a.harnesses.Get(req.HarnessID); ok {
			req.Command = harness.BuildCommand(h, req.FlagOverrides)
			if req.Title == "" {
				req.Title = h.Label
			}
		}
	}
	if req.Command == "" {
		req.Command = "echo 'No assistant command configured'"
	}
	if prefs.TerminalCommandTemplate != "" && strings.Contains(prefs.TerminalCommandTemplate, "{{command}}") {
		req.Command = applyTemplate(prefs.TerminalCommandTemplate, req)
	}
	if req.Title == "" {
		req.Title = req.HarnessID
	}
	return req
}

// bestEnabledAdapterID 返回当前最合适的已启用终端适配器
func (a *App) bestEnabledAdapterID() string {
	for _, adapter := range a.currentAdapters() {
		if adapter.ID == "auto" {
			continue
		}
		if adapter.Enabled && adapter.Installed {
			return adapter.ID
		}
	}
	return ""
}

// adapterLaunchable 判断终端适配器是否可用于启动
func (a *App) adapterLaunchable(id string) bool {
	for _, adapter := range a.currentAdapters() {
		if adapter.ID == id {
			return adapter.Enabled && adapter.Installed
		}
	}
	return false
}

// worktreeBaseDir 返回创建 worktree 时使用的基础目录
func (a *App) worktreeBaseDir(project domain.Project) string {
	prefs := a.config.Preferences()
	if prefs.WorktreeBaseDir != "" {
		return prefs.WorktreeBaseDir
	}
	if prefs.WorktreeLocation == "project-parent" && project.Path != "" {
		return ""
	}
	return ""
}

// resolveProject 根据项目 ID 解析项目配置
func (a *App) resolveProject(id string) (domain.Project, bool) {
	if id != "" {
		return a.projects.Get(id)
	}
	return a.projects.Selected()
}

// applyTemplate 将启动请求应用到用户自定义命令模板
func applyTemplate(template string, req session.LaunchRequest) string {
	replacements := map[string]string{
		"{{command}}":      req.Command,
		"{{cwd}}":          req.WorkingDir,
		"{{title}}":        req.Title,
		"{{harnessID}}":    req.HarnessID,
		"{{projectID}}":    req.ProjectID,
		"{{worktreePath}}": req.WorktreePath,
	}
	out := template
	for key, value := range replacements {
		out = strings.ReplaceAll(out, key, value)
	}
	return out
}
