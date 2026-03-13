package installers

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestClaudeInstallFreshSettings(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".claude"), 0755)

	if err := InstallClaude(dir, "/usr/local/bin/ai-attr"); err != nil {
		t.Fatal(err)
	}

	settingsPath := filepath.Join(dir, ".claude", "settings.json")
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatal(err)
	}

	var settings map[string]interface{}
	json.Unmarshal(data, &settings)

	hooks := settings["hooks"].(map[string]interface{})
	pre := hooks["PreToolUse"].([]interface{})
	post := hooks["PostToolUse"].([]interface{})
	if len(pre) != 1 {
		t.Errorf("expected 1 PreToolUse hook, got %d", len(pre))
	}
	if len(post) != 1 {
		t.Errorf("expected 1 PostToolUse hook, got %d", len(post))
	}
}

func TestClaudeInstallPreservesOtherHooks(t *testing.T) {
	dir := t.TempDir()
	settingsPath := filepath.Join(dir, ".claude", "settings.json")
	os.MkdirAll(filepath.Join(dir, ".claude"), 0755)

	existing := map[string]interface{}{
		"hooks": map[string]interface{}{
			"PreToolUse": []interface{}{
				map[string]interface{}{
					"matcher": "Bash",
					"hooks": []interface{}{
						map[string]interface{}{"type": "command", "command": "echo other"},
					},
				},
			},
		},
	}
	data, _ := json.MarshalIndent(existing, "", "  ")
	os.WriteFile(settingsPath, data, 0644)

	if err := InstallClaude(dir, "/usr/local/bin/ai-attr"); err != nil {
		t.Fatal(err)
	}

	data, _ = os.ReadFile(settingsPath)
	var settings map[string]interface{}
	json.Unmarshal(data, &settings)
	hooks := settings["hooks"].(map[string]interface{})
	pre := hooks["PreToolUse"].([]interface{})
	if len(pre) != 2 {
		t.Errorf("expected 2 PreToolUse hooks, got %d", len(pre))
	}
}

func TestClaudeInstallDeduplicates(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".claude"), 0755)

	InstallClaude(dir, "/usr/local/bin/ai-attr")
	InstallClaude(dir, "/usr/local/bin/ai-attr")

	data, _ := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))
	var settings map[string]interface{}
	json.Unmarshal(data, &settings)
	hooks := settings["hooks"].(map[string]interface{})
	pre := hooks["PreToolUse"].([]interface{})
	if len(pre) != 1 {
		t.Errorf("expected 1 PreToolUse hook after dedup, got %d", len(pre))
	}
}

func TestClaudeSettingsHasCorrectMatcher(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".claude"), 0755)

	InstallClaude(dir, "/usr/local/bin/ai-attr")

	data, _ := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))
	var settings map[string]interface{}
	json.Unmarshal(data, &settings)
	hooks := settings["hooks"].(map[string]interface{})
	pre := hooks["PreToolUse"].([]interface{})
	entry := pre[0].(map[string]interface{})
	matcher := entry["matcher"].(string)
	if matcher != "Write|Edit|MultiEdit|CreateFile" {
		t.Errorf("expected Write|Edit|MultiEdit|CreateFile, got %s", matcher)
	}
}
