package desktop

import (
	"path/filepath"
	"strings"

	"siwap/internal/domain"
)

// listSessions returns session snapshots enriched with project display names for the UI
func (a *App) listSessions() []domain.Session {
	return a.withSessionProjectNames(a.sessions.List())
}

// withSessionProjectName 为单个会话补齐项目展示名
func (a *App) withSessionProjectName(session domain.Session) domain.Session {
	return a.withSessionProjectNames([]domain.Session{session})[0]
}

// sessionActionResult 返回动作结果和最新会话列表，避免前端再次 ListSessions
func (a *App) sessionActionResult(action domain.ActionResult) domain.SessionActionResult {
	return domain.SessionActionResult{Action: action, Sessions: a.listSessions()}
}

// sessionActionResultWithSession 在动作结果中附带本次创建或重开的会话
func (a *App) sessionActionResultWithSession(action domain.ActionResult, session domain.Session) domain.SessionActionResult {
	return domain.SessionActionResult{
		Action:   action,
		Session:  a.withSessionProjectName(session),
		Sessions: a.listSessions(),
	}
}

// withSessionProjectNames 为会话列表批量注入项目展示名
func (a *App) withSessionProjectNames(sessions []domain.Session) []domain.Session {
	projects := a.projects.List()
	projectNames := make(map[string]string, len(projects))
	for _, project := range projects {
		projectNames[project.ID] = projectDisplayName(project)
	}
	for i := range sessions {
		sessions[i].ProjectName = projectNames[sessions[i].ProjectID]
	}
	return sessions
}

// projectDisplayName 返回项目在会话分组中使用的显示名
func projectDisplayName(project domain.Project) string {
	if label := strings.TrimSpace(project.Label); label != "" {
		return label
	}
	path := strings.TrimRight(project.Path, `\/`)
	if path == "" {
		return project.Path
	}
	base := filepath.Base(path)
	if base == "." || base == string(filepath.Separator) || base == "" {
		return project.Path
	}
	return base
}
