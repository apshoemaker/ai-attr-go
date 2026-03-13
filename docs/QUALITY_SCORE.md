# Quality Score

Quality grades per module. Updated when phases complete.

## Grading

- **A**: Fully implemented, tested, documented, no known issues
- **B**: Implemented and tested, minor gaps
- **C**: Partially implemented or stubbed, tests exist for implemented parts
- **D**: Stub only, no real implementation
- **F**: Missing or broken

## Current grades

| Module | Grade | Notes |
|--------|-------|-------|
| `core/linerange` | A | Full implementation, 8 tests |
| `core/serialization` | A | authorship/3.0.0 round-trip, 7 tests |
| `core/diff` | A | Histogram diff algorithm, 6 tests |
| `core/workinglog` | A | JSON checkpoint entries, auto-increment, 4 tests |
| `core/blame` | A | git blame porcelain parser, 3 tests |
| `storage/repostorage` | A | Directory management, snapshots, path resolution, 11 tests |
| `storage/gitnotes` | A | Notes CRUD + authorship round-trip, 7 tests |
| `cli/show` | B | Works, but no --json flag yet |
| `cli/checkpoint` | A | PreToolUse snapshot + PostToolUse diff, 5 tests |
| `cli/commit` | A | Two-phase diff, consolidates checkpoints into git note, 6 tests |
| `cli/install` | A | Dispatches to agent installers, 2 tests |
| `cli/blame` | A | Text + JSON output, cross-references git blame |
| `cli/stats` | A | Aggregates across commit ranges |
| `cli/uninstall` | A | Removes hooks, settings, ai-attr dir, 3 tests |
| `adapters/claude` | A | Parses PreToolUse/PostToolUse, validates edit tools, 8 tests |
| `adapters/copilot` | A | camelCase + snake_case support, 5 tests |
| `adapters/codex` | A | Notify events, whole-tree diff strategy, 2 tests |
| `adapters/cline` | A | taskId as session_id, 2 tests |
| `adapters/generic` | A | Minimal JSON format, required field validation, 3 tests |
| `installers/githooks` | A | Post-commit hook with idempotency, 4 tests |
| `installers/claude` | A | settings.json with deduplication, 4 tests |
| `installers/copilot` | B | .github/hooks/ai-attr.json, 1 test |
| `installers/cline` | B | .clinerules/hooks/ai-attr.json, 1 test |
| `installers/codex` | B | ~/.codex/config.toml, 1 test |

## Summary

96 tests total across all modules. All core functionality is implemented and tested.
