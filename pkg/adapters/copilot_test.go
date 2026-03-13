package adapters

import "testing"

func TestCopilotPreToolUse(t *testing.T) {
	input := `{
		"hookEventName":"PreToolUse",
		"toolName":"insert_edit_into_file",
		"chatSessionId":"copilot-sess-1",
		"toolInput":{"filePath":"/tmp/file.ts"}
	}`
	ctx, err := (&CopilotAdapter{}).ParseHookInput(input)
	if err != nil {
		t.Fatal(err)
	}
	if ctx.Phase != PreToolUse {
		t.Errorf("expected PreToolUse, got %v", ctx.Phase)
	}
	if ctx.Tool != "copilot" {
		t.Errorf("expected copilot, got %s", ctx.Tool)
	}
	if ctx.SessionID != "copilot-sess-1" {
		t.Errorf("expected copilot-sess-1, got %s", ctx.SessionID)
	}
	if ctx.FilePath != "/tmp/file.ts" {
		t.Errorf("expected /tmp/file.ts, got %s", ctx.FilePath)
	}
}

func TestCopilotCreateFile(t *testing.T) {
	input := `{
		"hookEventName":"PostToolUse",
		"toolName":"create_file",
		"chatSessionId":"sess-2",
		"toolInput":{"filePath":"/tmp/new.ts"}
	}`
	ctx, err := (&CopilotAdapter{}).ParseHookInput(input)
	if err != nil {
		t.Fatal(err)
	}
	if ctx.Phase != PostToolUse {
		t.Errorf("expected PostToolUse, got %v", ctx.Phase)
	}
	if ctx.FilePath != "/tmp/new.ts" {
		t.Errorf("expected /tmp/new.ts, got %s", ctx.FilePath)
	}
}

func TestCopilotExtractsModel(t *testing.T) {
	input := `{
		"hookEventName":"PostToolUse",
		"toolName":"insert_edit_into_file",
		"chatSessionId":"sess-3",
		"model":"gpt-4o",
		"toolInput":{"filePath":"/tmp/f.ts"}
	}`
	ctx, err := (&CopilotAdapter{}).ParseHookInput(input)
	if err != nil {
		t.Fatal(err)
	}
	if ctx.Model != "gpt-4o" {
		t.Errorf("expected gpt-4o, got %s", ctx.Model)
	}
}

func TestCopilotModelIdVariant(t *testing.T) {
	input := `{
		"hookEventName":"PostToolUse",
		"toolName":"create_file",
		"chatSessionId":"sess-4",
		"modelId":"gpt-4-turbo",
		"toolInput":{"filePath":"/tmp/f.ts"}
	}`
	ctx, err := (&CopilotAdapter{}).ParseHookInput(input)
	if err != nil {
		t.Fatal(err)
	}
	if ctx.Model != "gpt-4-turbo" {
		t.Errorf("expected gpt-4-turbo, got %s", ctx.Model)
	}
}

func TestCopilotRejectsUnknownTool(t *testing.T) {
	input := `{
		"hookEventName":"PreToolUse",
		"toolName":"run_terminal_command",
		"chatSessionId":"sess-1",
		"toolInput":{}
	}`
	_, err := (&CopilotAdapter{}).ParseHookInput(input)
	if err == nil {
		t.Error("expected error for unknown tool")
	}
}
