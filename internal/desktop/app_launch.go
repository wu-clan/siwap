package desktop

import (
	"strings"

	"siwap/internal/domain"
	"siwap/internal/harness"
	"siwap/internal/process"
	"siwap/internal/session"
	"siwap/internal/terminal"
)

func (a *App) LaunchSession(req session.LaunchRequest) domain.Session {
	prepared := a.prepareLaunch(req)
	sessionEnv := terminal.SessionID()
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
		a.emit("sessions:updated", a.sessions.List())
		return created
	}
	result.Ref.ProcessTreePIDs = process.TreePIDs(result.PID)
	created := a.sessions.Create(prepared, result, sessionEnv)
	a.emit("sessions:updated", a.sessions.List())
	return created
}

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
		a.emit("sessions:updated", a.sessions.List())
		return domain.ActionResult{OK: false, Status: "failed", Message: err.Error()}
	}
	result.Ref.ProcessTreePIDs = process.TreePIDs(result.PID)
	a.sessions.UpdateLaunch(existing.ID, result, sessionEnv)
	a.emit("sessions:updated", a.sessions.List())
	return domain.ActionResult{OK: true, Status: "reopened", Message: "Session terminal reopened."}
}

func terminalWindowTitle(sessionEnv string) string {
	return "Siwap " + sessionEnv
}

func (a *App) prepareLaunch(req session.LaunchRequest) session.LaunchRequest {
	prefs := a.config.Preferences()
	requestedAdapter := strings.TrimSpace(req.AdapterID)
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

func (a *App) adapterLaunchable(id string) bool {
	for _, adapter := range a.currentAdapters() {
		if adapter.ID == id {
			return adapter.Enabled && adapter.Installed
		}
	}
	return false
}

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

func (a *App) resolveProject(id string) (domain.Project, bool) {
	if id != "" {
		return a.projects.Get(id)
	}
	return a.projects.Selected()
}

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
