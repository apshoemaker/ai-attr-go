package adapters

import "testing"

func TestClinePostToolUse(t *testing.T) {
	input := `{
		"hook_event_name":"PostToolUse",
		"taskId":"cline-task-abc",
		"tool_input":{"file_path":"/tmp/repo/src/lib.rs"}
	}`
	ctx, err := (&ClineAdapter{}).ParseHookInput(input)
	if err != nil {
		t.Fatal(err)
	}
	if ctx.Phase != PostToolUse {
		t.Errorf("expected PostToolUse, got %v", ctx.Phase)
	}
	if ctx.SessionID != "cline-task-abc" {
		t.Errorf("expected cline-task-abc, got %s", ctx.SessionID)
	}
	if ctx.Tool != "cline" {
		t.Errorf("expected cline, got %s", ctx.Tool)
	}
}

func TestClineFilePath(t *testing.T) {
	input := `{
		"hook_event_name":"PreToolUse",
		"taskId":"task-1",
		"tool_input":{"file_path":"/home/user/project/main.py"}
	}`
	ctx, err := (&ClineAdapter{}).ParseHookInput(input)
	if err != nil {
		t.Fatal(err)
	}
	if ctx.FilePath != "/home/user/project/main.py" {
		t.Errorf("expected /home/user/project/main.py, got %s", ctx.FilePath)
	}
}
