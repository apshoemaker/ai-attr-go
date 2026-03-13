package adapters

import (
	"encoding/json"
	"fmt"

	aierr "github.com/apshoemaker/ai-attr/pkg/errors"
)

type ClineAdapter struct{}

func (a *ClineAdapter) Name() string { return "cline" }

func (a *ClineAdapter) ParseHookInput(input string) (*AgentContext, error) {
	var data map[string]json.RawMessage
	if err := json.Unmarshal([]byte(input), &data); err != nil {
		return nil, aierr.NewParse(fmt.Sprintf("invalid JSON: %v", err))
	}

	eventName := getStr(data, "hook_event_name", "hookEventName")
	if eventName == "" {
		return nil, aierr.NewParse("missing hook_event_name")
	}

	var phase HookPhase
	switch eventName {
	case "PreToolUse", "pre_tool_use":
		phase = PreToolUse
	case "PostToolUse", "post_tool_use":
		phase = PostToolUse
	default:
		return nil, aierr.NewParse(fmt.Sprintf("unknown hook_event_name: %s", eventName))
	}

	sessionID := getStr(data, "taskId", "task_id", "session_id")
	if sessionID == "" {
		sessionID = "unknown"
	}

	filePath := getNestedStr(data, []string{"tool_input", "toolInput"}, []string{"file_path", "filePath", "path"})

	return &AgentContext{
		Tool:      "cline",
		SessionID: sessionID,
		Model:     ExtractModel(data),
		FilePath:  filePath,
		Phase:     phase,
	}, nil
}
