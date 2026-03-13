package adapters

import (
	"encoding/json"
	"fmt"

	aierr "github.com/apshoemaker/ai-attr/pkg/errors"
)

type CodexAdapter struct{}

func (a *CodexAdapter) Name() string { return "codex" }

func (a *CodexAdapter) ParseHookInput(input string) (*AgentContext, error) {
	var data map[string]json.RawMessage
	if err := json.Unmarshal([]byte(input), &data); err != nil {
		return nil, aierr.NewParse(fmt.Sprintf("invalid JSON: %v", err))
	}

	eventType := getStr(data, "event_type", "type")
	if eventType == "" {
		eventType = "agent_turn_complete"
	}

	var phase HookPhase
	switch eventType {
	case "agent_turn_complete", "notify":
		phase = PostToolUse
	default:
		return nil, aierr.NewParse(fmt.Sprintf("unknown event_type: %s", eventType))
	}

	sessionID := getStr(data, "session_id", "thread_id")
	if sessionID == "" {
		sessionID = "unknown"
	}

	return &AgentContext{
		Tool:      "codex",
		SessionID: sessionID,
		Model:     ExtractModel(data),
		Phase:     phase,
	}, nil
}
