package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/apshoemaker/ai-attr/pkg/core"
	"github.com/apshoemaker/ai-attr/pkg/storage"
)

func setupGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, args := range [][]string{
		{"init"},
		{"config", "user.email", "test@test.com"},
		{"config", "user.name", "Test User"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.CombinedOutput()
	}
	os.WriteFile(filepath.Join(dir, "README.md"), []byte("# Test\n"), 0644)
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = dir
	cmd.CombinedOutput()
	cmd = exec.Command("git", "commit", "-m", "initial")
	cmd.Dir = dir
	cmd.CombinedOutput()
	return dir
}

func makeHookJSON(event, filePath string) string {
	return fmt.Sprintf(`{
		"hook_event_name":"%s",
		"tool_name":"Write",
		"session_id":"test-session",
		"cwd":"/tmp",
		"tool_input":{"file_path":"%s"}
	}`, event, filePath)
}

func TestPreToolUseCreatesSnapshot(t *testing.T) {
	dir := setupGitRepo(t)
	filePath := filepath.Join(dir, "src", "hello.rs")
	os.MkdirAll(filepath.Join(dir, "src"), 0755)
	os.WriteFile(filePath, []byte("fn main() {}\n"), 0644)

	input := makeHookJSON("PreToolUse", filePath)
	if err := ProcessHookInput(input, "claude", dir); err != nil {
		t.Fatal(err)
	}

	store, _ := storage.NewRepoStorage(filepath.Join(dir, ".git"), dir)
	content, ok, _ := store.LoadSnapshot("src/hello.rs")
	if !ok {
		t.Fatal("expected snapshot to exist")
	}
	if content != "fn main() {}\n" {
		t.Errorf("unexpected content: %s", content)
	}
}

func TestPostToolUseWritesCheckpoint(t *testing.T) {
	dir := setupGitRepo(t)
	filePath := filepath.Join(dir, "src", "hello.rs")
	os.MkdirAll(filepath.Join(dir, "src"), 0755)
	os.WriteFile(filePath, []byte("fn main() {}\n"), 0644)

	// PreToolUse
	ProcessHookInput(makeHookJSON("PreToolUse", filePath), "claude", dir)

	// Simulate edit
	os.WriteFile(filePath, []byte("fn main() {\n    println!(\"hello\");\n}\n"), 0644)

	// PostToolUse
	if err := ProcessHookInput(makeHookJSON("PostToolUse", filePath), "claude", dir); err != nil {
		t.Fatal(err)
	}

	store, _ := storage.NewRepoStorage(filepath.Join(dir, ".git"), dir)
	head, _ := storage.GetHeadSHA(dir)
	logDir := filepath.Join(store.WorkingLogs, head)
	entries, _ := core.ReadEntries(logDir)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].FilePath != "src/hello.rs" {
		t.Errorf("expected src/hello.rs, got %s", entries[0].FilePath)
	}
}

func TestPostToolUseDeletesSnapshot(t *testing.T) {
	dir := setupGitRepo(t)
	filePath := filepath.Join(dir, "src", "hello.rs")
	os.MkdirAll(filepath.Join(dir, "src"), 0755)
	os.WriteFile(filePath, []byte("fn main() {}\n"), 0644)

	ProcessHookInput(makeHookJSON("PreToolUse", filePath), "claude", dir)
	ProcessHookInput(makeHookJSON("PostToolUse", filePath), "claude", dir)

	store, _ := storage.NewRepoStorage(filepath.Join(dir, ".git"), dir)
	_, ok, _ := store.LoadSnapshot("src/hello.rs")
	if ok {
		t.Error("expected snapshot to be deleted")
	}
}

func TestNewFileAllLinesAttributed(t *testing.T) {
	dir := setupGitRepo(t)
	filePath := filepath.Join(dir, "src", "new.rs")
	os.MkdirAll(filepath.Join(dir, "src"), 0755)

	// PreToolUse on non-existent file
	ProcessHookInput(makeHookJSON("PreToolUse", filePath), "claude", dir)

	// Create file
	os.WriteFile(filePath, []byte("line 1\nline 2\nline 3\n"), 0644)

	ProcessHookInput(makeHookJSON("PostToolUse", filePath), "claude", dir)

	store, _ := storage.NewRepoStorage(filepath.Join(dir, ".git"), dir)
	head, _ := storage.GetHeadSHA(dir)
	logDir := filepath.Join(store.WorkingLogs, head)
	entries, _ := core.ReadEntries(logDir)
	expected := []uint32{1, 2, 3}
	if len(entries[0].AddedLines) != len(expected) {
		t.Fatalf("expected %v, got %v", expected, entries[0].AddedLines)
	}
	for i, v := range expected {
		if entries[0].AddedLines[i] != v {
			t.Errorf("index %d: expected %d, got %d", i, v, entries[0].AddedLines[i])
		}
	}
}

func TestPostToolUseSavesPostSnapshot(t *testing.T) {
	dir := setupGitRepo(t)
	filePath := filepath.Join(dir, "src", "hello.rs")
	os.MkdirAll(filepath.Join(dir, "src"), 0755)
	os.WriteFile(filePath, []byte("fn main() {}\n"), 0644)

	ProcessHookInput(makeHookJSON("PreToolUse", filePath), "claude", dir)
	os.WriteFile(filePath, []byte("fn main() {\n    println!(\"hello\");\n}\n"), 0644)
	ProcessHookInput(makeHookJSON("PostToolUse", filePath), "claude", dir)

	store, _ := storage.NewRepoStorage(filepath.Join(dir, ".git"), dir)
	content, ok, _ := store.LoadPostSnapshot("src/hello.rs")
	if !ok {
		t.Fatal("expected post snapshot")
	}
	if content != "fn main() {\n    println!(\"hello\");\n}\n" {
		t.Errorf("unexpected content: %s", content)
	}
}
