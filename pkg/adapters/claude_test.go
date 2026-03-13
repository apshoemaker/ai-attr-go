package adapters

import "testing"

func makeClaudeHookJSON(event, tool string, filePath string) string {
	fileInput := ""
	if filePath != "" {
		fileInput = `,"file_path":"` + filePath + `"`
	}
	return `{
		"hook_event_name":"` + event + `",
		"tool_name":"` + tool + `",
		"session_id":"sess-abc-123",
		"transcript_path":"/Users/test/.claude/projects/proj/4fa1c610-eb62-4fd2-b397-8aaee4d711f2.jsonl",
		"cwd":"/tmp/repo",
		"tool_input":{"command":"write"` + fileInput + `}
	}`
}

func TestClaudePreToolUse(t *testing.T) {
	input := makeClaudeHookJSON("PreToolUse", "Write", "/tmp/repo/src/main.rs")
	ctx, err := (&ClaudeAdapter{}).ParseHookInput(input)
	if err != nil {
		t.Fatal(err)
	}
	if ctx.Phase != PreToolUse {
		t.Errorf("expected PreToolUse, got %v", ctx.Phase)
	}
	if ctx.Tool != "claude" {
		t.Errorf("expected claude, got %s", ctx.Tool)
	}
}

func TestClaudePostToolUse(t *testing.T) {
	input := makeClaudeHookJSON("PostToolUse", "Edit", "/tmp/repo/src/lib.rs")
	ctx, err := (&ClaudeAdapter{}).ParseHookInput(input)
	if err != nil {
		t.Fatal(err)
	}
	if ctx.Phase != PostToolUse {
		t.Errorf("expected PostToolUse, got %v", ctx.Phase)
	}
}

func TestClaudeExtractsSessionID(t *testing.T) {
	input := makeClaudeHookJSON("PreToolUse", "Write", "")
	ctx, err := (&ClaudeAdapter{}).ParseHookInput(input)
	if err != nil {
		t.Fatal(err)
	}
	if ctx.SessionID != "sess-abc-123" {
		t.Errorf("expected sess-abc-123, got %s", ctx.SessionID)
	}
}

func TestClaudeExtractsSessionIDFromTranscript(t *testing.T) {
	input := `{
		"hook_event_name":"PreToolUse",
		"tool_name":"Write",
		"transcript_path":"/Users/test/.claude/projects/proj/4fa1c610-eb62.jsonl",
		"cwd":"/tmp/repo",
		"tool_input":{}
	}`
	ctx, err := (&ClaudeAdapter{}).ParseHookInput(input)
	if err != nil {
		t.Fatal(err)
	}
	if ctx.SessionID != "4fa1c610-eb62" {
		t.Errorf("expected 4fa1c610-eb62, got %s", ctx.SessionID)
	}
}

func TestClaudeExtractsFilePath(t *testing.T) {
	input := makeClaudeHookJSON("PostToolUse", "Write", "/tmp/repo/src/main.rs")
	ctx, err := (&ClaudeAdapter{}).ParseHookInput(input)
	if err != nil {
		t.Fatal(err)
	}
	if ctx.FilePath != "/tmp/repo/src/main.rs" {
		t.Errorf("expected /tmp/repo/src/main.rs, got %s", ctx.FilePath)
	}
}

func TestClaudeRejectsMalformedJSON(t *testing.T) {
	_, err := (&ClaudeAdapter{}).ParseHookInput("not json")
	if err == nil {
		t.Error("expected error for malformed JSON")
	}
}

func TestClaudeRejectsNonEditTool(t *testing.T) {
	input := makeClaudeHookJSON("PostToolUse", "Read", "/tmp/file")
	_, err := (&ClaudeAdapter{}).ParseHookInput(input)
	if err == nil {
		t.Error("expected error for non-edit tool")
	}
}

func TestClaudeSessionStart(t *testing.T) {
	input := `{
		"hook_event_name":"SessionStart",
		"session_id":"sess-xyz",
		"model":"claude-sonnet-4-20250514"
	}`
	ctx, err := (&ClaudeAdapter{}).ParseHookInput(input)
	if err != nil {
		t.Fatal(err)
	}
	if ctx.Phase != SessionStart {
		t.Errorf("expected SessionStart, got %v", ctx.Phase)
	}
	if ctx.SessionID != "sess-xyz" {
		t.Errorf("expected sess-xyz, got %s", ctx.SessionID)
	}
	if ctx.Model != "claude-sonnet-4-20250514" {
		t.Errorf("expected claude-sonnet-4-20250514, got %s", ctx.Model)
	}
	if ctx.FilePath != "" {
		t.Errorf("expected empty file_path, got %s", ctx.FilePath)
	}
}
