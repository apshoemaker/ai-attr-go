package adapters

import (
	"encoding/json"
	"fmt"
	"strings"

	aierr "github.com/apshoemaker/ai-attr/pkg/errors"
)

var editTools = []string{"Write", "Edit", "MultiEdit", "CreateFile"}

type ClaudeAdapter struct{}

func (a *ClaudeAdapter) Name() string { return "claude" }

func (a *ClaudeAdapter) ParseHookInput(input string) (*AgentContext, error) {
	var data map[string]json.RawMessage
	if err := json.Unmarshal([]byte(input), &data); err != nil {
		return nil, aierr.NewParse(fmt.Sprintf("invalid JSON: %v", err))
	}

	eventName := getStr(data, "hook_event_name")
	if eventName == "" {
		return nil, aierr.NewParse("missing hook_event_name")
	}

	var phase HookPhase
	switch eventName {
	case "SessionStart":
		phase = SessionStart
	case "PreToolUse":
		phase = PreToolUse
	case "PostToolUse":
		phase = PostToolUse
	default:
		return nil, aierr.NewParse(fmt.Sprintf("unknown hook_event_name: %s", eventName))
	}

	if phase == SessionStart {
		sessionID := extractClaudeSessionID(data)
		model := ExtractModel(data)
		return &AgentContext{
			Tool:      "claude",
			SessionID: sessionID,
			Model:     model,
			Phase:     phase,
		}, nil
	}

	toolName := getStr(data, "tool_name")
	if !isEditTool(toolName) {
		return nil, aierr.NewParse(fmt.Sprintf("tool '%s' is not an edit tool", toolName))
	}

	sessionID := extractClaudeSessionID(data)
	filePath := getNestedStr(data, []string{"tool_input"}, []string{"file_path"})

	return &AgentContext{
		Tool:      "claude",
		SessionID: sessionID,
		FilePath:  filePath,
		Phase:     phase,
	}, nil
}

func extractClaudeSessionID(data map[string]json.RawMessage) string {
	if sid := getStr(data, "session_id"); sid != "" {
		return sid
	}
	if tp := getStr(data, "transcript_path"); tp != "" {
		if sid := extractSessionFromTranscriptPath(tp); sid != "" {
			return sid
		}
	}
	return "unknown"
}

func extractSessionFromTranscriptPath(path string) string {
	lastSlash := strings.LastIndex(path, "/")
	if lastSlash == -1 {
		return ""
	}
	filename := path[lastSlash+1:]
	if !strings.HasSuffix(filename, ".jsonl") {
		return ""
	}
	return strings.TrimSuffix(filename, ".jsonl")
}

func isEditTool(name string) bool {
	for _, t := range editTools {
		if t == name {
			return true
		}
	}
	return false
}
