package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/apshoemaker/ai-attr/pkg/storage"
)

type toolStats struct {
	Sessions      uint32 `json:"sessions"`
	TotalAdds     uint32 `json:"total_additions"`
	TotalDels     uint32 `json:"total_deletions"`
	AcceptedLines uint32 `json:"accepted_lines"`
}

// RunStats displays AI composition statistics for a commit range.
func RunStats(commitRange string, jsonOutput bool) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	commits, err := resolveCommits(cwd, commitRange)
	if err != nil {
		return err
	}

	stats := make(map[string]*toolStats)
	var totalCommits, attributedCommits uint32

	for _, sha := range commits {
		totalCommits++
		log, ok, _ := storage.GetAuthorship(cwd, sha)
		if !ok || log == nil {
			continue
		}
		attributedCommits++
		for _, prompt := range log.Metadata.Prompts {
			tool := prompt.AgentID.Tool
			s, ok := stats[tool]
			if !ok {
				s = &toolStats{}
				stats[tool] = s
			}
			s.TotalAdds += prompt.TotalAdds
			s.TotalDels += prompt.TotalDels
			s.AcceptedLines += prompt.AcceptedLines
			s.Sessions++
		}
	}

	if jsonOutput {
		tools := make([]map[string]interface{}, 0)
		sortedTools := make([]string, 0, len(stats))
		for k := range stats {
			sortedTools = append(sortedTools, k)
		}
		sort.Strings(sortedTools)
		for _, tool := range sortedTools {
			s := stats[tool]
			tools = append(tools, map[string]interface{}{
				"tool":            tool,
				"sessions":        s.Sessions,
				"total_additions": s.TotalAdds,
				"total_deletions": s.TotalDels,
				"accepted_lines":  s.AcceptedLines,
			})
		}
		output := map[string]interface{}{
			"total_commits":      totalCommits,
			"attributed_commits": attributedCommits,
			"tools":              tools,
		}
		data, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Printf("Commits: %d total, %d with AI attribution\n", totalCommits, attributedCommits)
		sortedTools := make([]string, 0, len(stats))
		for k := range stats {
			sortedTools = append(sortedTools, k)
		}
		sort.Strings(sortedTools)
		for _, tool := range sortedTools {
			s := stats[tool]
			fmt.Printf("  %s: %d sessions, +%d -%d, %d lines attributed\n",
				tool, s.Sessions, s.TotalAdds, s.TotalDels, s.AcceptedLines)
		}
	}

	return nil
}

func resolveCommits(repoDir, rangeArg string) ([]string, error) {
	if rangeArg == "" {
		rangeArg = "HEAD~10..HEAD"
	}

	cmd := exec.Command("git", "log", "--format=%H", rangeArg)
	cmd.Dir = repoDir
	out, err := cmd.Output()
	if err != nil {
		// Fallback: just HEAD
		head, err := storage.GetHeadSHA(repoDir)
		if err != nil {
			return nil, err
		}
		return []string{head}, nil
	}

	var commits []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line != "" {
			commits = append(commits, line)
		}
	}
	return commits, nil
}
