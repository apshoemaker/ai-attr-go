package core

import "strings"

// DiffResult holds the result of diffing two strings.
type DiffResult struct {
	// 1-indexed line numbers that were added/modified in the new content
	AddedLines     []uint32
	TotalAdditions uint32
	TotalDeletions uint32
}

// ChangedLines computes which lines changed between old and new content.
// Returns 1-indexed line numbers of lines present in new that are additions.
// Uses a Histogram diff algorithm.
func ChangedLines(old, new string) DiffResult {
	if old == new {
		return DiffResult{}
	}

	oldLines := splitLines(old)
	newLines := splitLines(new)

	// Compute the LCS-based diff using histogram approach
	matched := histogramDiff(oldLines, newLines)

	// matched[i] = true means newLines[i] was matched to an old line (not added)
	var addedLines []uint32
	for i, isMatched := range matched {
		if !isMatched {
			addedLines = append(addedLines, uint32(i+1)) // 1-indexed
		}
	}

	// Count deletions: old lines not matched to any new line
	oldMatched := histogramDiffOld(oldLines, newLines)
	var totalDeletions uint32
	for _, isMatched := range oldMatched {
		if !isMatched {
			totalDeletions++
		}
	}

	return DiffResult{
		AddedLines:     addedLines,
		TotalAdditions: uint32(len(addedLines)),
		TotalDeletions: totalDeletions,
	}
}

func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	lines := strings.Split(s, "\n")
	// Remove trailing empty string from trailing newline
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// histogramDiff implements a histogram-based diff algorithm.
// Returns a boolean slice for newLines where true = matched (not added).
func histogramDiff(oldLines, newLines []string) []bool {
	newMatched := make([]bool, len(newLines))
	oldUsed := make([]bool, len(oldLines))
	histogramRecurse(oldLines, newLines, 0, len(oldLines), 0, len(newLines), oldUsed, newMatched)
	return newMatched
}

// histogramDiffOld returns which old lines were matched.
func histogramDiffOld(oldLines, newLines []string) []bool {
	oldUsed := make([]bool, len(oldLines))
	newMatched := make([]bool, len(newLines))
	histogramRecurse(oldLines, newLines, 0, len(oldLines), 0, len(newLines), oldUsed, newMatched)
	return oldUsed
}

func histogramRecurse(oldLines, newLines []string, oldStart, oldEnd, newStart, newEnd int, oldUsed, newMatched []bool) {
	if oldStart >= oldEnd || newStart >= newEnd {
		return
	}

	// Build occurrence count for old lines in this region
	occurrences := make(map[string]int)
	for i := oldStart; i < oldEnd; i++ {
		occurrences[oldLines[i]]++
	}

	// Find the lowest-occurrence-count line that also appears in new
	bestLine := ""
	bestCount := int(^uint(0) >> 1) // max int
	bestNewIdx := -1

	for i := newStart; i < newEnd; i++ {
		line := newLines[i]
		if count, ok := occurrences[line]; ok && count < bestCount {
			bestCount = count
			bestLine = line
			bestNewIdx = i
		}
	}

	if bestNewIdx == -1 {
		// No common lines in this region
		return
	}

	// Find the corresponding position in old for this anchor line
	// Use patience-like approach: find the best match position
	bestOldIdx := -1
	for i := oldStart; i < oldEnd; i++ {
		if oldLines[i] == bestLine && !oldUsed[i] {
			bestOldIdx = i
			break
		}
	}

	if bestOldIdx == -1 {
		return
	}

	// Extend the match as far as possible in both directions
	matchStartOld := bestOldIdx
	matchStartNew := bestNewIdx
	for matchStartOld > oldStart && matchStartNew > newStart &&
		oldLines[matchStartOld-1] == newLines[matchStartNew-1] &&
		!oldUsed[matchStartOld-1] {
		matchStartOld--
		matchStartNew--
	}

	matchEndOld := bestOldIdx + 1
	matchEndNew := bestNewIdx + 1
	for matchEndOld < oldEnd && matchEndNew < newEnd &&
		oldLines[matchEndOld] == newLines[matchEndNew] &&
		!oldUsed[matchEndOld] {
		matchEndOld++
		matchEndNew++
	}

	// Mark matched lines
	for i := 0; i < matchEndOld-matchStartOld; i++ {
		oldUsed[matchStartOld+i] = true
		newMatched[matchStartNew+i] = true
	}

	// Recurse on regions before and after the match
	histogramRecurse(oldLines, newLines, oldStart, matchStartOld, newStart, matchStartNew, oldUsed, newMatched)
	histogramRecurse(oldLines, newLines, matchEndOld, oldEnd, matchEndNew, newEnd, oldUsed, newMatched)
}
