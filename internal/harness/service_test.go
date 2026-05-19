package harness

import (
	"testing"

	"siwap/internal/domain"
)

// TestBuildCommandReturnsConfiguredCommandOnly 验证对应功能行为
func TestBuildCommandReturnsConfiguredCommandOnly(t *testing.T) {
	h := domain.Harness{
		Command: "codex --dangerously-bypass-approvals-and-sandbox",
		Flags: map[string]string{
			"hidden": "true",
		},
		FlagOptions: []domain.HarnessFlag{
			{Key: "hidden", Type: "toggle", CommandFlag: "--must-not-append", Default: "false"},
		},
	}
	got := BuildCommand(h, map[string]string{"hidden": "true"})
	want := "codex --dangerously-bypass-approvals-and-sandbox"
	if got != want {
		t.Fatalf("command=%q want %q", got, want)
	}
}
