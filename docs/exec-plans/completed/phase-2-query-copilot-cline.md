# Phase 2: Query Commands + Copilot + Cline (Completed)

## Goal

Users can query attribution data with `blame` and `stats`, and ai-attr supports Copilot and Cline agents end-to-end.

## Deliverables

1. `pkg/cli/blame.go` -- line-level file attribution with text and JSON output
2. `pkg/cli/stats.go` -- AI composition statistics across commit ranges
3. `pkg/adapters/copilot.go` -- parse Copilot hook JSON (PreToolUse + PostToolUse, camelCase and snake_case variants)
4. `pkg/adapters/cline.go` -- parse Cline hook JSON (PreToolUse + PostToolUse, taskId-based sessions)
5. `pkg/installers/copilot.go` -- write `.github/hooks/ai-attr.json` with tool matcher
6. `pkg/installers/cline.go` -- write `.clinerules/hooks/ai-attr.json`

## Acceptance criteria (all passed)

- [x] `ai-attr blame <file>` displays per-line AI/human attribution
- [x] `ai-attr blame <file> --json` outputs structured JSON with tool, model, and line metadata
- [x] `ai-attr stats` aggregates attribution across `HEAD~10..HEAD` by default
- [x] `ai-attr stats [range] --json` outputs structured JSON with per-tool breakdowns
- [x] `ai-attr checkpoint copilot --hook-input` parses Copilot PreToolUse/PostToolUse JSON
- [x] `ai-attr checkpoint cline --hook-input` parses Cline PostToolUse JSON with taskId
- [x] `ai-attr install --agents copilot` writes correct `.github/hooks/ai-attr.json`
- [x] `ai-attr install --agents cline` writes correct `.clinerules/hooks/ai-attr.json`
- [x] Copilot adapter supports both camelCase and snake_case field variants
- [x] All new code has tests. `go test ./...` passes.

## Decision log

- Blame uses `git blame --porcelain` and cross-references line numbers against AuthorshipLog attestations
- Blame classifies lines as "ai" (with tool/model) or "human" via `classifyLine()` checking LineRanges
- Stats aggregates PromptRecords by tool name, defaults to `HEAD~10..HEAD` range
- Copilot adapter validates tool names (`insert_edit_into_file`, `create_file`); Cline adapter does not (more general-purpose)
- Copilot installer uses tool matcher regex; Cline installer omits matcher
- Both query commands support `--json` for machine consumption
