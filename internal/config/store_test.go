package config

import (
	"os"
	"path/filepath"
	"testing"

	"siwap/internal/domain"
)

// TestUpdateHarnessCanDisableAssistant 验证对应功能行为
func TestUpdateHarnessCanDisableAssistant(t *testing.T) {
	store := testStore(t)

	updated, err := store.UpdateHarness(domain.Harness{
		ID:      "codex",
		Command: "codex",
		Enabled: false,
	})
	if err != nil {
		t.Fatalf("UpdateHarness returned error: %v", err)
	}
	if updated.Enabled {
		t.Fatal("UpdateHarness should preserve an explicit disabled state")
	}
}

// TestMergeHarnessesPreservesDisabledState 验证对应功能行为
func TestMergeHarnessesPreservesDisabledState(t *testing.T) {
	loaded := defaultConfig()
	loaded.Harnesses[1].Enabled = false

	merged := mergeConfig(loaded)
	for _, harness := range merged.Harnesses {
		if harness.ID == "codex" && harness.Enabled {
			t.Fatal("mergeConfig should preserve disabled built-in assistants")
		}
	}
}

// TestMergePreferencesPreservesWindowPosition 验证对应功能行为
func TestMergePreferencesPreservesWindowPosition(t *testing.T) {
	loaded := defaultConfig()
	loaded.Preferences.WindowX = 42
	loaded.Preferences.WindowY = 84
	loaded.Preferences.WindowPositionSaved = true

	merged := mergeConfig(loaded)
	if !merged.Preferences.WindowPositionSaved || merged.Preferences.WindowX != 42 || merged.Preferences.WindowY != 84 {
		t.Fatalf("window position was not preserved: %+v", merged.Preferences)
	}
}

// TestUpdatePreferencesCanClearSavedWindowPosition 验证对应功能行为
func TestUpdatePreferencesCanClearSavedWindowPosition(t *testing.T) {
	store := testStore(t)
	prefs := store.Preferences()
	prefs.WindowX = 42
	prefs.WindowY = 84
	prefs.WindowPositionSaved = true
	if _, err := store.UpdatePreferences(prefs); err != nil {
		t.Fatalf("saving position returned error: %v", err)
	}

	prefs = store.Preferences()
	prefs.WindowX = 0
	prefs.WindowY = 0
	prefs.WindowPositionSaved = false
	updated, err := store.UpdatePreferences(prefs)
	if err != nil {
		t.Fatalf("clearing position returned error: %v", err)
	}
	if updated.WindowPositionSaved || updated.WindowX != 0 || updated.WindowY != 0 {
		t.Fatalf("window position should be cleared: %+v", updated)
	}
}

// TestNormalizeProjectsUsesDefaultInsteadOfAllProjectsScope 验证对应功能行为
func TestNormalizeProjectsUsesDefaultInsteadOfAllProjectsScope(t *testing.T) {
	cfg := defaultConfig()
	cfg.Projects = []domain.Project{
		{ID: "project-a", Path: filepath.Join(t.TempDir(), "project-a"), IsDefault: true},
	}
	cfg.Preferences.DefaultProjectID = "project-a"
	cfg.Preferences.SelectedProjectID = AllProjectsScopeID

	normalizeProjects(&cfg)

	if cfg.Preferences.SelectedProjectID != "project-a" {
		t.Fatalf("selected project should fall back to default project, got %q", cfg.Preferences.SelectedProjectID)
	}
	if cfg.Preferences.DefaultProjectID != "project-a" {
		t.Fatalf("default project should remain real project, got %q", cfg.Preferences.DefaultProjectID)
	}
}

// TestSetDefaultProjectAlsoSelectsDefaultProject 验证对应功能行为
func TestSetDefaultProjectAlsoSelectsDefaultProject(t *testing.T) {
	store := testStore(t)
	root := t.TempDir()
	projectAPath := filepath.Join(root, "project-a")
	projectBPath := filepath.Join(root, "project-b")
	if err := os.MkdirAll(projectAPath, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(projectBPath, 0o755); err != nil {
		t.Fatal(err)
	}
	projectA, err := store.AddProject(projectAPath, "")
	if err != nil {
		t.Fatalf("AddProject A returned error: %v", err)
	}
	projectB, err := store.AddProject(projectBPath, "")
	if err != nil {
		t.Fatalf("AddProject B returned error: %v", err)
	}
	if _, err := store.SelectProject(AllProjectsScopeID); err != nil {
		t.Fatalf("SelectProject(all-projects) returned error: %v", err)
	}
	if _, err := store.SetDefaultProject(projectB.ID); err != nil {
		t.Fatalf("SetDefaultProject returned error: %v", err)
	}

	prefs := store.Preferences()
	if prefs.DefaultProjectID != projectB.ID || prefs.SelectedProjectID != projectB.ID {
		t.Fatalf("default project should also be selected, got default=%q selected=%q projectA=%q", prefs.DefaultProjectID, prefs.SelectedProjectID, projectA.ID)
	}
}

// TestSelectProjectSupportsAllProjectsScope 验证对应功能行为
func TestSelectProjectSupportsAllProjectsScope(t *testing.T) {
	store := testStore(t)
	projectDir := t.TempDir()
	project, err := store.AddProject(projectDir, "")
	if err != nil {
		t.Fatalf("AddProject returned error: %v", err)
	}
	if project.ID == "" {
		t.Fatal("AddProject returned empty project id")
	}

	selected, err := store.SelectProject(AllProjectsScopeID)
	if err != nil {
		t.Fatalf("SelectProject(all-projects) returned error: %v", err)
	}
	if selected.ID != "" {
		t.Fatalf("all-projects selection should not return a real project: %+v", selected)
	}
	if got := store.Preferences().SelectedProjectID; got != AllProjectsScopeID {
		t.Fatalf("selected project id should be all-projects scope, got %q", got)
	}
}

// testStore 创建测试用配置存储
func testStore(t *testing.T) *Store {
	t.Helper()
	return &Store{path: filepath.Join(t.TempDir(), "config.json"), config: defaultConfig()}
}
