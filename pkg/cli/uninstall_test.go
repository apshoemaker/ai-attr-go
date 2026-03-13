package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/apshoemaker/ai-attr/pkg/installers"
)

func TestUninstallRemovesPostCommitHookLine(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".git", "hooks"), 0755)
	installers.InstallGitHook(dir, "/usr/local/bin/ai-attr")
	installers.InstallClaude(dir, "/usr/local/bin/ai-attr")

	// Add extra line
	hookPath := filepath.Join(dir, ".git", "hooks", "post-commit")
	data, _ := os.ReadFile(hookPath)
	os.WriteFile(hookPath, append(data, []byte("\necho 'other hook'\n")...), 0755)

	UninstallAt(dir)

	content, _ := os.ReadFile(hookPath)
	if strings.Contains(string(content), "ai-attr") {
		t.Error("expected ai-attr removed")
	}
	if !strings.Contains(string(content), "other hook") {
		t.Error("expected other hook preserved")
	}
}

func TestUninstallNoopWhenNotInstalled(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".git", "hooks"), 0755)
	if err := UninstallAt(dir); err != nil {
		t.Fatal(err)
	}
}

func TestUninstallCleansAiAttrDir(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".git", "hooks"), 0755)
	os.MkdirAll(filepath.Join(dir, ".git", "ai-attr", "working_logs"), 0755)
	installers.InstallGitHook(dir, "/usr/local/bin/ai-attr")

	UninstallAt(dir)

	if _, err := os.Stat(filepath.Join(dir, ".git", "ai-attr")); !os.IsNotExist(err) {
		t.Error("expected ai-attr dir to be removed")
	}
}
