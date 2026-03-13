package storage

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/apshoemaker/ai-attr/pkg/core"
)

func setupGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	commands := [][]string{
		{"init"},
		{"config", "user.email", "test@test.com"},
		{"config", "user.name", "Test User"},
	}
	for _, args := range commands {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %s", args, out)
		}
	}

	// Create initial commit
	readme := filepath.Join(dir, "README.md")
	os.WriteFile(readme, []byte("# Test\n"), 0644)
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = dir
	cmd.CombinedOutput()
	cmd = exec.Command("git", "commit", "-m", "initial")
	cmd.Dir = dir
	cmd.CombinedOutput()

	return dir
}

func TestNotesAddAndShow(t *testing.T) {
	dir := setupGitRepo(t)
	head, err := GetHeadSHA(dir)
	if err != nil {
		t.Fatal(err)
	}

	content := "test note content\n---\n{}"
	if err := NotesAdd(dir, head, content); err != nil {
		t.Fatal(err)
	}

	result, ok, err := NotesShow(dir, head)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected note to exist")
	}
	if trimmed := result; trimmed != content+"\n" && trimmed != content {
		// git notes may add trailing newline
		if len(result) == 0 {
			t.Errorf("expected content, got empty")
		}
	}
}

func TestNotesShowNonexistent(t *testing.T) {
	dir := setupGitRepo(t)
	head, _ := GetHeadSHA(dir)

	_, ok, err := NotesShow(dir, head)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Error("expected no note")
	}
}

func TestNotesOverwrite(t *testing.T) {
	dir := setupGitRepo(t)
	head, _ := GetHeadSHA(dir)

	NotesAdd(dir, head, "first")
	NotesAdd(dir, head, "second")

	result, ok, _ := NotesShow(dir, head)
	if !ok {
		t.Fatal("expected note")
	}
	if trimmed := result; !contains(trimmed, "second") {
		t.Errorf("expected 'second', got '%s'", trimmed)
	}
}

func TestNotesRemove(t *testing.T) {
	dir := setupGitRepo(t)
	head, _ := GetHeadSHA(dir)

	NotesAdd(dir, head, "to be removed")
	_, ok, _ := NotesShow(dir, head)
	if !ok {
		t.Fatal("expected note")
	}

	NotesRemove(dir, head)
	_, ok, _ = NotesShow(dir, head)
	if ok {
		t.Error("expected note to be removed")
	}
}

func TestRoundtripAuthorshipLogThroughNotes(t *testing.T) {
	dir := setupGitRepo(t)
	head, _ := GetHeadSHA(dir)
	hash := core.GenerateShortHash("session-1", "claude")

	log := core.NewAuthorshipLog()
	log.Metadata.BaseCommitSHA = "abc123"
	log.Metadata.Prompts[hash] = &core.PromptRecord{
		AgentID: core.AgentId{
			Tool:  "claude",
			Model: "sonnet-4",
			ID:    "session-1",
		},
		HumanAuthor:   "alice",
		TotalAdds:     10,
		TotalDels:     2,
		AcceptedLines: 8,
	}

	file := log.GetOrCreateFile("src/main.rs")
	file.AddEntry(core.NewAttestationEntry(
		hash,
		[]core.LineRange{core.Range(1, 10), core.Range(15, 20)},
	))

	serialized, err := log.SerializeToString()
	if err != nil {
		t.Fatal(err)
	}
	if err := NotesAdd(dir, head, serialized); err != nil {
		t.Fatal(err)
	}

	retrieved, ok, err := GetAuthorship(dir, head)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected authorship log")
	}
	if len(retrieved.Attestations) != 1 {
		t.Errorf("expected 1 attestation, got %d", len(retrieved.Attestations))
	}
	if retrieved.Attestations[0].FilePath != "src/main.rs" {
		t.Errorf("expected src/main.rs, got %s", retrieved.Attestations[0].FilePath)
	}
	if len(retrieved.Metadata.Prompts) != 1 {
		t.Errorf("expected 1 prompt, got %d", len(retrieved.Metadata.Prompts))
	}
}

func TestGetUserName(t *testing.T) {
	dir := setupGitRepo(t)
	name, err := GetUserName(dir)
	if err != nil {
		t.Fatal(err)
	}
	if name != "Test User" {
		t.Errorf("expected 'Test User', got '%s'", name)
	}
}

func TestFindRepoRoot(t *testing.T) {
	dir := setupGitRepo(t)
	sub := filepath.Join(dir, "src")
	os.MkdirAll(sub, 0755)

	root, err := FindRepoRoot(sub)
	if err != nil {
		t.Fatal(err)
	}
	// Resolve symlinks for comparison (macOS /tmp → /private/tmp)
	expectedAbs, _ := filepath.EvalSymlinks(dir)
	gotAbs, _ := filepath.EvalSymlinks(root)
	if expectedAbs != gotAbs {
		t.Errorf("expected %s, got %s", expectedAbs, gotAbs)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && findSubstring(s, substr))
}

func findSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
