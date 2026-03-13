package cli

import (
	"fmt"
	"os"
	"sort"

	"github.com/apshoemaker/ai-attr/pkg/core"
	"github.com/apshoemaker/ai-attr/pkg/storage"
)

// RunCommit is called by post-commit hook to consolidate checkpoints into a git note.
func RunCommit() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	return CommitAt(cwd)
}

// CommitAt consolidates checkpoints into a git note for the repo at the given path.
func CommitAt(repoWorkdir string) error {
	gitDir, workDir, err := storage.FindGitDir(repoWorkdir)
	if err != nil {
		return err
	}
	store, err := storage.NewRepoStorage(gitDir, workDir)
	if err != nil {
		return err
	}
	headSHA, err := storage.GetHeadSHA(workDir)
	if err != nil {
		return err
	}

	workingLogDirs, err := store.ListWorkingLogDirs()
	if err != nil {
		return err
	}
	if len(workingLogDirs) == 0 {
		return nil
	}

	// Collect all entries
	var allEntries []core.CheckpointEntry
	for _, dir := range workingLogDirs {
		entries, err := core.ReadEntries(dir[1])
		if err != nil {
			return err
		}
		allEntries = append(allEntries, entries...)
	}

	if len(allEntries) == 0 {
		for _, dir := range workingLogDirs {
			store.DeleteWorkingLog(dir[0])
		}
		return nil
	}

	userName, _ := storage.GetUserName(workDir)
	userEmail, _ := storage.GetUserEmail(workDir)
	humanAuthor := userName
	if userEmail != "" {
		humanAuthor = fmt.Sprintf("%s <%s>", userName, userEmail)
	}

	baseSHA := workingLogDirs[0][0]

	// Group entries by session_id
	sessions := make(map[string][]*core.CheckpointEntry)
	for i := range allEntries {
		e := &allEntries[i]
		sessions[e.SessionID] = append(sessions[e.SessionID], e)
	}

	// Collect all files with AI edits
	filesWithAIEdits := make(map[string]bool)
	for _, e := range allEntries {
		filesWithAIEdits[e.FilePath] = true
	}

	// Two-phase diff
	aiLinesByFile := make(map[string]map[uint32]bool)
	humanLinesByFile := make(map[string]map[uint32]bool)

	for filePath := range filesWithAIEdits {
		baseContent, _, _ := storage.GitShowFile(workDir, baseSHA, filePath)
		headContent, _, _ := storage.GitShowFile(workDir, headSHA, filePath)

		postAIContent, hasPostSnapshot, _ := store.LoadPostSnapshot(filePath)

		if hasPostSnapshot {
			// Two-phase diff
			allChanged := core.ChangedLines(baseContent, headContent)
			allChangedSet := toSet(allChanged.AddedLines)

			humanDiff := core.ChangedLines(postAIContent, headContent)
			humanChangedSet := toSet(humanDiff.AddedLines)

			// AI = all_changed - human_changed
			aiSet := difference(allChangedSet, humanChangedSet)
			// Human = all_changed ∩ human_changed
			humanSet := intersection(allChangedSet, humanChangedSet)

			aiLinesByFile[filePath] = aiSet
			humanLinesByFile[filePath] = humanSet
		} else {
			// Fallback: raw checkpoint data
			lines := make(map[uint32]bool)
			for _, e := range allEntries {
				if e.FilePath == filePath {
					for _, l := range e.AddedLines {
						lines[l] = true
					}
				}
			}
			aiLinesByFile[filePath] = lines
		}
	}

	log := core.NewAuthorshipLog()
	log.Metadata.BaseCommitSHA = baseSHA

	// Build prompt records sorted by session_id
	sortedSessions := sortedKeys(sessions)
	for _, sessionID := range sortedSessions {
		entries := sessions[sessionID]
		first := entries[0]
		hash := core.GenerateShortHash(sessionID, first.Tool)

		var totalAdds, totalDels uint32
		for _, e := range entries {
			totalAdds += e.TotalAdditions
			totalDels += e.TotalDeletions
		}

		var accepted uint32
		for _, e := range entries {
			if aiLines, ok := aiLinesByFile[e.FilePath]; ok {
				accepted += uint32(len(aiLines))
			}
		}

		model := ""
		if first.Model != nil {
			model = *first.Model
		}

		log.Metadata.Prompts[hash] = &core.PromptRecord{
			AgentID: core.AgentId{
				Tool:  first.Tool,
				ID:    sessionID,
				Model: model,
			},
			HumanAuthor:   humanAuthor,
			TotalAdds:     totalAdds,
			TotalDels:     totalDels,
			AcceptedLines: accepted,
		}
	}

	// Build attestation entries for AI lines
	for _, sessionID := range sortedSessions {
		entries := sessions[sessionID]
		first := entries[0]
		hash := core.GenerateShortHash(sessionID, first.Tool)

		sessionFiles := make(map[string]bool)
		for _, e := range entries {
			sessionFiles[e.FilePath] = true
		}

		sortedFiles := sortedKeysFromBoolMap(sessionFiles)
		for _, filePath := range sortedFiles {
			if aiLines, ok := aiLinesByFile[filePath]; ok && len(aiLines) > 0 {
				lines := sortedUint32FromSet(aiLines)
				ranges := core.CompressLines(lines)
				file := log.GetOrCreateFile(filePath)
				file.AddEntry(core.NewAttestationEntry(hash, ranges))
			}
		}
	}

	// Build attestation entries for human lines
	var totalHumanAdds uint32
	sortedHumanFiles := sortedKeysFromBoolMapOfSets(humanLinesByFile)
	for _, filePath := range sortedHumanFiles {
		humanLines := humanLinesByFile[filePath]
		if len(humanLines) == 0 {
			continue
		}
		totalHumanAdds += uint32(len(humanLines))
		lines := sortedUint32FromSet(humanLines)
		ranges := core.CompressLines(lines)
		file := log.GetOrCreateFile(filePath)
		file.AddEntry(core.NewAttestationEntry(core.HumanSentinel, ranges))
	}

	if totalHumanAdds > 0 {
		log.Metadata.Prompts[core.HumanSentinel] = &core.PromptRecord{
			AgentID: core.AgentId{
				Tool: "human",
			},
			HumanAuthor:   humanAuthor,
			TotalAdds:     totalHumanAdds,
			AcceptedLines: totalHumanAdds,
		}
	}

	serialized, err := log.SerializeToString()
	if err != nil {
		return fmt.Errorf("failed to serialize authorship log: %w", err)
	}
	if err := storage.NotesAdd(workDir, headSHA, serialized); err != nil {
		return err
	}

	// Clean up
	for _, dir := range workingLogDirs {
		store.DeleteWorkingLog(dir[0])
	}
	store.ClearAllPostSnapshots()

	return nil
}

func toSet(lines []uint32) map[uint32]bool {
	s := make(map[uint32]bool, len(lines))
	for _, l := range lines {
		s[l] = true
	}
	return s
}

func difference(a, b map[uint32]bool) map[uint32]bool {
	result := make(map[uint32]bool)
	for k := range a {
		if !b[k] {
			result[k] = true
		}
	}
	return result
}

func intersection(a, b map[uint32]bool) map[uint32]bool {
	result := make(map[uint32]bool)
	for k := range a {
		if b[k] {
			result[k] = true
		}
	}
	return result
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func sortedKeysFromBoolMap(m map[string]bool) []string {
	return sortedKeys(m)
}

func sortedKeysFromBoolMapOfSets(m map[string]map[uint32]bool) []string {
	return sortedKeys(m)
}

func sortedUint32FromSet(s map[uint32]bool) []uint32 {
	result := make([]uint32, 0, len(s))
	for k := range s {
		result = append(result, k)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}
