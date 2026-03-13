package core

import (
	"fmt"
	"strconv"
	"strings"

	aierr "github.com/apshoemaker/ai-attr/pkg/errors"
)

// LineRange represents either a single line or a range of lines (1-indexed, inclusive).
type LineRange struct {
	Start uint32
	End   uint32 // Equal to Start for single-line ranges
}

func SingleLine(n uint32) LineRange {
	return LineRange{Start: n, End: n}
}

func Range(start, end uint32) LineRange {
	return LineRange{Start: start, End: end}
}

func (lr LineRange) IsSingle() bool {
	return lr.Start == lr.End
}

func (lr LineRange) String() string {
	if lr.IsSingle() {
		return strconv.FormatUint(uint64(lr.Start), 10)
	}
	return fmt.Sprintf("%d-%d", lr.Start, lr.End)
}

func (lr LineRange) Contains(line uint32) bool {
	return line >= lr.Start && line <= lr.End
}

func (lr LineRange) Overlaps(other LineRange) bool {
	return lr.Start <= other.End && other.Start <= lr.End
}

// Remove removes a line or range from this range, returning the remaining parts.
func (lr LineRange) Remove(toRemove LineRange) []LineRange {
	// No overlap
	if toRemove.Start > lr.End || toRemove.End < lr.Start {
		return []LineRange{lr}
	}

	// Complete removal
	if toRemove.Start <= lr.Start && toRemove.End >= lr.End {
		return nil
	}

	var result []LineRange
	if lr.Start < toRemove.Start {
		result = append(result, Range(lr.Start, toRemove.Start-1))
	}
	if lr.End > toRemove.End {
		result = append(result, Range(toRemove.End+1, lr.End))
	}
	return result
}

// Expand returns all individual line numbers in the range.
func (lr LineRange) Expand() []uint32 {
	lines := make([]uint32, 0, lr.End-lr.Start+1)
	for i := lr.Start; i <= lr.End; i++ {
		lines = append(lines, i)
	}
	return lines
}

// CompressLines converts a sorted list of line numbers into compressed ranges.
func CompressLines(lines []uint32) []LineRange {
	if len(lines) == 0 {
		return nil
	}

	ranges := make([]LineRange, 0)
	currentStart := lines[0]
	currentEnd := lines[0]

	for _, line := range lines[1:] {
		if line == currentEnd+1 {
			currentEnd = line
		} else {
			if currentStart == currentEnd {
				ranges = append(ranges, SingleLine(currentStart))
			} else {
				ranges = append(ranges, Range(currentStart, currentEnd))
			}
			currentStart = line
			currentEnd = line
		}
	}

	if currentStart == currentEnd {
		ranges = append(ranges, SingleLine(currentStart))
	} else {
		ranges = append(ranges, Range(currentStart, currentEnd))
	}

	return ranges
}

// FormatLineRanges formats line ranges as comma-separated values, sorted by start.
func FormatLineRanges(ranges []LineRange) string {
	sorted := make([]LineRange, len(ranges))
	copy(sorted, ranges)
	sortLineRanges(sorted)

	parts := make([]string, len(sorted))
	for i, r := range sorted {
		parts[i] = r.String()
	}
	return strings.Join(parts, ",")
}

// ParseLineRanges parses a string like "1,2,19-222" into LineRanges.
func ParseLineRanges(input string) ([]LineRange, error) {
	var ranges []LineRange
	for _, part := range strings.Split(input, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if dashPos := strings.Index(part, "-"); dashPos >= 0 {
			start, err := strconv.ParseUint(part[:dashPos], 10, 32)
			if err != nil {
				return nil, aierr.NewParse(fmt.Sprintf("invalid line number in range: %s", part))
			}
			end, err := strconv.ParseUint(part[dashPos+1:], 10, 32)
			if err != nil {
				return nil, aierr.NewParse(fmt.Sprintf("invalid line number in range: %s", part))
			}
			ranges = append(ranges, Range(uint32(start), uint32(end)))
		} else {
			line, err := strconv.ParseUint(part, 10, 32)
			if err != nil {
				return nil, aierr.NewParse(fmt.Sprintf("invalid line number: %s", part))
			}
			ranges = append(ranges, SingleLine(uint32(line)))
		}
	}
	return ranges, nil
}

func sortLineRanges(ranges []LineRange) {
	for i := 1; i < len(ranges); i++ {
		key := ranges[i]
		j := i - 1
		for j >= 0 && ranges[j].Start > key.Start {
			ranges[j+1] = ranges[j]
			j--
		}
		ranges[j+1] = key
	}
}
