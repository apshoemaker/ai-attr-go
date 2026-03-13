package installers

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestClineInstallCreatesConfig(t *testing.T) {
	dir := t.TempDir()
	if err := InstallCline(dir, "/usr/local/bin/ai-attr"); err != nil {
		t.Fatal(err)
	}

	configPath := filepath.Join(dir, ".clinerules", "hooks", "ai-attr.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}

	var config map[string]interface{}
	json.Unmarshal(data, &config)
	if config["agent"] != "cline" {
		t.Errorf("expected cline, got %v", config["agent"])
	}
	hooks := config["hooks"].(map[string]interface{})
	post := hooks["PostToolUse"].(map[string]interface{})
	cmd := post["command"].(string)
	if !strings.Contains(cmd, "ai-attr") {
		t.Errorf("expected ai-attr in command, got %s", cmd)
	}
}
