package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func sampleEntry(session, file string) *CheckpointEntry {
	model := "sonnet-4"
	return &CheckpointEntry{
		SessionID:      session,
		Tool:           "claude",
		Model:          &model,
		FilePath:       file,
		AddedLines:     []uint32{1, 2, 3},
		TotalAdditions: 3,
		TotalDeletions: 0,
		Timestamp:      1700000000,
	}
}

func TestCheckpointEntryRoundtrip(t *testing.T) {
	entry := sampleEntry("sess-1", "src/main.rs")
	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatal(err)
	}
	var deserialized CheckpointEntry
	if err := json.Unmarshal(data, &deserialized); err != nil {
		t.Fatal(err)
	}
	if deserialized.SessionID != entry.SessionID {
		t.Errorf("session_id mismatch: %s != %s", deserialized.SessionID, entry.SessionID)
	}
	if deserialized.FilePath != entry.FilePath {
		t.Errorf("file_path mismatch: %s != %s", deserialized.FilePath, entry.FilePath)
	}
	if *deserialized.Model != *entry.Model {
		t.Errorf("model mismatch: %s != %s", *deserialized.Model, *entry.Model)
	}
}

func TestWriteAndReadEntries(t *testing.T) {
	logDir := filepath.Join(t.TempDir(), "log")

	e1 := sampleEntry("sess-1", "a.rs")
	e2 := sampleEntry("sess-1", "b.rs")
	e3 := sampleEntry("sess-2", "c.rs")

	if err := WriteEntry(logDir, e1); err != nil {
		t.Fatal(err)
	}
	if err := WriteEntry(logDir, e2); err != nil {
		t.Fatal(err)
	}
	if err := WriteEntry(logDir, e3); err != nil {
		t.Fatal(err)
	}

	entries, err := ReadEntries(logDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
	if entries[0].FilePath != "a.rs" {
		t.Errorf("expected a.rs, got %s", entries[0].FilePath)
	}
	if entries[1].FilePath != "b.rs" {
		t.Errorf("expected b.rs, got %s", entries[1].FilePath)
	}
	if entries[2].FilePath != "c.rs" {
		t.Errorf("expected c.rs, got %s", entries[2].FilePath)
	}
}

func TestReadEmptyDir(t *testing.T) {
	logDir := filepath.Join(t.TempDir(), "nonexistent")
	entries, err := ReadEntries(logDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("expected empty, got %d entries", len(entries))
	}
}

func TestAutoIncrementIndex(t *testing.T) {
	logDir := filepath.Join(t.TempDir(), "log")

	if err := WriteEntry(logDir, sampleEntry("s", "a.rs")); err != nil {
		t.Fatal(err)
	}
	if err := WriteEntry(logDir, sampleEntry("s", "b.rs")); err != nil {
		t.Fatal(err)
	}
	if err := WriteEntry(logDir, sampleEntry("s", "c.rs")); err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"entry-0.json", "entry-1.json", "entry-2.json"} {
		if _, err := os.Stat(filepath.Join(logDir, name)); os.IsNotExist(err) {
			t.Errorf("expected %s to exist", name)
		}
	}
}
