package adapters

import "testing"

func TestCodexAgentTurnComplete(t *testing.T) {
	input := `{
		"event_type":"agent_turn_complete",
		"session_id":"codex-sess-abc"
	}`
	ctx, err := (&CodexAdapter{}).ParseHookInput(input)
	if err != nil {
		t.Fatal(err)
	}
	if ctx.Phase != PostToolUse {
		t.Errorf("expected PostToolUse, got %v", ctx.Phase)
	}
	if ctx.Tool != "codex" {
		t.Errorf("expected codex, got %s", ctx.Tool)
	}
	if ctx.SessionID != "codex-sess-abc" {
		t.Errorf("expected codex-sess-abc, got %s", ctx.SessionID)
	}
}

func TestCodexReturnsNoFilePath(t *testing.T) {
	input := `{
		"event_type":"agent_turn_complete",
		"session_id":"codex-sess-1"
	}`
	ctx, err := (&CodexAdapter{}).ParseHookInput(input)
	if err != nil {
		t.Fatal(err)
	}
	if ctx.FilePath != "" {
		t.Errorf("expected empty file_path, got %s", ctx.FilePath)
	}
}
