package installers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCodexInstallCreatesToml(t *testing.T) {
	dir := t.TempDir()
	codexDir := filepath.Join(dir, ".codex")

	if err := InstallCodex("/usr/local/bin/ai-attr", codexDir); err != nil {
		t.Fatal(err)
	}

	configPath := filepath.Join(codexDir, "config.toml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	if !strings.Contains(content, "notify") {
		t.Error("expected notify line")
	}
	if !strings.Contains(content, "ai-attr") {
		t.Error("expected ai-attr in content")
	}
	if !strings.Contains(content, "checkpoint codex") {
		t.Error("expected checkpoint codex in content")
	}
}
