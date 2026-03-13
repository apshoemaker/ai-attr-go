package adapters

import "encoding/json"

// HookPhase represents which phase of the hook lifecycle we're in.
type HookPhase int

const (
	SessionStart HookPhase = iota
	PreToolUse
	PostToolUse
)

func (p HookPhase) String() string {
	switch p {
	case SessionStart:
		return "SessionStart"
	case PreToolUse:
		return "PreToolUse"
	case PostToolUse:
		return "PostToolUse"
	default:
		return "Unknown"
	}
}

// AgentContext is the common agent identity extracted from hook input.
type AgentContext struct {
	Tool      string
	SessionID string
	Model     string // empty if unknown
	FilePath  string // empty if not applicable
	Phase     HookPhase
}

// AgentAdapter parses agent-specific hook payloads.
type AgentAdapter interface {
	ParseHookInput(input string) (*AgentContext, error)
	Name() string
}

// NewAdapter creates the appropriate adapter for the given agent name.
func NewAdapter(name string) AgentAdapter {
	switch name {
	case "claude":
		return &ClaudeAdapter{}
	case "copilot":
		return &CopilotAdapter{}
	case "cline":
		return &ClineAdapter{}
	case "codex":
		return &CodexAdapter{}
	default:
		return &GenericAdapter{agentName: name}
	}
}

// ExtractModel extracts model from common field names in hook JSON.
func ExtractModel(data map[string]json.RawMessage) string {
	for _, key := range []string{"model", "modelId", "model_id"} {
		if raw, ok := data[key]; ok {
			var s string
			if json.Unmarshal(raw, &s) == nil && s != "" {
				return s
			}
		}
	}
	return ""
}

func getStr(data map[string]json.RawMessage, keys ...string) string {
	for _, key := range keys {
		if raw, ok := data[key]; ok {
			var s string
			if json.Unmarshal(raw, &s) == nil {
				return s
			}
		}
	}
	return ""
}

func getNestedStr(data map[string]json.RawMessage, outerKeys []string, innerKeys []string) string {
	for _, outerKey := range outerKeys {
		if raw, ok := data[outerKey]; ok {
			var nested map[string]json.RawMessage
			if json.Unmarshal(raw, &nested) == nil {
				for _, innerKey := range innerKeys {
					if innerRaw, ok := nested[innerKey]; ok {
						var s string
						if json.Unmarshal(innerRaw, &s) == nil {
							return s
						}
					}
				}
			}
		}
	}
	return ""
}
