package desktop

import (
	"strings"

	"siwap/internal/domain"
)

// listAllWorktrees 聚合所有项目的 worktree，避免前端按项目循环查询
func (a *App) listAllWorktrees() []domain.Worktree {
	projects := a.projects.List()
	items := make([]domain.Worktree, 0, len(projects))
	for _, project := range projects {
		items = append(items, a.worktrees.List(project)...)
	}
	return items
}

// worktreeActionResult 返回动作结果和最新 worktree 快照，供前端直接同步状态
func (a *App) worktreeActionResult(action domain.ActionResult) domain.WorktreeActionResult {
	return domain.WorktreeActionResult{Action: action, Worktrees: a.listAllWorktrees()}
}

// worktreeActionResultWithWorktree 在动作结果中附带本次创建或操作的 worktree
func (a *App) worktreeActionResultWithWorktree(action domain.ActionResult, item domain.Worktree) domain.WorktreeActionResult {
	return domain.WorktreeActionResult{
		Action:    action,
		Worktree:  item,
		Worktrees: a.listAllWorktrees(),
	}
}

// worktreeBranchState 在 Go 侧过滤远端分支并计算默认基准分支
func worktreeBranchState(projectID string, branches []string) domain.WorktreeBranchState {
	filtered := make([]string, 0, len(branches))
	for _, branch := range branches {
		if !isOriginBranch(branch) {
			filtered = append(filtered, branch)
		}
	}
	return domain.WorktreeBranchState{
		ProjectID:         projectID,
		Branches:          filtered,
		DefaultBaseBranch: defaultBaseBranch(filtered),
	}
}

// isOriginBranch 判断分支名是否是 origin 远端引用
func isOriginBranch(branch string) bool {
	return branch == "origin" ||
		strings.HasPrefix(branch, "origin/") ||
		strings.HasPrefix(branch, "remotes/origin/")
}

// defaultBaseBranch 优先选择 main/master，否则回退到第一个可用分支
func defaultBaseBranch(branches []string) string {
	for _, branch := range branches {
		if branch == "main" {
			return branch
		}
	}
	for _, branch := range branches {
		if branch == "master" {
			return branch
		}
	}
	if len(branches) == 0 {
		return ""
	}
	return branches[0]
}
