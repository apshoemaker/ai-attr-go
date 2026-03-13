package core

import (
	"strings"
	"testing"
)

func TestGenerateShortHash(t *testing.T) {
	hash := GenerateShortHash("session-123", "claude")
	if len(hash) != 16 {
		t.Errorf("expected 16 chars, got %d", len(hash))
	}
	// Deterministic
	if hash != GenerateShortHash("session-123", "claude") {
		t.Error("hash should be deterministic")
	}
	// Different inputs produce different hashes
	if hash == GenerateShortHash("session-456", "claude") {
		t.Error("different session_id should produce different hash")
	}
	if hash == GenerateShortHash("session-123", "copilot") {
		t.Error("different tool should produce different hash")
	}
}

func TestFormatLineRangesSerialization(t *testing.T) {
	ranges := []LineRange{Range(19, 222), SingleLine(1), SingleLine(2)}
	result := FormatLineRanges(ranges)
	if result != "1,2,19-222" {
		t.Errorf("expected '1,2,19-222', got '%s'", result)
	}
}

func TestParseLineRangesSerialization(t *testing.T) {
	ranges, err := ParseLineRanges("1,2,19-222")
	if err != nil {
		t.Fatal(err)
	}
	if len(ranges) != 3 {
		t.Fatalf("expected 3 ranges, got %d", len(ranges))
	}
	if ranges[0] != SingleLine(1) {
		t.Errorf("expected Single(1), got %v", ranges[0])
	}
	if ranges[1] != SingleLine(2) {
		t.Errorf("expected Single(2), got %v", ranges[1])
	}
	if ranges[2] != Range(19, 222) {
		t.Errorf("expected Range(19,222), got %v", ranges[2])
	}
}

func TestSerializeDeserializeRoundtrip(t *testing.T) {
	log := NewAuthorshipLog()
	log.Metadata.BaseCommitSHA = "abc123"

	file1 := log.GetOrCreateFile("src/file.rs")
	file1.AddEntry(NewAttestationEntry(
		"a1b2c3d4e5f6g7h8",
		[]LineRange{SingleLine(1), SingleLine(2), Range(19, 222)},
	))
	file1.AddEntry(NewAttestationEntry(
		"1234567890abcdef",
		[]LineRange{Range(400, 405)},
	))

	file2 := log.GetOrCreateFile("src/file2.rs")
	file2.AddEntry(NewAttestationEntry(
		"1234567890abcdef",
		[]LineRange{Range(1, 111), SingleLine(245), SingleLine(260)},
	))

	serialized, err := log.SerializeToString()
	if err != nil {
		t.Fatal(err)
	}

	deserialized, err := DeserializeFromString(serialized)
	if err != nil {
		t.Fatal(err)
	}

	if len(deserialized.Attestations) != 2 {
		t.Fatalf("expected 2 attestations, got %d", len(deserialized.Attestations))
	}
	if deserialized.Attestations[0].FilePath != "src/file.rs" {
		t.Errorf("expected src/file.rs, got %s", deserialized.Attestations[0].FilePath)
	}
	if len(deserialized.Attestations[0].Entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(deserialized.Attestations[0].Entries))
	}
	if deserialized.Attestations[1].FilePath != "src/file2.rs" {
		t.Errorf("expected src/file2.rs, got %s", deserialized.Attestations[1].FilePath)
	}
	if len(deserialized.Attestations[1].Entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(deserialized.Attestations[1].Entries))
	}
	if deserialized.Metadata.SchemaVersion != AuthorshipLogVersion {
		t.Errorf("expected %s, got %s", AuthorshipLogVersion, deserialized.Metadata.SchemaVersion)
	}
	if deserialized.Metadata.BaseCommitSHA != "abc123" {
		t.Errorf("expected abc123, got %s", deserialized.Metadata.BaseCommitSHA)
	}
}

func TestQuotedFilePaths(t *testing.T) {
	log := NewAuthorshipLog()
	file := log.GetOrCreateFile("path with spaces/file.rs")
	file.AddEntry(NewAttestationEntry(
		"a1b2c3d4e5f6g7h8",
		[]LineRange{Range(1, 10)},
	))

	serialized, err := log.SerializeToString()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(serialized, "\"path with spaces/file.rs\"") {
		t.Errorf("expected quoted path in output, got: %s", serialized)
	}

	deserialized, err := DeserializeFromString(serialized)
	if err != nil {
		t.Fatal(err)
	}
	if deserialized.Attestations[0].FilePath != "path with spaces/file.rs" {
		t.Errorf("expected unquoted path, got %s", deserialized.Attestations[0].FilePath)
	}
}

func TestEmptyLogRoundtrip(t *testing.T) {
	log := NewAuthorshipLog()
	serialized, err := log.SerializeToString()
	if err != nil {
		t.Fatal(err)
	}
	deserialized, err := DeserializeFromString(serialized)
	if err != nil {
		t.Fatal(err)
	}
	if len(deserialized.Attestations) != 0 {
		t.Errorf("expected no attestations, got %d", len(deserialized.Attestations))
	}
	if len(deserialized.Metadata.Prompts) != 0 {
		t.Errorf("expected no prompts, got %d", len(deserialized.Metadata.Prompts))
	}
}

func TestMetadataWithPrompts(t *testing.T) {
	log := NewAuthorshipLog()
	hash := GenerateShortHash("session-1", "claude")

	log.Metadata.Prompts[hash] = &PromptRecord{
		AgentID: AgentId{
			Tool:  "claude",
			Model: "sonnet-4",
			ID:    "session-1",
		},
		HumanAuthor:   "alice",
		TotalAdds:     30,
		TotalDels:     5,
		AcceptedLines: 25,
	}

	file := log.GetOrCreateFile("src/main.rs")
	file.AddEntry(NewAttestationEntry(
		hash,
		[]LineRange{Range(1, 10), Range(15, 20)},
	))

	serialized, err := log.SerializeToString()
	if err != nil {
		t.Fatal(err)
	}
	deserialized, err := DeserializeFromString(serialized)
	if err != nil {
		t.Fatal(err)
	}

	prompt, ok := deserialized.Metadata.Prompts[hash]
	if !ok {
		t.Fatal("prompt not found for hash")
	}
	if prompt.AgentID.Tool != "claude" {
		t.Errorf("expected claude, got %s", prompt.AgentID.Tool)
	}
	if prompt.AgentID.Model != "sonnet-4" {
		t.Errorf("expected sonnet-4, got %s", prompt.AgentID.Model)
	}
	if prompt.HumanAuthor != "alice" {
		t.Errorf("expected alice, got %s", prompt.HumanAuthor)
	}
	if prompt.TotalAdds != 30 {
		t.Errorf("expected 30, got %d", prompt.TotalAdds)
	}
	if prompt.AcceptedLines != 25 {
		t.Errorf("expected 25, got %d", prompt.AcceptedLines)
	}
}
