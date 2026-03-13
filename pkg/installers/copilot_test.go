package installers

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCopilotInstallCreatesConfig(t *testing.T) {
	dir := t.TempDir()
	if err := InstallCopilot(dir, "/usr/local/bin/ai-attr"); err != nil {
		t.Fatal(err)
	}

	configPath := filepath.Join(dir, ".github", "hooks", "ai-attr.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}

	var config map[string]interface{}
	json.Unmarshal(data, &config)
	if config["agent"] != "copilot" {
		t.Errorf("expected copilot, got %v", config["agent"])
	}
	hooks := config["hooks"].(map[string]interface{})
	pre := hooks["PreToolUse"].(map[string]interface{})
	cmd := pre["command"].(string)
	if !strings.Contains(cmd, "ai-attr") {
		t.Errorf("expected ai-attr in command, got %s", cmd)
	}
}
