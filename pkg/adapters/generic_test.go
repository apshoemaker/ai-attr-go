package adapters

import "testing"

func TestGenericMinimalJSON(t *testing.T) {
	input := `{
		"phase":"PostToolUse",
		"session_id":"gen-sess-1",
		"file_path":"/tmp/repo/src/app.js"
	}`
	adapter := NewGenericAdapter("windsurf")
	ctx, err := adapter.ParseHookInput(input)
	if err != nil {
		t.Fatal(err)
	}
	if ctx.Tool != "windsurf" {
		t.Errorf("expected windsurf, got %s", ctx.Tool)
	}
	if ctx.Phase != PostToolUse {
		t.Errorf("expected PostToolUse, got %v", ctx.Phase)
	}
	if ctx.SessionID != "gen-sess-1" {
		t.Errorf("expected gen-sess-1, got %s", ctx.SessionID)
	}
	if ctx.FilePath != "/tmp/repo/src/app.js" {
		t.Errorf("expected /tmp/repo/src/app.js, got %s", ctx.FilePath)
	}
}

func TestGenericWithModel(t *testing.T) {
	input := `{
		"phase":"PostToolUse",
		"session_id":"gen-sess-2",
		"file_path":"/tmp/f.rs",
		"model":"claude-opus-4-6"
	}`
	adapter := NewGenericAdapter("cursor")
	ctx, err := adapter.ParseHookInput(input)
	if err != nil {
		t.Fatal(err)
	}
	if ctx.Model != "claude-opus-4-6" {
		t.Errorf("expected claude-opus-4-6, got %s", ctx.Model)
	}
}

func TestGenericMissingRequiredFields(t *testing.T) {
	adapter := NewGenericAdapter("custom")

	// Missing phase
	_, err := adapter.ParseHookInput(`{"session_id":"s"}`)
	if err == nil {
		t.Error("expected error for missing phase")
	}

	// Missing session_id
	_, err = adapter.ParseHookInput(`{"phase":"post"}`)
	if err == nil {
		t.Error("expected error for missing session_id")
	}
}
