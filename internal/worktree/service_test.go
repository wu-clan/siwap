package worktree

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestParseWorktrees 验证对应功能行为
func TestParseWorktrees(t *testing.T) {
	data := "worktree /repo\nHEAD abc123\nbranch refs/heads/main\n\nworktree /repo.worktrees/feature\nHEAD def456\nbranch refs/heads/feature/test\n\n"
	items := parseWorktrees("project-1", "/repo", data)
	if len(items) != 2 {
		t.Fatalf("got %d items", len(items))
	}
	if !items[0].IsMain {
		t.Fatalf("expected project root to be marked as main: %#v", items[0])
	}
	if items[1].IsMain {
		t.Fatalf("expected secondary worktree to not be marked as main: %#v", items[1])
	}
	if items[1].Branch != "feature/test" || items[1].Path != "/repo.worktrees/feature" {
		t.Fatalf("unexpected parsed item: %#v", items[1])
	}
}

// TestRemoveProtectsMainWorktreeAndDefaultBranches 验证对应功能行为
func TestRemoveProtectsMainWorktreeAndDefaultBranches(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	root := t.TempDir()
	repo := filepath.Join(root, "repo")
	masterWorktree := filepath.Join(root, "repo-master")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatal(err)
	}
	runGit(t, repo, "init")
	runGit(t, repo, "checkout", "-b", "main")
	runGit(t, repo, "config", "user.email", "siwap@example.invalid")
	runGit(t, repo, "config", "user.name", "Siwap Test")
	if err := os.WriteFile(filepath.Join(repo, "README.md"), []byte("# repo\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, repo, "add", "README.md")
	runGit(t, repo, "commit", "-m", "initial")
	runGit(t, repo, "branch", "master")
	runGit(t, repo, "worktree", "add", masterWorktree, "master")

	service := NewService()
	if result := service.Remove(repo, repo, true); result.OK || result.Status != "protected" {
		t.Fatalf("expected main worktree to be protected, got %#v", result)
	}
	if result := service.Remove(repo, masterWorktree, true); result.OK || result.Status != "protected" {
		t.Fatalf("expected master branch worktree to be protected, got %#v", result)
	}
}

// TestBranchesListsExistingGitBranches 验证对应功能行为
func TestBranchesListsExistingGitBranches(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	repo := initTestRepo(t)
	runGit(t, repo, "branch", "feature/demo")

	branches := NewService().Branches(repo)
	if !containsBranch(branches, "main") || !containsBranch(branches, "feature/demo") {
		t.Fatalf("expected real git branches, got %#v", branches)
	}
}

// TestParseBranchesSkipsOriginBranches 验证对应功能行为
func TestParseBranchesSkipsOriginBranches(t *testing.T) {
	branches := parseBranches("main\norigin/main\nremotes/origin/dev\nfeature/demo\norigin/HEAD\n")
	if containsBranch(branches, "origin/main") || containsBranch(branches, "remotes/origin/dev") {
		t.Fatalf("origin branches should be filtered out: %#v", branches)
	}
	if !containsBranch(branches, "main") || !containsBranch(branches, "feature/demo") {
		t.Fatalf("local branches should be kept: %#v", branches)
	}
}

// TestCreateRejectsUnknownBaseBranch 验证对应功能行为
func TestCreateRejectsUnknownBaseBranch(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	root := t.TempDir()
	repo := initTestRepoAt(t, filepath.Join(root, "repo"))
	_, err := NewService().Create(CreateRequest{
		ProjectID:   "project-1",
		ProjectPath: repo,
		Branch:      "siwap/test",
		BaseBranch:  "missing",
		Path:        filepath.Join(root, "repo-worktree"),
	})
	if err == nil || !strings.Contains(err.Error(), "base branch does not exist") {
		t.Fatalf("expected missing base branch error, got %v", err)
	}
}

// TestCreateRequiresBranchName 验证对应功能行为
func TestCreateRequiresBranchName(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	root := t.TempDir()
	repo := initTestRepoAt(t, filepath.Join(root, "repo"))
	_, err := NewService().Create(CreateRequest{
		ProjectID:   "project-1",
		ProjectPath: repo,
		BaseBranch:  "main",
		Path:        filepath.Join(root, "repo-worktree"),
	})
	if err == nil || !strings.Contains(err.Error(), "branch is required") {
		t.Fatalf("expected branch required error, got %v", err)
	}
}

// runGit 在指定目录执行 Git 命令
func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %s %v", args, string(out), err)
	}
}

// initTestRepo 初始化测试用 Git 仓库
func initTestRepo(t *testing.T) string {
	return initTestRepoAt(t, filepath.Join(t.TempDir(), "repo"))
}

// initTestRepoAt 初始化测试用 Git 仓库
func initTestRepoAt(t *testing.T, repo string) string {
	t.Helper()
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatal(err)
	}
	runGit(t, repo, "init")
	runGit(t, repo, "checkout", "-b", "main")
	runGit(t, repo, "config", "user.email", "siwap@example.invalid")
	runGit(t, repo, "config", "user.name", "Siwap Test")
	if err := os.WriteFile(filepath.Join(repo, "README.md"), []byte("# repo\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, repo, "add", "README.md")
	runGit(t, repo, "commit", "-m", "initial")
	return repo
}
