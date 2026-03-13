package core

import (
	"strconv"
	"strings"
)

// BlameLine represents a parsed line from git blame --porcelain output.
type BlameLine struct {
	CommitSHA string
	FinalLine uint32
	Content   string
}

// ParsePorcelainBlame parses git blame --porcelain output into per-line records.
func ParsePorcelainBlame(output string) []BlameLine {
	var results []BlameLine
	lines := strings.Split(output, "\n")
	i := 0

	for i < len(lines) {
		header := lines[i]
		parts := strings.Fields(header)
		if len(parts) < 3 {
			i++
			continue
		}

		commitSHA := parts[0]
		finalLine, err := strconv.ParseUint(parts[2], 10, 32)
		if err != nil {
			i++
			continue
		}

		// Skip header lines until we find the content line (starts with \t)
		i++
		content := ""
		for i < len(lines) {
			if strings.HasPrefix(lines[i], "\t") {
				content = lines[i][1:] // strip the leading tab
				i++
				break
			}
			i++
		}

		results = append(results, BlameLine{
			CommitSHA: commitSHA,
			FinalLine: uint32(finalLine),
			Content:   content,
		})
	}

	return results
}
