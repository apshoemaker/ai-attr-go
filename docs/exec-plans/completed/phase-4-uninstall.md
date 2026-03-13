# Phase 4: Uninstall (Completed)

## Goal

Users can cleanly remove all ai-attr hooks, configurations, and working state with a single command.

## Deliverables

1. `pkg/cli/uninstall.go` -- orchestrates full removal
2. `pkg/installers/githooks.go` (`UninstallGitHook`) -- removes ai-attr lines from `.git/hooks/post-commit`
3. `pkg/installers/claude.go` (`UninstallClaude`) -- removes ai-attr hook entries from `.claude/settings.json`
4. Removal of `.git/ai-attr/` directory (checkpoints, snapshots, working logs)

## Acceptance criteria (all passed)

- [x] `ai-attr uninstall` removes ai-attr lines from post-commit hook
- [x] `ai-attr uninstall` preserves non-ai-attr lines in post-commit hook
- [x] `ai-attr uninstall` deletes entire hook file if only ai-attr content remains
- [x] `ai-attr uninstall` removes ai-attr entries from `.claude/settings.json`
- [x] `ai-attr uninstall` preserves non-ai-attr entries in `.claude/settings.json`
- [x] `ai-attr uninstall` removes `.git/ai-attr/` directory
- [x] `ai-attr uninstall` is idempotent (no error if already uninstalled)
- [x] All new code has tests. `go test ./...` passes.

## Decision log

- Git hook removal uses marker-based detection (`# ai-attr post-commit hook`) with two-line skip (marker + command)
- Claude settings removal uses JSON filtering -- iterates hook phases and removes entries where command contains `ai-attr`
- All uninstall operations are no-ops if target files don't exist
- Uninstall preserves non-ai-attr content in shared configuration files
- `.git/ai-attr/` is removed entirely with `os.RemoveAll` -- git notes on `refs/notes/ai-attribution` are intentionally preserved (they are part of repo history)
