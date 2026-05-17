package desktop

import (
	"strings"
	"testing"

	"siwap/internal/domain"
	"siwap/internal/session"
)

func TestSanitizeSettingsSection(t *testing.T) {
	for _, section := range []string{"general", "projects", "worktrees", "terminal", "ai"} {
		if got := sanitizeSettingsSection(section); got != section {
			t.Fatalf("sanitizeSettingsSection(%q) = %q", section, got)
		}
	}
	if got := sanitizeSettingsSection("bad"); got != "general" {
		t.Fatalf("sanitizeSettingsSection should fall back to general, got %q", got)
	}
}

func TestParseSettingsTargetSupportsWorktreeCreate(t *testing.T) {
	target := parseSettingsTarget("worktrees:create")
	if target.Section != "worktrees" || target.Action != "create" {
		t.Fatalf("unexpected target: %#v", target)
	}
	if payload := settingsTargetPayload(target); payload != "worktrees:create" {
		t.Fatalf("payload=%q", payload)
	}
	if route := settingsWindowRoute(target); !strings.Contains(route, "section=worktrees") || !strings.Contains(route, "action=create") {
		t.Fatalf("route missing worktree create query: %q", route)
	}
}

func TestPositivePanelWidth(t *testing.T) {
	tests := map[int]int{
		0:   320,
		-10: 320,
		120: 260,
		360: 360,
		900: 520,
	}
	for input, want := range tests {
		if got := positivePanelWidth(input); got != want {
			t.Fatalf("positivePanelWidth(%d) = %d, want %d", input, got, want)
		}
	}
}

func TestCursorAtExactLeftEdge(t *testing.T) {
	tests := []struct {
		name      string
		x         int
		y         int
		left      int
		top       int
		height    int
		hasScreen bool
		want      bool
	}{
		{name: "primary left edge", x: 0, y: 100, left: 0, top: 0, height: 900, hasScreen: true, want: true},
		{name: "inside old hot zone but not exact edge", x: 1, y: 100, left: 0, top: 0, height: 900, hasScreen: true, want: false},
		{name: "secondary screen negative left edge", x: -1920, y: 100, left: -1920, top: 0, height: 900, hasScreen: true, want: true},
		{name: "not secondary screen edge", x: -1919, y: 100, left: -1920, top: 0, height: 900, hasScreen: true, want: false},
		{name: "outside vertical screen bounds", x: 0, y: -1, left: 0, top: 0, height: 900, hasScreen: true, want: false},
		{name: "fallback left", x: 0, y: 100, hasScreen: false, want: true},
		{name: "fallback not left", x: 1, y: 100, hasScreen: false, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cursorAtExactLeftEdge(tt.x, tt.y, tt.left, tt.top, tt.height, tt.hasScreen)
			if got != tt.want {
				t.Fatalf("cursorAtExactLeftEdge() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPointInRect(t *testing.T) {
	tests := []struct {
		name   string
		x      int
		y      int
		left   int
		top    int
		width  int
		height int
		want   bool
	}{
		{name: "inside", x: 10, y: 10, left: 0, top: 0, width: 320, height: 900, want: true},
		{name: "left edge included", x: 0, y: 10, left: 0, top: 0, width: 320, height: 900, want: true},
		{name: "right edge excluded", x: 320, y: 10, left: 0, top: 0, width: 320, height: 900, want: false},
		{name: "bottom edge excluded", x: 10, y: 900, left: 0, top: 0, width: 320, height: 900, want: false},
		{name: "outside", x: -1, y: 10, left: 0, top: 0, width: 320, height: 900, want: false},
		{name: "invalid size", x: 0, y: 0, left: 0, top: 0, width: 0, height: 900, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pointInRect(tt.x, tt.y, tt.left, tt.top, tt.width, tt.height)
			if got != tt.want {
				t.Fatalf("pointInRect() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseShortcut(t *testing.T) {
	mods, key, ok := parseShortcut("Control+Command+S")
	if !ok || len(mods) != 2 || key == 0 {
		t.Fatalf("default shortcut should parse, mods=%v key=%v ok=%v", mods, key, ok)
	}
	if _, _, ok := parseShortcut("Control+S+K"); ok {
		t.Fatal("shortcut with multiple keys should be invalid")
	}
	if _, _, ok := parseShortcut("Control+Nope"); ok {
		t.Fatal("unknown key should be invalid")
	}
	if _, _, ok := parseShortcut("Control+Command"); ok {
		t.Fatal("shortcut without key should be invalid")
	}
}

func TestApplyTemplate(t *testing.T) {
	request := session.LaunchRequest{
		HarnessID:    "codex",
		ProjectID:    "project-1",
		Command:      "codex run",
		WorkingDir:   "/tmp/project",
		Title:        "Codex",
		WorktreePath: "/tmp/project-wt",
	}
	got := applyTemplate("cd {{cwd}} && {{command}} --title {{title}} --wt {{worktreePath}} --h {{harnessID}} --p {{projectID}}", request)
	for _, want := range []string{"/tmp/project", "codex run", "Codex", "/tmp/project-wt", "codex", "project-1"} {
		if !strings.Contains(got, want) {
			t.Fatalf("template output %q does not contain %q", got, want)
		}
	}
}

func TestNilWindowActionsDoNotPanic(t *testing.T) {
	app := NewApp()
	if result := app.ShowWindow(); result.OK || result.Status != "missing" {
		t.Fatalf("ShowWindow without Wails window = %+v", result)
	}
	if result := app.HideWindow(); result.OK || result.Status != "missing" {
		t.Fatalf("HideWindow without Wails window = %+v", result)
	}
	if app.settingsWindowIsVisible() {
		t.Fatal("nil settings window should not be visible")
	}
}

func TestCurrentAdaptersApplyDisabledPreference(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	app := NewApp()
	prefs := app.config.Preferences()
	prefs.DisabledTerminalIDs = []string{"ghostty"}
	if _, err := app.config.UpdatePreferences(prefs); err != nil {
		t.Fatalf("UpdatePreferences returned error: %v", err)
	}

	for _, adapter := range app.currentAdapters() {
		if adapter.ID == "ghostty" && adapter.Enabled {
			t.Fatal("disabled terminal adapter should not be enabled")
		}
		if adapter.ID == "auto" && !adapter.Enabled {
			t.Fatal("auto adapter should always remain enabled")
		}
	}
}

func TestCurrentAdaptersDoNotExposeGenericProcess(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	app := NewApp()

	for _, adapter := range app.currentAdapters() {
		if adapter.ID == "generic-process" {
			t.Fatal("generic shell process adapter should not be exposed")
		}
	}
}

func TestPrepareLaunchFallsBackFromDisabledDefaultAdapter(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	app := NewApp()
	prefs := app.config.Preferences()
	prefs.DefaultAdapterID = "ghostty"
	prefs.DisabledTerminalIDs = []string{"ghostty"}
	if _, err := app.config.UpdatePreferences(prefs); err != nil {
		t.Fatalf("UpdatePreferences returned error: %v", err)
	}

	prepared := app.prepareLaunch(session.LaunchRequest{})
	if prepared.AdapterID == "ghostty" || prepared.AdapterID == "auto" {
		t.Fatalf("disabled default adapter should not be selected directly, got %q", prepared.AdapterID)
	}
}

func TestPrepareLaunchHonorsExplicitAdapter(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	app := NewApp()
	prefs := app.config.Preferences()
	prefs.DisabledTerminalIDs = []string{"ghostty"}
	if _, err := app.config.UpdatePreferences(prefs); err != nil {
		t.Fatalf("UpdatePreferences returned error: %v", err)
	}

	prepared := app.prepareLaunch(session.LaunchRequest{AdapterID: " ghostty "})
	if prepared.AdapterID != "ghostty" {
		t.Fatalf("explicit adapter should be preserved, got %q", prepared.AdapterID)
	}
}

func TestShouldReopenMissingTerminalSkipsUnsupportedAdapters(t *testing.T) {
	custom := domain.Session{AdapterID: "custom-terminal", PID: 123}
	if shouldReopenMissingTerminal(custom, "unsupported") || shouldReopenMissingTerminal(custom, "gone") {
		t.Fatal("custom or unsupported focus results must not reopen automatically")
	}

	ghostty := domain.Session{AdapterID: "ghostty", Ref: domain.TerminalSessionRef{WindowID: "42"}}
	if !shouldReopenMissingTerminal(ghostty, "missing") {
		t.Fatal("missing managed Ghostty window should reopen")
	}
	terminalApp := domain.Session{AdapterID: "terminal-app", Ref: domain.TerminalSessionRef{WindowID: "42"}}
	if !shouldReopenMissingTerminal(terminalApp, "missing") {
		t.Fatal("missing managed Terminal.app window should reopen")
	}
}

func TestUpdateHarnessCanDisableAssistantThroughApp(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	app := NewApp()

	updated, err := app.UpdateHarness(appHarness("codex", false))
	if err != nil {
		t.Fatalf("UpdateHarness returned error: %v", err)
	}
	if updated.Enabled {
		t.Fatal("app UpdateHarness should preserve disabled assistants")
	}
}

func appHarness(id string, enabled bool) domain.Harness {
	return domain.Harness{ID: id, Command: id, Enabled: enabled}
}
