package session

import (
	"testing"

	"siwap/internal/domain"
	"siwap/internal/terminal"
)

// TestCreateUsesActualLaunchedAdapter 验证对应功能行为
func TestCreateUsesActualLaunchedAdapter(t *testing.T) {
	svc := NewService()
	created := svc.Create(LaunchRequest{
		HarnessID: "claude",
		AdapterID: "ghostty",
		Title:     "Claude Code",
	}, terminal.LaunchResult{
		Status: "launched",
		Ref: domain.TerminalSessionRef{
			AdapterID: "terminal-app",
			Title:     "Claude Code",
		},
	}, "siwap-test")

	if created.AdapterID != "terminal-app" {
		t.Fatalf("session should record actual launched adapter, got %q", created.AdapterID)
	}
}
