package core

import "testing"

func TestParsePorcelainSingleCommit(t *testing.T) {
	output := `abc123def456 1 1 3
author Test User
author-mail <test@test.com>
author-time 1700000000
author-tz +0000
committer Test User
committer-mail <test@test.com>
committer-time 1700000000
committer-tz +0000
summary initial commit
filename src/main.rs
	line one
abc123def456 2 2
	line two
abc123def456 3 3
	line three
`
	result := ParsePorcelainBlame(output)
	if len(result) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(result))
	}
	if result[0].CommitSHA != "abc123def456" {
		t.Errorf("expected sha abc123def456, got %s", result[0].CommitSHA)
	}
	if result[0].FinalLine != 1 {
		t.Errorf("expected line 1, got %d", result[0].FinalLine)
	}
	if result[0].Content != "line one" {
		t.Errorf("expected 'line one', got '%s'", result[0].Content)
	}
	if result[1].FinalLine != 2 {
		t.Errorf("expected line 2, got %d", result[1].FinalLine)
	}
	if result[2].FinalLine != 3 {
		t.Errorf("expected line 3, got %d", result[2].FinalLine)
	}
	if result[2].Content != "line three" {
		t.Errorf("expected 'line three', got '%s'", result[2].Content)
	}
}

func TestParsePorcelainMultipleCommits(t *testing.T) {
	output := `aaaa 1 1 1
author A
filename f.rs
	first
bbbb 2 2 1
author B
filename f.rs
	second
`
	result := ParsePorcelainBlame(output)
	if len(result) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(result))
	}
	if result[0].CommitSHA != "aaaa" {
		t.Errorf("expected sha aaaa, got %s", result[0].CommitSHA)
	}
	if result[1].CommitSHA != "bbbb" {
		t.Errorf("expected sha bbbb, got %s", result[1].CommitSHA)
	}
}

func TestParsePorcelainEmpty(t *testing.T) {
	result := ParsePorcelainBlame("")
	if len(result) != 0 {
		t.Errorf("expected empty, got %d lines", len(result))
	}
}
