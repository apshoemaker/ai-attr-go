package adapters

import (
	"encoding/json"
	"fmt"

	aierr "github.com/apshoemaker/ai-attr/pkg/errors"
)

type GenericAdapter struct {
	agentName string
}

func NewGenericAdapter(name string) *GenericAdapter {
	return &GenericAdapter{agentName: name}
}

func (a *GenericAdapter) Name() string { return a.agentName }

func (a *GenericAdapter) ParseHookInput(input string) (*AgentContext, error) {
	var data map[string]json.RawMessage
	if err := json.Unmarshal([]byte(input), &data); err != nil {
		return nil, aierr.NewParse(fmt.Sprintf("invalid JSON: %v", err))
	}

	phaseStr := getStr(data, "phase", "hook_event_name")
	if phaseStr == "" {
		return nil, aierr.NewParse("missing required field: phase")
	}

	var phase HookPhase
	switch phaseStr {
	case "PreToolUse", "pre":
		phase = PreToolUse
	case "PostToolUse", "post":
		phase = PostToolUse
	default:
		return nil, aierr.NewParse(fmt.Sprintf("unknown phase: %s", phaseStr))
	}

	sessionID := getStr(data, "session_id")
	if sessionID == "" {
		return nil, aierr.NewParse("missing required field: session_id")
	}

	filePath := getStr(data, "file_path")

	return &AgentContext{
		Tool:      a.agentName,
		SessionID: sessionID,
		Model:     ExtractModel(data),
		FilePath:  filePath,
		Phase:     phase,
	}, nil
}
