package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInstallClaudeCreatesSettingsAndHook(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".git", "hooks"), 0755)

	if err := InstallAt(dir, "/usr/local/bin/ai-attr", []string{"claude"}); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dir, ".git", "hooks", "post-commit")); os.IsNotExist(err) {
		t.Error("expected post-commit hook")
	}
	if _, err := os.Stat(filepath.Join(dir, ".claude", "settings.json")); os.IsNotExist(err) {
		t.Error("expected claude settings")
	}
}

func TestInstallUnknownAgentErrors(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".git", "hooks"), 0755)

	err := InstallAt(dir, "/usr/local/bin/ai-attr", []string{"unknown"})
	if err == nil {
		t.Error("expected error for unknown agent")
	}
}
