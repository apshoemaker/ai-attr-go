package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/apshoemaker/ai-attr/pkg/core"
	"github.com/apshoemaker/ai-attr/pkg/storage"
)

func makeCommitWithFile(t *testing.T, dir, file, content string) {
	t.Helper()
	filePath := filepath.Join(dir, file)
	os.MkdirAll(filepath.Dir(filePath), 0755)
	os.WriteFile(filePath, []byte(content), 0644)
	cmd := exec.Command("git", "add", file)
	cmd.Dir = dir
	cmd.CombinedOutput()
	cmd = exec.Command("git", "commit", "-m", "update")
	cmd.Dir = dir
	cmd.CombinedOutput()
}

func makeEntry(session, tool, file string, lines []uint32) *core.CheckpointEntry {
	model := "sonnet-4"
	return &core.CheckpointEntry{
		SessionID:      session,
		Tool:           tool,
		Model:          &model,
		FilePath:       file,
		AddedLines:     lines,
		TotalAdditions: uint32(len(lines)),
		TotalDeletions: 0,
		Timestamp:      1700000000,
	}
}

func writeTestEntry(t *testing.T, store *storage.RepoStorage, headSHA string, entry *core.CheckpointEntry) {
	t.Helper()
	logDir, err := store.WorkingLogDir(headSHA)
	if err != nil {
		t.Fatal(err)
	}
	if err := core.WriteEntry(logDir, entry); err != nil {
		t.Fatal(err)
	}
}

func TestCommitWritesGitNote(t *testing.T) {
	dir := setupGitRepo(t)
	baseHead, _ := storage.GetHeadSHA(dir)

	makeCommitWithFile(t, dir, "src/main.rs", "line1\nline2\nline3\n")
	head, _ := storage.GetHeadSHA(dir)

	store, _ := storage.NewRepoStorage(filepath.Join(dir, ".git"), dir)
	writeTestEntry(t, store, baseHead, makeEntry("sess-1", "claude", "src/main.rs", []uint32{1, 2, 3}))
	store.SavePostSnapshot("src/main.rs", "line1\nline2\nline3\n")

	if err := CommitAt(dir); err != nil {
		t.Fatal(err)
	}

	_, ok, err := storage.NotesShow(dir, head)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected note to exist")
	}
}

func TestCommitNoteHasCorrectFormat(t *testing.T) {
	dir := setupGitRepo(t)
	baseHead, _ := storage.GetHeadSHA(dir)

	makeCommitWithFile(t, dir, "src/main.rs", "line1\nline2\nline3\n")
	head, _ := storage.GetHeadSHA(dir)

	store, _ := storage.NewRepoStorage(filepath.Join(dir, ".git"), dir)
	writeTestEntry(t, store, baseHead, makeEntry("sess-1", "claude", "src/main.rs", []uint32{1, 2, 3}))
	store.SavePostSnapshot("src/main.rs", "line1\nline2\nline3\n")

	CommitAt(dir)

	content, _, _ := storage.NotesShow(dir, head)
	log, err := core.DeserializeFromString(content)
	if err != nil {
		t.Fatal(err)
	}
	if log.Metadata.SchemaVersion != core.AuthorshipLogVersion {
		t.Errorf("expected %s, got %s", core.AuthorshipLogVersion, log.Metadata.SchemaVersion)
	}
	if len(log.Attestations) != 1 {
		t.Fatalf("expected 1 attestation, got %d", len(log.Attestations))
	}
	if log.Attestations[0].FilePath != "src/main.rs" {
		t.Errorf("expected src/main.rs, got %s", log.Attestations[0].FilePath)
	}
}

func TestCommitTwoPhaseAIOnly(t *testing.T) {
	dir := setupGitRepo(t)
	baseHead, _ := storage.GetHeadSHA(dir)

	makeCommitWithFile(t, dir, "src/main.rs", "aaa\nbbb\nccc\n")
	head, _ := storage.GetHeadSHA(dir)

	store, _ := storage.NewRepoStorage(filepath.Join(dir, ".git"), dir)
	writeTestEntry(t, store, baseHead, makeEntry("sess-1", "claude", "src/main.rs", []uint32{1, 2, 3}))
	store.SavePostSnapshot("src/main.rs", "aaa\nbbb\nccc\n")

	CommitAt(dir)

	content, _, _ := storage.NotesShow(dir, head)
	log, _ := core.DeserializeFromString(content)

	if _, ok := log.Metadata.Prompts[core.HumanSentinel]; ok {
		t.Error("expected no human sentinel")
	}
	fa := log.Attestations[0]
	if len(fa.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(fa.Entries))
	}
	var aiLines []uint32
	for _, r := range fa.Entries[0].LineRanges {
		aiLines = append(aiLines, r.Expand()...)
	}
	if len(aiLines) != 3 || aiLines[0] != 1 || aiLines[1] != 2 || aiLines[2] != 3 {
		t.Errorf("expected [1,2,3], got %v", aiLines)
	}
}

func TestCommitTwoPhaseHumanEditsAfterAI(t *testing.T) {
	dir := setupGitRepo(t)
	baseHead, _ := storage.GetHeadSHA(dir)

	makeCommitWithFile(t, dir, "src/main.rs", "aaa\nbbb\nccc\nhuman line\n")
	head, _ := storage.GetHeadSHA(dir)

	store, _ := storage.NewRepoStorage(filepath.Join(dir, ".git"), dir)
	writeTestEntry(t, store, baseHead, makeEntry("sess-1", "claude", "src/main.rs", []uint32{1, 2, 3}))
	store.SavePostSnapshot("src/main.rs", "aaa\nbbb\nccc\n")

	CommitAt(dir)

	content, _, _ := storage.NotesShow(dir, head)
	log, _ := core.DeserializeFromString(content)

	if _, ok := log.Metadata.Prompts[core.HumanSentinel]; !ok {
		t.Fatal("expected human sentinel")
	}

	fa := log.Attestations[0]
	if len(fa.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(fa.Entries))
	}

	var aiEntry, humanEntry *core.AttestationEntry
	for i := range fa.Entries {
		if fa.Entries[i].Hash == core.HumanSentinel {
			humanEntry = &fa.Entries[i]
		} else {
			aiEntry = &fa.Entries[i]
		}
	}

	var aiLines, humanLines []uint32
	for _, r := range aiEntry.LineRanges {
		aiLines = append(aiLines, r.Expand()...)
	}
	for _, r := range humanEntry.LineRanges {
		humanLines = append(humanLines, r.Expand()...)
	}

	if len(aiLines) != 3 || aiLines[0] != 1 || aiLines[1] != 2 || aiLines[2] != 3 {
		t.Errorf("expected AI [1,2,3], got %v", aiLines)
	}
	if len(humanLines) != 1 || humanLines[0] != 4 {
		t.Errorf("expected human [4], got %v", humanLines)
	}
}

func TestCommitTwoPhaseHumanModifiesAILine(t *testing.T) {
	dir := setupGitRepo(t)
	baseHead, _ := storage.GetHeadSHA(dir)

	makeCommitWithFile(t, dir, "src/main.rs", "aaa\nBBB\nccc\n")
	head, _ := storage.GetHeadSHA(dir)

	store, _ := storage.NewRepoStorage(filepath.Join(dir, ".git"), dir)
	writeTestEntry(t, store, baseHead, makeEntry("sess-1", "claude", "src/main.rs", []uint32{1, 2, 3}))
	store.SavePostSnapshot("src/main.rs", "aaa\nbbb\nccc\n")

	CommitAt(dir)

	content, _, _ := storage.NotesShow(dir, head)
	log, _ := core.DeserializeFromString(content)

	fa := log.Attestations[0]
	var aiEntry, humanEntry *core.AttestationEntry
	for i := range fa.Entries {
		if fa.Entries[i].Hash == core.HumanSentinel {
			humanEntry = &fa.Entries[i]
		} else {
			aiEntry = &fa.Entries[i]
		}
	}

	var aiLines, humanLines []uint32
	for _, r := range aiEntry.LineRanges {
		aiLines = append(aiLines, r.Expand()...)
	}
	for _, r := range humanEntry.LineRanges {
		humanLines = append(humanLines, r.Expand()...)
	}

	if len(aiLines) != 2 || aiLines[0] != 1 || aiLines[1] != 3 {
		t.Errorf("expected AI [1,3], got %v", aiLines)
	}
	if len(humanLines) != 1 || humanLines[0] != 2 {
		t.Errorf("expected human [2], got %v", humanLines)
	}
}

func TestCommitNoEntriesIsNoop(t *testing.T) {
	dir := setupGitRepo(t)
	head, _ := storage.GetHeadSHA(dir)

	CommitAt(dir)

	_, ok, _ := storage.NotesShow(dir, head)
	if ok {
		t.Error("expected no note")
	}
}

func TestCommitCleansUpWorkingLogs(t *testing.T) {
	dir := setupGitRepo(t)
	baseHead, _ := storage.GetHeadSHA(dir)

	store, _ := storage.NewRepoStorage(filepath.Join(dir, ".git"), dir)
	writeTestEntry(t, store, baseHead, makeEntry("sess-1", "claude", "src/a.rs", []uint32{1}))
	if !store.HasWorkingLog(baseHead) {
		t.Fatal("expected working log")
	}

	makeCommitWithFile(t, dir, "src/a.rs", "line1\n")
	CommitAt(dir)

	if store.HasWorkingLog(baseHead) {
		t.Error("expected working log to be cleaned up")
	}
}
