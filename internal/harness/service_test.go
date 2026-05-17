package harness

import (
	"strings"
	"testing"

	"siwap/internal/domain"
)

func TestBuildCommandSkipsDefaultAndFalseFlags(t *testing.T) {
	h := domain.Harness{
		Command: "claude --resume",
		Flags: map[string]string{
			"dangerouslySkipPermissions": "false",
			"permissionMode":             "default",
		},
		FlagOptions: []domain.HarnessFlag{
			{Key: "dangerouslySkipPermissions", Type: "toggle", CommandFlag: "--dangerously-skip-permissions", Default: "false"},
			{Key: "permissionMode", Type: "select", CommandFlag: "--permission-mode", Default: "default", Options: []string{"default", "plan"}},
		},
	}
	got := BuildCommand(h, nil)
	if got != "claude --resume" {
		t.Fatalf("unexpected command: %q", got)
	}
}

func TestBuildCommandAddsToggleAndPicker(t *testing.T) {
	h := domain.Harness{
		Command: "claude",
		Flags: map[string]string{
			"dangerouslySkipPermissions": "true",
			"permissionMode":             "plan",
		},
		FlagOptions: []domain.HarnessFlag{
			{Key: "dangerouslySkipPermissions", Type: "toggle", CommandFlag: "--dangerously-skip-permissions", Default: "false"},
			{Key: "permissionMode", Type: "select", CommandFlag: "--permission-mode", Default: "default", Options: []string{"default", "plan"}},
		},
	}
	got := BuildCommand(h, nil)
	for _, want := range []string{"claude", "--dangerously-skip-permissions", "--permission-mode", "'plan'"} {
		if !strings.Contains(got, want) {
			t.Fatalf("%q does not contain %q", got, want)
		}
	}
}
