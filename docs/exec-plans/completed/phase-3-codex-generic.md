# Phase 3: Codex + Generic Adapter (Completed)

## Goal

ai-attr supports Codex CLI and any third-party agent via a generic adapter.

## Deliverables

1. `pkg/adapters/codex.go` -- parse Codex `agent_turn_complete` notify events
2. `pkg/adapters/generic.go` -- factory-constructed adapter for arbitrary agents with strict field validation
3. `pkg/installers/codex.go` -- append `notify` line to `~/.codex/config.toml`

## Acceptance criteria (all passed)

- [x] `ai-attr checkpoint codex --hook-input` parses Codex notify events
- [x] Codex adapter maps all events to PostToolUse (no per-file hooks available)
- [x] Codex adapter returns empty FilePath (whole-tree diff strategy)
- [x] Generic adapter requires `phase` and `session_id` fields
- [x] Generic adapter accepts both full names (`PreToolUse`) and shorthand (`pre`, `post`)
- [x] `ai-attr install --agents codex` writes correct `~/.codex/config.toml` notify entry
- [x] Codex installer is idempotent (detects existing `ai-attr` lines and updates in place)
- [x] All new code has tests. `go test ./...` passes.

## Decision log

- Codex provides only turn-level events, not per-file edit hooks -- attribution diffs the entire working tree per turn, which may mix human and AI edits between turns
- Codex adapter supports dual field names (`event_type`/`type`, `session_id`/`thread_id`) for API flexibility
- Generic adapter uses factory pattern (`NewGenericAdapter(name)`) so custom agent names are set at construction
- Codex installer manipulates `config.toml` textually (no TOML parser) -- finds and replaces existing ai-attr lines or appends
