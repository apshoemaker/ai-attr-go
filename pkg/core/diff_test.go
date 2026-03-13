package core

import (
	"testing"
)

func TestIdenticalContent(t *testing.T) {
	content := "line 1\nline 2\nline 3\n"
	result := ChangedLines(content, content)
	if len(result.AddedLines) != 0 {
		t.Errorf("expected no added lines, got %v", result.AddedLines)
	}
	if result.TotalAdditions != 0 {
		t.Errorf("expected 0 additions, got %d", result.TotalAdditions)
	}
	if result.TotalDeletions != 0 {
		t.Errorf("expected 0 deletions, got %d", result.TotalDeletions)
	}
}

func TestEmptyToContent(t *testing.T) {
	result := ChangedLines("", "line 1\nline 2\nline 3\n")
	expected := []uint32{1, 2, 3}
	if len(result.AddedLines) != len(expected) {
		t.Fatalf("expected %v, got %v", expected, result.AddedLines)
	}
	for i, v := range expected {
		if result.AddedLines[i] != v {
			t.Errorf("index %d: expected %d, got %d", i, v, result.AddedLines[i])
		}
	}
	if result.TotalAdditions != 3 {
		t.Errorf("expected 3 additions, got %d", result.TotalAdditions)
	}
	if result.TotalDeletions != 0 {
		t.Errorf("expected 0 deletions, got %d", result.TotalDeletions)
	}
}

func TestContentToEmpty(t *testing.T) {
	result := ChangedLines("line 1\nline 2\nline 3\n", "")
	if len(result.AddedLines) != 0 {
		t.Errorf("expected no added lines, got %v", result.AddedLines)
	}
	if result.TotalAdditions != 0 {
		t.Errorf("expected 0 additions, got %d", result.TotalAdditions)
	}
	if result.TotalDeletions != 3 {
		t.Errorf("expected 3 deletions, got %d", result.TotalDeletions)
	}
}

func TestSingleLineModified(t *testing.T) {
	old := "line 1\nline 2\nline 3\n"
	new := "line 1\nmodified\nline 3\n"
	result := ChangedLines(old, new)
	if len(result.AddedLines) != 1 || result.AddedLines[0] != 2 {
		t.Errorf("expected [2], got %v", result.AddedLines)
	}
	if result.TotalAdditions != 1 {
		t.Errorf("expected 1 addition, got %d", result.TotalAdditions)
	}
	if result.TotalDeletions != 1 {
		t.Errorf("expected 1 deletion, got %d", result.TotalDeletions)
	}
}

func TestLinesInsertedAtEnd(t *testing.T) {
	old := "line 1\nline 2\n"
	new := "line 1\nline 2\nline 3\nline 4\n"
	result := ChangedLines(old, new)
	expected := []uint32{3, 4}
	if len(result.AddedLines) != len(expected) {
		t.Fatalf("expected %v, got %v", expected, result.AddedLines)
	}
	for i, v := range expected {
		if result.AddedLines[i] != v {
			t.Errorf("index %d: expected %d, got %d", i, v, result.AddedLines[i])
		}
	}
	if result.TotalAdditions != 2 {
		t.Errorf("expected 2 additions, got %d", result.TotalAdditions)
	}
	if result.TotalDeletions != 0 {
		t.Errorf("expected 0 deletions, got %d", result.TotalDeletions)
	}
}

func TestMixedChanges(t *testing.T) {
	old := "aaa\nbbb\nccc\nddd\n"
	new := "aaa\nBBB\nccc\neee\nfff\n"
	result := ChangedLines(old, new)

	addedSet := make(map[uint32]bool)
	for _, l := range result.AddedLines {
		addedSet[l] = true
	}

	// bbb -> BBB (modify line 2), ddd removed, eee+fff added at lines 4,5
	if !addedSet[2] {
		t.Error("expected line 2 to be added")
	}
	if !addedSet[4] {
		t.Error("expected line 4 to be added")
	}
	if !addedSet[5] {
		t.Error("expected line 5 to be added")
	}
	if result.TotalAdditions != 3 {
		t.Errorf("expected 3 additions, got %d", result.TotalAdditions)
	}
	if result.TotalDeletions != 2 {
		t.Errorf("expected 2 deletions, got %d", result.TotalDeletions)
	}
}
