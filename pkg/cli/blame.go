package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/apshoemaker/ai-attr/pkg/core"
	"github.com/apshoemaker/ai-attr/pkg/storage"
)

// RunBlame displays line-level AI attribution for a file.
func RunBlame(file string, jsonOutput bool) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	blameOutput, err := storage.GitBlamePorcelain(cwd, file)
	if err != nil {
		return err
	}
	blameLines := core.ParsePorcelainBlame(blameOutput)

	if len(blameLines) == 0 {
		fmt.Fprintf(os.Stderr, "No blame data for file: %s\n", file)
		return nil
	}

	// Collect unique commits and fetch attribution notes
	seen := make(map[string]bool)
	var uniqueCommits []string
	for _, bl := range blameLines {
		if !seen[bl.CommitSHA] {
			seen[bl.CommitSHA] = true
			uniqueCommits = append(uniqueCommits, bl.CommitSHA)
		}
	}

	notesCache := make(map[string]*core.AuthorshipLog)
	for _, sha := range uniqueCommits {
		log, ok, _ := storage.GetAuthorship(cwd, sha)
		if ok {
			notesCache[sha] = log
		}
	}

	if jsonOutput {
		var result []map[string]interface{}
		for _, bl := range blameLines {
			attr, tool, model := classifyLine(notesCache, bl.CommitSHA, file, bl.FinalLine)
			shortSHA := bl.CommitSHA
			if len(shortSHA) > 12 {
				shortSHA = shortSHA[:12]
			}
			result = append(result, map[string]interface{}{
				"line":        bl.FinalLine,
				"commit":      shortSHA,
				"attribution": attr,
				"tool":        tool,
				"model":       model,
				"content":     bl.Content,
			})
		}
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(data))
	} else {
		for _, bl := range blameLines {
			attr, _, _ := classifyLine(notesCache, bl.CommitSHA, file, bl.FinalLine)
			shortSHA := bl.CommitSHA
			if len(shortSHA) > 8 {
				shortSHA = shortSHA[:8]
			}
			marker := "??"
			switch attr {
			case "ai":
				marker = "AI"
			case "human":
				marker = "  "
			}
			fmt.Printf("%s %s %4d %s\n", shortSHA, marker, bl.FinalLine, bl.Content)
		}
	}

	return nil
}

func classifyLine(notesCache map[string]*core.AuthorshipLog, commitSHA, filePath string, lineNum uint32) (string, string, string) {
	log, ok := notesCache[commitSHA]
	if !ok || log == nil {
		return "human", "", ""
	}

	var fileAtt *core.FileAttestation
	for i := range log.Attestations {
		if log.Attestations[i].FilePath == filePath {
			fileAtt = &log.Attestations[i]
			break
		}
	}
	if fileAtt == nil {
		return "human", "", ""
	}

	// Check AI-attributed ranges
	for _, entry := range fileAtt.Entries {
		if entry.Hash == core.HumanSentinel {
			continue
		}
		for _, r := range entry.LineRanges {
			if r.Contains(lineNum) {
				prompt, ok := log.Metadata.Prompts[entry.Hash]
				if ok {
					return "ai", prompt.AgentID.Tool, prompt.AgentID.Model
				}
				return "ai", "", ""
			}
		}
	}

	// Check explicit human ranges
	for _, entry := range fileAtt.Entries {
		if entry.Hash == core.HumanSentinel {
			for _, r := range entry.LineRanges {
				if r.Contains(lineNum) {
					return "human", "", ""
				}
			}
		}
	}

	return "human", "", ""
}
