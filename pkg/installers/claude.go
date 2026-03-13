package installers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const toolMatcher = "Write|Edit|MultiEdit|CreateFile"

// InstallClaude installs Claude Code hooks in .claude/settings.json.
func InstallClaude(repoWorkdir, binaryPath string) error {
	claudeDir := filepath.Join(repoWorkdir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		return err
	}

	settingsPath := filepath.Join(claudeDir, "settings.json")
	var settings map[string]interface{}

	if data, err := os.ReadFile(settingsPath); err == nil {
		json.Unmarshal(data, &settings)
	}
	if settings == nil {
		settings = make(map[string]interface{})
	}

	command := fmt.Sprintf("%s checkpoint claude --hook-input", binaryPath)

	toolHookEntry := map[string]interface{}{
		"matcher": toolMatcher,
		"hooks": []interface{}{
			map[string]interface{}{
				"type":    "command",
				"command": command,
			},
		},
	}

	sessionHookEntry := map[string]interface{}{
		"matcher": "",
		"hooks": []interface{}{
			map[string]interface{}{
				"type":    "command",
				"command": command,
			},
		},
	}

	hooks, _ := settings["hooks"].(map[string]interface{})
	if hooks == nil {
		hooks = make(map[string]interface{})
		settings["hooks"] = hooks
	}

	// Install SessionStart
	installHookPhase(hooks, "SessionStart", sessionHookEntry)

	// Install PreToolUse and PostToolUse
	for _, phase := range []string{"PreToolUse", "PostToolUse"} {
		installHookPhase(hooks, phase, toolHookEntry)
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(settingsPath, data, 0644)
}

func installHookPhase(hooks map[string]interface{}, phase string, entry map[string]interface{}) {
	arr, _ := hooks[phase].([]interface{})
	if arr == nil {
		arr = []interface{}{}
	}

	// Remove existing ai-attr entries
	var filtered []interface{}
	for _, item := range arr {
		if !isAiAttrHook(item) {
			filtered = append(filtered, item)
		}
	}
	filtered = append(filtered, entry)
	hooks[phase] = filtered
}

// UninstallClaude removes ai-attr hooks from Claude settings.
func UninstallClaude(repoWorkdir string) error {
	settingsPath := filepath.Join(repoWorkdir, ".claude", "settings.json")
	data, err := os.ReadFile(settingsPath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		return err
	}

	hooks, _ := settings["hooks"].(map[string]interface{})
	if hooks == nil {
		return nil
	}

	for _, phase := range []string{"SessionStart", "PreToolUse", "PostToolUse"} {
		arr, _ := hooks[phase].([]interface{})
		if arr == nil {
			continue
		}
		var filtered []interface{}
		for _, item := range arr {
			if !isAiAttrHook(item) {
				filtered = append(filtered, item)
			}
		}
		hooks[phase] = filtered
	}

	out, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(settingsPath, out, 0644)
}

func isAiAttrHook(entry interface{}) bool {
	m, ok := entry.(map[string]interface{})
	if !ok {
		return false
	}
	hooksArr, _ := m["hooks"].([]interface{})
	for _, h := range hooksArr {
		hMap, _ := h.(map[string]interface{})
		if cmd, _ := hMap["command"].(string); strings.Contains(cmd, "ai-attr") {
			return true
		}
	}
	return false
}
