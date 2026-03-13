# Phase 0: Scaffold (Completed)

## Goal

Go project with CLI skeleton, storage layer, serialization format, and git notes -- all tested.

## Deliverables (all done)

1. Go module with Cobra CLI skeleton, all subcommand stubs
2. `pkg/storage/repostorage.go` -- `.git/ai-attr/` directory management
3. `pkg/storage/gitnotes.go` -- read/write notes via `git notes --ref=ai-attribution`
4. `pkg/core/linerange.go` -- `LineRange` struct with compress/expand/remove
5. `pkg/core/serialization.go` -- `authorship/3.0.0` format
6. `pkg/cli/show.go` -- working `ai-attr show [commit]` command
7. Adapter interface + implementations for claude, copilot, codex, cline, generic
8. Installer implementations for all agents + git hooks

## Acceptance criteria (all passed)

- [x] `go build ./cmd/ai-attr` succeeds
- [x] `go test ./...` passes (96 tests)
- [x] `ai-attr --help` shows all subcommands
- [x] `ai-attr show` works in a git repo
- [x] AuthorshipLog round-trips through serialize/deserialize
- [x] Git notes round-trip through add/show
- [x] `GenerateShortHash` produces deterministic 16-char hex hashes

## Decision log

- Used `refs/notes/ai-attribution` as the notes ref
- All git commands use `exec.Command().Dir` instead of changing process cwd (parallel-safe tests)
- 16-char hex hashes from SHA256 truncation
- Custom `AiAttrError` type with error kinds instead of sentinel errors
- Cobra for CLI framework
- Histogram diff algorithm (custom implementation, not Myers)
