package worktree

import (
	"bufio"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"siwap/internal/domain"
	"siwap/internal/terminal"
)

// CreateRequest 表示创建 Git worktree 所需的参数
type CreateRequest struct {
	ProjectID   string `json:"projectId"`
	ProjectPath string `json:"projectPath"`
	Branch      string `json:"branch"`
	BaseBranch  string `json:"baseBranch"`
	Path        string `json:"path"`
	BaseDir     string `json:"baseDir"`
}

// Service 提供 Git worktree 的查询、创建和删除能力
type Service struct{}

// NewService 创建 worktree 服务
func NewService() *Service {
	return &Service{}
}

// List 返回指定仓库的 worktree 列表
func (s *Service) List(project domain.Project) []domain.Worktree {
	if !isGitRepo(project.Path) {
		return []domain.Worktree{}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", "-C", project.Path, "worktree", "list", "--porcelain")
	out, err := cmd.Output()
	if err != nil {
		return []domain.Worktree{}
	}
	return parseWorktrees(project.ID, project.Path, string(out))
}

// Branches 返回仓库中的本地分支列表
func (s *Service) Branches(projectPath string) []string {
	projectPath = strings.TrimSpace(projectPath)
	if projectPath == "" || !isGitRepo(projectPath) {
		return []string{}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", "-C", projectPath, "for-each-ref", "--format=%(refname:short)", "refs/heads")
	out, err := cmd.Output()
	if err != nil {
		return []string{}
	}
	return parseBranches(string(out))
}

// Create 根据请求创建新的 Git worktree
func (s *Service) Create(req CreateRequest) (domain.Worktree, error) {
	projectPath := strings.TrimSpace(req.ProjectPath)
	if projectPath == "" {
		return domain.Worktree{}, errors.New("project path is required")
	}
	if !isGitRepo(projectPath) {
		return domain.Worktree{}, fmt.Errorf("not a git repository: %s", projectPath)
	}
	baseBranch := strings.TrimSpace(req.BaseBranch)
	branches := s.Branches(projectPath)
	if len(branches) == 0 {
		return domain.Worktree{}, fmt.Errorf("git repository has no branches: %s", projectPath)
	}
	if baseBranch != "" && !containsBranch(branches, baseBranch) {
		return domain.Worktree{}, fmt.Errorf("base branch does not exist: %s", baseBranch)
	}
	branch := strings.TrimSpace(req.Branch)
	if branch == "" {
		return domain.Worktree{}, errors.New("branch is required")
	}
	branch = safeBranch(branch)
	if branch == "" {
		return domain.Worktree{}, errors.New("branch is required")
	}
	target := strings.TrimSpace(req.Path)
	if target == "" {
		baseDir := strings.TrimSpace(req.BaseDir)
		if baseDir == "" {
			baseDir = filepath.Join(filepath.Dir(projectPath), filepath.Base(projectPath)+".worktrees")
		}
		target = filepath.Join(baseDir, terminal.WorktreeSafeName(branch))
	}
	if _, err := os.Stat(target); err == nil {
		return domain.Worktree{}, fmt.Errorf("worktree path already exists: %s", target)
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return domain.Worktree{}, err
	}

	args := []string{"-C", projectPath, "worktree", "add"}
	if baseBranch != "" {
		args = append(args, "-b", branch, target, baseBranch)
	} else {
		args = append(args, "-b", branch, target, defaultBaseBranch(branches))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return domain.Worktree{}, fmt.Errorf("git worktree add failed: %s %w", strings.TrimSpace(string(out)), err)
	}
	return domain.Worktree{
		ID:         worktreeID(target),
		ProjectID:  req.ProjectID,
		Path:       target,
		Branch:     branch,
		BaseBranch: baseBranch,
		Exists:     true,
		Status:     "created",
		CreatedAt:  time.Now().Format(time.RFC3339),
	}, nil
}

// Remove 删除指定 Git worktree
func (s *Service) Remove(projectPath string, path string, force bool) domain.ActionResult {
	projectPath = strings.TrimSpace(projectPath)
	path = strings.TrimSpace(path)
	if projectPath == "" || path == "" {
		return domain.ActionResult{OK: false, Status: "invalid", Message: "project path and worktree path are required"}
	}
	if !isGitRepo(projectPath) {
		return domain.ActionResult{OK: false, Status: "invalid", Message: "project path is not a git repository"}
	}
	target, ok := s.findManagedWorktree(projectPath, path)
	if !ok {
		return domain.ActionResult{OK: false, Status: "not-found", Message: "Worktree does not belong to this project."}
	}
	if target.IsMain {
		return domain.ActionResult{OK: false, Status: "protected", Message: "Main worktree cannot be removed."}
	}
	if isProtectedBranch(target.Branch) {
		return domain.ActionResult{OK: false, Status: "protected", Message: "Default branch worktree cannot be removed."}
	}
	if isDirty(target.Path) && !force {
		return domain.ActionResult{OK: false, Status: "dirty", Message: "Worktree has uncommitted changes. Commit, stash, or remove manually with --force."}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	args := []string{"-C", projectPath, "worktree", "remove"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, target.Path)
	cmd := exec.CommandContext(ctx, "git", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return domain.ActionResult{OK: false, Status: "failed", Message: strings.TrimSpace(string(out)) + " " + err.Error()}
	}
	return domain.ActionResult{OK: true, Status: "removed", Message: "Worktree removed."}
}

// findManagedWorktree 在仓库中查找受管理的 worktree
func (s *Service) findManagedWorktree(projectPath string, path string) (domain.Worktree, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", "-C", projectPath, "worktree", "list", "--porcelain")
	out, err := cmd.Output()
	if err != nil {
		return domain.Worktree{}, false
	}
	for _, item := range parseWorktrees("", projectPath, string(out)) {
		if samePath(item.Path, path) {
			return item, true
		}
	}
	return domain.Worktree{}, false
}

// isGitRepo 判断目录是否为 Git 仓库
func isGitRepo(path string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", "-C", path, "rev-parse", "--is-inside-work-tree")
	out, err := cmd.Output()
	return err == nil && strings.TrimSpace(string(out)) == "true"
}

// parseWorktrees 解析 git worktree list --porcelain 输出
func parseWorktrees(projectID string, projectPath string, data string) []domain.Worktree {
	var out []domain.Worktree
	current := domain.Worktree{ProjectID: projectID, Exists: true, Status: "ready"}
	scanner := bufio.NewScanner(strings.NewReader(data))
	flush := func() {
		if current.Path != "" {
			current.ID = worktreeID(current.Path)
			current.IsMain = samePath(current.Path, projectPath)
			current.Dirty = isDirty(current.Path)
			if current.Dirty {
				current.Status = "dirty"
			}
			out = append(out, current)
		}
		current = domain.Worktree{ProjectID: projectID, Exists: true, Status: "ready"}
	}
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			flush()
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			continue
		}
		switch parts[0] {
		case "worktree":
			current.Path = parts[1]
		case "HEAD":
			current.Head = parts[1]
		case "branch":
			current.Branch = strings.TrimPrefix(parts[1], "refs/heads/")
		}
	}
	flush()
	return out
}

// isProtectedBranch 判断分支是否为受保护的主分支
func isProtectedBranch(branch string) bool {
	switch strings.TrimSpace(branch) {
	case "main", "master":
		return true
	default:
		return false
	}
}

// samePath 判断两个路径是否指向同一位置
func samePath(a string, b string) bool {
	a = strings.TrimSpace(a)
	b = strings.TrimSpace(b)
	if a == "" || b == "" {
		return false
	}
	if aInfo, aErr := os.Stat(a); aErr == nil {
		if bInfo, bErr := os.Stat(b); bErr == nil {
			return os.SameFile(aInfo, bInfo)
		}
	}
	aAbs, aErr := filepath.Abs(a)
	if aErr == nil {
		a = aAbs
	}
	bAbs, bErr := filepath.Abs(b)
	if bErr == nil {
		b = bAbs
	}
	return filepath.Clean(a) == filepath.Clean(b)
}

// safeBranch 将分支名称转换为可用于路径的安全名称
func safeBranch(branch string) string {
	branch = strings.TrimSpace(branch)
	branch = strings.ReplaceAll(branch, " ", "-")
	branch = strings.ReplaceAll(branch, "\\", "-")
	branch = strings.Trim(branch, "/")
	return branch
}

// parseBranches 解析 git branch --format 输出
func parseBranches(data string) []string {
	seen := map[string]bool{}
	var out []string
	scanner := bufio.NewScanner(strings.NewReader(data))
	for scanner.Scan() {
		branch := strings.TrimSpace(scanner.Text())
		if branch == "" || isOriginBranch(branch) || strings.HasSuffix(branch, "/HEAD") || seen[branch] {
			continue
		}
		seen[branch] = true
		out = append(out, branch)
	}
	return out
}

func isOriginBranch(branch string) bool {
	return branch == "origin" || strings.HasPrefix(branch, "origin/") || strings.HasPrefix(branch, "remotes/origin/")
}

// containsBranch 判断分支列表是否包含指定分支
func containsBranch(branches []string, target string) bool {
	for _, branch := range branches {
		if branch == target {
			return true
		}
	}
	return false
}

// defaultBaseBranch 返回仓库默认基准分支
func defaultBaseBranch(branches []string) string {
	for _, name := range []string{"main", "master"} {
		for _, branch := range branches {
			if branch == name || strings.HasSuffix(branch, "/"+name) {
				return branch
			}
		}
	}
	return branches[0]
}

// isDirty 判断 Git 工作区是否存在未提交修改
func isDirty(path string) bool {
	if !isGitRepo(path) {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", "-C", path, "status", "--porcelain")
	out, err := cmd.Output()
	return err == nil && strings.TrimSpace(string(out)) != ""
}

// worktreeID 根据 worktree 路径生成稳定 ID
func worktreeID(path string) string {
	sum := sha1.Sum([]byte(path))
	return "worktree-" + hex.EncodeToString(sum[:])[:12]
}
