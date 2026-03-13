# Phase 1: Claude Code End-to-End (Completed)

## Goal

A user can install ai-attr, use Claude Code to edit files, commit, and see correct line-level attribution in the git note.

## Deliverables

1. `pkg/adapters/claude.go` -- parse Claude hook JSON (PreToolUse + PostToolUse)
2. `pkg/core/diff.go` -- Histogram diff for line-level diffing
3. `pkg/core/workinglog.go` -- checkpoint data structure and persistence
4. `pkg/cli/checkpoint.go` -- read hook input, diff, write checkpoint to `.git/ai-attr/`
5. `pkg/cli/commit.go` -- consolidate checkpoints into `AuthorshipLog`, write git note
6. `pkg/installers/claude.go` -- add hook entries to `.claude/settings.json`
7. `pkg/installers/githooks.go` -- write post-commit hook to `.git/hooks/post-commit`
8. E2E test: Claude edits file -> checkpoint -> commit -> note -> `ai-attr show HEAD`

## Acceptance criteria (all passed)

- [x] `ai-attr install --agents claude` writes correct `.claude/settings.json` hooks
- [x] `ai-attr install --agents claude` writes correct `.git/hooks/post-commit`
- [x] `ai-attr checkpoint claude --hook-input` parses Claude PostToolUse JSON
- [x] `ai-attr checkpoint claude --hook-input` writes a checkpoint file
- [x] `ai-attr commit` reads checkpoints and writes a valid git note
- [x] `ai-attr show HEAD` displays the note with correct file paths and line ranges
- [x] `git notes --ref=ai-attribution show HEAD` produces valid `authorship/3.0.0` output
- [x] All new code has tests. `go test ./...` passes.
