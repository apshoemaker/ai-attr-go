package adapters

import (
	"encoding/json"
	"fmt"

	aierr "github.com/apshoemaker/ai-attr/pkg/errors"
)

var copilotEditTools = []string{"insert_edit_into_file", "create_file"}

type CopilotAdapter struct{}

func (a *CopilotAdapter) Name() string { return "copilot" }

func (a *CopilotAdapter) ParseHookInput(input string) (*AgentContext, error) {
	var data map[string]json.RawMessage
	if err := json.Unmarshal([]byte(input), &data); err != nil {
		return nil, aierr.NewParse(fmt.Sprintf("invalid JSON: %v", err))
	}

	eventName := getStr(data, "hookEventName", "hook_event_name")
	if eventName == "" {
		return nil, aierr.NewParse("missing hookEventName")
	}

	var phase HookPhase
	switch eventName {
	case "PreToolUse", "before_edit":
		phase = PreToolUse
	case "PostToolUse", "after_edit":
		phase = PostToolUse
	default:
		return nil, aierr.NewParse(fmt.Sprintf("unknown hookEventName: %s", eventName))
	}

	toolName := getStr(data, "toolName", "tool_name")
	if !isCopilotEditTool(toolName) {
		return nil, aierr.NewParse(fmt.Sprintf("tool '%s' is not an edit tool", toolName))
	}

	sessionID := getStr(data, "chatSessionId", "session_id", "sessionId")
	if sessionID == "" {
		sessionID = "unknown"
	}

	filePath := getNestedStr(data, []string{"toolInput", "tool_input"}, []string{"filePath", "file_path", "uri"})

	return &AgentContext{
		Tool:      "copilot",
		SessionID: sessionID,
		Model:     ExtractModel(data),
		FilePath:  filePath,
		Phase:     phase,
	}, nil
}

func isCopilotEditTool(name string) bool {
	for _, t := range copilotEditTools {
		if t == name {
			return true
		}
	}
	return false
}
