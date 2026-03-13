package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// CheckpointEntry records which lines an AI agent touched in a file.
type CheckpointEntry struct {
	SessionID      string   `json:"session_id"`
	Tool           string   `json:"tool"`
	Model          *string  `json:"model,omitempty"`
	FilePath       string   `json:"file_path"`
	AddedLines     []uint32 `json:"added_lines"`
	TotalAdditions uint32   `json:"total_additions"`
	TotalDeletions uint32   `json:"total_deletions"`
	Timestamp      uint64   `json:"timestamp"`
}

// WriteEntry writes a checkpoint entry to the working log directory with auto-incrementing filename.
func WriteEntry(logDir string, entry *CheckpointEntry) error {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	nextIndex, err := nextEntryIndex(logDir)
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("entry-%d.json", nextIndex)
	path := filepath.Join(logDir, filename)

	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// ReadEntries reads all checkpoint entries from a working log directory, sorted by index.
func ReadEntries(logDir string) ([]CheckpointEntry, error) {
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		return nil, nil
	}

	dirEntries, err := os.ReadDir(logDir)
	if err != nil {
		return nil, err
	}

	type indexedEntry struct {
		index uint32
		entry CheckpointEntry
	}
	var entries []indexedEntry

	for _, de := range dirEntries {
		name := de.Name()
		if idx, ok := parseEntryIndex(name); ok {
			data, err := os.ReadFile(filepath.Join(logDir, name))
			if err != nil {
				return nil, err
			}
			var entry CheckpointEntry
			if err := json.Unmarshal(data, &entry); err != nil {
				return nil, err
			}
			entries = append(entries, indexedEntry{index: idx, entry: entry})
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].index < entries[j].index
	})

	result := make([]CheckpointEntry, len(entries))
	for i, ie := range entries {
		result[i] = ie.entry
	}
	return result, nil
}

func nextEntryIndex(logDir string) (uint32, error) {
	dirEntries, err := os.ReadDir(logDir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}

	var maxIndex uint32
	foundAny := false

	for _, de := range dirEntries {
		if idx, ok := parseEntryIndex(de.Name()); ok {
			foundAny = true
			if idx >= maxIndex {
				maxIndex = idx + 1
			}
		}
	}

	if !foundAny {
		return 0, nil
	}
	return maxIndex, nil
}

func parseEntryIndex(filename string) (uint32, bool) {
	if !strings.HasPrefix(filename, "entry-") || !strings.HasSuffix(filename, ".json") {
		return 0, false
	}
	stem := strings.TrimPrefix(filename, "entry-")
	stem = strings.TrimSuffix(stem, ".json")
	n, err := strconv.ParseUint(stem, 10, 32)
	if err != nil {
		return 0, false
	}
	return uint32(n), true
}
