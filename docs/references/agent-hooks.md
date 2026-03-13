# Agent Hook Reference

Each agent has a different hook system. This document describes the hook payload format and integration pattern for each supported agent.

## Claude Code

**Hook type:** `PreToolUse` + `PostToolUse` in `.claude/settings.json`

**Tool matcher:** `Edit|Write|MultiEdit|CreateFile`

**Flow:**
1. `PreToolUse` fires before an edit -- ai-attr snapshots the file
2. `PostToolUse` fires after the edit -- ai-attr diffs the snapshot against the new content
3. Checkpoint written to `.git/ai-attr/working_logs/<HEAD>/`

**Hook input (stdin JSON):**
```json
{
  "session_id": "...",
  "transcript_path": "~/.claude/projects/.../transcript.jsonl",
  "tool_name": "Edit",
  "tool_input": {
    "file_path": "/absolute/path/to/file.rs"
  }
}
```

**Adapter:** `pkg/adapters/claude.go`
**Installer:** `pkg/installers/claude.go`

## GitHub Copilot

**Hook type:** `preToolUse` + `postToolUse` in `.github/hooks/ai-attr.json`

**Tool matcher:** `insert_edit_into_file`, `create_file`

**Limitation:** Agent mode only. Inline tab completions have no hook system.

**Adapter:** `pkg/adapters/copilot.go`
**Installer:** `pkg/installers/copilot.go`

## Cline

**Hook type:** `PostToolUse` in `.clinerules/hooks/`

**Session ID:** Uses `taskId` from the hook payload.

**Adapter:** `pkg/adapters/cline.go`
**Installer:** `pkg/installers/cline.go`

## Codex CLI

**Hook type:** `notify` event (`agent-turn-complete`) in `~/.codex/config.toml`

**Strategy:** No per-file-edit events. On each turn completion, ai-attr diffs the entire working tree against the last checkpoint.

**Limitation:** Turn-level granularity. Human+Codex edits between turns may be attributed imprecisely.

**Adapter:** `pkg/adapters/codex.go`
**Installer:** `pkg/installers/codex.go`

## Adding a new agent

1. Create `pkg/adapters/{agent}.go` implementing `AgentAdapter`
2. Create `pkg/installers/{agent}.go` with an `Install...()` function
3. Add the agent name to `pkg/cli/install.go` and `pkg/adapters/adapter.go` dispatch
4. Document the hook format in this file
5. Add tests
