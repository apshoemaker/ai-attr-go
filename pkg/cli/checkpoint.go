package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/apshoemaker/ai-attr/pkg/adapters"
	"github.com/apshoemaker/ai-attr/pkg/core"
	"github.com/apshoemaker/ai-attr/pkg/storage"
)

// RunCheckpoint is called by agent hooks to record a checkpoint.
func RunCheckpoint(agent string, hookInput bool) error {
	if !hookInput {
		return fmt.Errorf("checkpoint requires --hook-input")
	}

	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	return ProcessHookInput(string(input), agent, cwd)
}

// ProcessHookInput processes hook input for a repo at the given path (testable).
func ProcessHookInput(input, agent, repoWorkdir string) error {
	adapter := adapters.NewAdapter(agent)
	ctx, err := adapter.ParseHookInput(input)
	if err != nil {
		return err
	}

	gitDir := filepath.Join(repoWorkdir, ".git")
	store, err := storage.NewRepoStorage(gitDir, repoWorkdir)
	if err != nil {
		return err
	}
	headSHA, err := storage.GetHeadSHA(repoWorkdir)
	if err != nil {
		return err
	}

	switch ctx.Phase {
	case adapters.SessionStart:
		if ctx.Model != "" {
			store.SaveSessionModel(ctx.SessionID, ctx.Model)
		}

	case adapters.PreToolUse:
		if ctx.FilePath != "" {
			relative := store.ToRelativePath(ctx.FilePath)
			absPath := filepath.Join(repoWorkdir, relative)
			content := ""
			if data, err := os.ReadFile(absPath); err == nil {
				content = string(data)
			}
			store.SaveSnapshot(relative, content)
		}

	case adapters.PostToolUse:
		if ctx.FilePath != "" {
			relative := store.ToRelativePath(ctx.FilePath)
			absPath := filepath.Join(repoWorkdir, relative)

			oldContent, _, _ := store.LoadSnapshot(relative)
			newContent := ""
			if data, err := os.ReadFile(absPath); err == nil {
				newContent = string(data)
			}

			diffResult := core.ChangedLines(oldContent, newContent)

			model := ctx.Model
			if model == "" {
				if m, ok, _ := store.LoadSessionModel(ctx.SessionID); ok {
					model = m
				}
			}

			var modelPtr *string
			if model != "" {
				modelPtr = &model
			}

			entry := &core.CheckpointEntry{
				SessionID:      ctx.SessionID,
				Tool:           ctx.Tool,
				Model:          modelPtr,
				FilePath:       relative,
				AddedLines:     diffResult.AddedLines,
				TotalAdditions: diffResult.TotalAdditions,
				TotalDeletions: diffResult.TotalDeletions,
				Timestamp:      uint64(time.Now().Unix()),
			}

			logDir, err := store.WorkingLogDir(headSHA)
			if err != nil {
				return err
			}
			if err := core.WriteEntry(logDir, entry); err != nil {
				return err
			}
			store.SavePostSnapshot(relative, newContent)
			store.DeleteSnapshot(relative)
		}
	}

	fmt.Println(`{"continue": true}`)
	return nil
}
