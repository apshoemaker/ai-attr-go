package installers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".git", "hooks"), 0755)
	return dir
}

func TestInstallFreshRepo(t *testing.T) {
	dir := setupRepo(t)
	if err := InstallGitHook(dir, "/usr/local/bin/ai-attr"); err != nil {
		t.Fatal(err)
	}

	hookPath := filepath.Join(dir, ".git", "hooks", "post-commit")
	data, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "#!/bin/sh") {
		t.Error("expected shebang")
	}
	if !strings.Contains(content, "ai-attr commit") {
		t.Error("expected ai-attr commit")
	}
}

func TestInstallPreservesExistingHook(t *testing.T) {
	dir := setupRepo(t)
	hookPath := filepath.Join(dir, ".git", "hooks", "post-commit")
	os.WriteFile(hookPath, []byte("#!/bin/sh\necho 'existing hook'\n"), 0755)

	if err := InstallGitHook(dir, "/usr/local/bin/ai-attr"); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(hookPath)
	content := string(data)
	if !strings.Contains(content, "echo 'existing hook'") {
		t.Error("expected existing hook preserved")
	}
	if !strings.Contains(content, "ai-attr commit") {
		t.Error("expected ai-attr commit added")
	}
}

func TestInstallIdempotent(t *testing.T) {
	dir := setupRepo(t)
	InstallGitHook(dir, "/usr/local/bin/ai-attr")
	InstallGitHook(dir, "/usr/local/bin/ai-attr")

	data, _ := os.ReadFile(filepath.Join(dir, ".git", "hooks", "post-commit"))
	count := strings.Count(string(data), aiAttrMarker)
	if count != 1 {
		t.Errorf("expected 1 marker, got %d", count)
	}
}

func TestHookIsExecutable(t *testing.T) {
	dir := setupRepo(t)
	InstallGitHook(dir, "/usr/local/bin/ai-attr")

	hookPath := filepath.Join(dir, ".git", "hooks", "post-commit")
	info, err := os.Stat(hookPath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&0111 == 0 {
		t.Error("hook should be executable")
	}
}
