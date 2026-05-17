package terminal

import (
	"strings"
	"testing"

	"siwap/internal/domain"
)

func TestSplitArgsPreservesQuotedValues(t *testing.T) {
	got := splitArgs(`--title "Hello World" -e sh -lc 'pnpm dev'`)
	want := []string{"--title", "Hello World", "-e", "sh", "-lc", "pnpm dev"}
	if len(got) != len(want) {
		t.Fatalf("len=%d want %d: %#v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("arg %d=%q want %q; all=%#v", i, got[i], want[i], got)
		}
	}
}

func TestGhosttyLaunchArgsUseInlineConfigValues(t *testing.T) {
	req := LaunchRequest{
		WorkingDir: "/tmp/siwap-project",
		Title:      "Claude Code",
		Command:    "claude",
	}

	got := ghosttyLaunchArgs(req, "/bin/zsh")
	want := []string{
		"--working-directory=/tmp/siwap-project",
		"--title=Claude Code",
		"-e",
		"/bin/zsh",
		"-lc",
		"claude",
	}

	if len(got) != len(want) {
		t.Fatalf("len=%d want %d: %#v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("arg %d=%q want %q; all=%#v", i, got[i], want[i], got)
		}
	}
}

func TestRenderArgsKeepsPlaceholderValuesAtomic(t *testing.T) {
	req := LaunchRequest{
		WorkingDir: "/tmp/Siwap Project",
		Title:      "Claude Code",
		Command:    "claude --resume",
	}

	got := renderArgs(`--title {{title}} --cwd={{cwd}} -e sh -lc {{command}}`, req)
	want := []string{"--title", "Claude Code", "--cwd=/tmp/Siwap Project", "-e", "sh", "-lc", "claude --resume"}
	if len(got) != len(want) {
		t.Fatalf("len=%d want %d: %#v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("arg %d=%q want %q; all=%#v", i, got[i], want[i], got)
		}
	}
}

func TestRenderWorkingDirArgsSupportsSeparateInlineAndTemplateForms(t *testing.T) {
	req := LaunchRequest{WorkingDir: "/tmp/Siwap Project"}
	tests := []struct {
		name string
		flag string
		want []string
	}{
		{name: "separate", flag: "--working-directory", want: []string{"--working-directory", "/tmp/Siwap Project"}},
		{name: "inline", flag: "--working-directory=", want: []string{"--working-directory=/tmp/Siwap Project"}},
		{name: "template", flag: "--working-directory={{cwd}}", want: []string{"--working-directory=/tmp/Siwap Project"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := renderWorkingDirArgs(tt.flag, req)
			if len(got) != len(tt.want) {
				t.Fatalf("len=%d want %d: %#v", len(got), len(tt.want), got)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Fatalf("arg %d=%q want %q; all=%#v", i, got[i], tt.want[i], got)
				}
			}
		})
	}
}

func TestTerminalShellCommandDoesNotEmbedRawControlCharacters(t *testing.T) {
	req := LaunchRequest{
		WorkingDir: "/tmp/siwap-project",
		Title:      "Siwap siwap-123",
		Command:    "claude",
		Environment: map[string]string{
			"SIWAP_SESSION_ID": "siwap-123",
		},
	}

	got := terminalShellCommand(req)
	if strings.ContainsAny(got, "\033\007") {
		t.Fatalf("command should not embed raw terminal control bytes: %q", got)
	}
	for _, want := range []string{
		`printf '\033]0;%s\007' 'Siwap siwap-123'`,
		`cd '/tmp/siwap-project'`,
		`export SIWAP_SESSION_ID='siwap-123'`,
		`; clear; claude`,
		`claude`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("command %q does not contain %q", got, want)
		}
	}
}

func TestWindowsTerminalArgsKeepTabOpenInWorkingDirectory(t *testing.T) {
	req := LaunchRequest{
		WorkingDir: `C:\Users\c\My Project`,
		Title:      "Claude Code",
		Command:    "claude",
	}

	got := windowsTerminalArgs(req, "cmd.exe")
	want := []string{"-w", "0", "new-tab", "--title", "Claude Code", "--startingDirectory", `C:\Users\c\My Project`, "cmd.exe", "/K", "claude"}
	if len(got) != len(want) {
		t.Fatalf("len=%d want %d: %#v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("arg %d=%q want %q; all=%#v", i, got[i], want[i], got)
		}
	}
}

func TestLinuxTerminalCandidatesPreferExplicitWorkingDirectory(t *testing.T) {
	req := LaunchRequest{
		WorkingDir: "/home/c/My Project",
		Title:      "Claude Code",
		Command:    "claude",
	}
	got := linuxTerminalCandidates(req, "/bin/zsh")
	if len(got) < 2 {
		t.Fatalf("expected multiple candidates: %#v", got)
	}
	gnome := got[1]
	want := []string{"gnome-terminal", "--working-directory", "/home/c/My Project", "--title", "Claude Code", "--", "/bin/zsh", "-lc", "claude"}
	if len(gnome) != len(want) {
		t.Fatalf("len=%d want %d: %#v", len(gnome), len(want), gnome)
	}
	for i := range want {
		if gnome[i] != want[i] {
			t.Fatalf("arg %d=%q want %q; all=%#v", i, gnome[i], want[i], gnome)
		}
	}
}

func TestListDoesNotExposeGenericProcessAdapter(t *testing.T) {
	for _, adapter := range NewService().List() {
		if adapter.ID == "generic-process" {
			t.Fatal("generic shell process adapter should not be exposed")
		}
	}
}

func TestWorktreeSafeName(t *testing.T) {
	got := WorktreeSafeName("Claude Feature/One")
	if got != "claude-feature-one" {
		t.Fatalf("got %q", got)
	}
}

func TestFocusUnsupportedWithoutPID(t *testing.T) {
	result := NewService().Focus(domainSession("unknown-adapter", 0))
	if result.OK || result.Status != "unsupported" {
		t.Fatalf("unexpected focus result: %#v", result)
	}
}

func TestFocusDeadTrackedPidDoesNotRequestReopen(t *testing.T) {
	result := NewService().Focus(domainSession("custom-terminal", 999999))
	if !result.OK || result.Status != "unverified" {
		t.Fatalf("unexpected focus result: %#v", result)
	}
}

func TestCloseWithoutPIDRemovesOnly(t *testing.T) {
	result := NewService().Close(domainSession("ghostty", 0))
	if !result.OK || result.Status != "removed" {
		t.Fatalf("unexpected close result: %#v", result)
	}
}

func domainSession(adapterID string, pid int) domain.Session {
	return domain.Session{AdapterID: adapterID, PID: pid}
}
