# Design

Design principles, format specifications, and key decisions for ai-attr.

## Git Note Format: authorship/3.0.0

Stored under `refs/notes/ai-attribution`.

### Structure

A note has two sections separated by a `---` divider:

```
<attestation section>
---
<JSON metadata section>
```

### Attestation section

Each file with AI-attributed lines gets a block:

```
<file_path>
  <hash> <line_ranges>
  <hash> <line_ranges>
```

- **file_path**: Repo-relative POSIX path. Quoted with `"` if it contains spaces.
- **hash**: 16-character hex string identifying the AI session (see Hash Generation).
- **line_ranges**: Comma-separated, no spaces. Single lines as `N`, ranges as `N-M` (inclusive). Sorted ascending.

Example:

```
src/main.rs
  a1b2c3d4e5f6a7b8 1-10,15-20
  9f8e7d6c5b4a3210 25,30-35
"path with spaces/file.rs"
  a1b2c3d4e5f6a7b8 1-5
```

### Metadata section (JSON)

```json
{
  "schema_version": "authorship/3.0.0",
  "tool_version": "ai-attr/0.2.0",
  "base_commit_sha": "abc123def456...",
  "prompts": {
    "a1b2c3d4e5f6a7b8": {
      "agent_id": {
        "tool": "claude",
        "model": "sonnet-4",
        "id": "session-uuid"
      },
      "human_author": "alice",
      "total_additions": 30,
      "total_deletions": 5,
      "accepted_lines": 25
    }
  }
}
```

| Field | Type | Description |
|-------|------|-------------|
| `schema_version` | string | Always `"authorship/3.0.0"` |
| `tool_version` | string | `"ai-attr/{version}"` |
| `base_commit_sha` | string | HEAD when checkpoints were recorded |
| `prompts` | map | Keyed by 16-char hash. One entry per AI session |

| PromptRecord field | Type | Description |
|--------------------|------|-------------|
| `agent_id.tool` | string | `claude`, `copilot`, `codex`, `cline` |
| `agent_id.model` | string | Model name (e.g., `sonnet-4`) |
| `agent_id.id` | string | Session identifier from the agent |
| `human_author` | string? | Git user.name of the invoking human |
| `total_additions` | uint32 | Lines added by this session |
| `total_deletions` | uint32 | Lines deleted by this session |
| `accepted_lines` | uint32 | Lines surviving to commit |

### Hash generation

```
hash = SHA256("{tool}:{agent_id}")[0..16]   // first 16 hex chars
```

Deterministic. Same input always produces the same hash.

### Reading notes

```bash
git notes --ref=ai-attribution show HEAD
```

## Key Design Decisions

### Why native hooks, not PATH shimming?

PATH shimming (symlinking a binary as `git`) intercepts every git operation. Enterprise security teams flag this as a supply chain risk. ai-attr avoids this entirely by using each agent's own hook system.

**Trade-off:** We can only track agents that have hooks. Copilot inline tab completions are invisible (no hook exists). This is an acceptable gap -- agent mode is where most substantial AI editing happens.

### Why git notes, not a database?

Git notes travel with the repo. No external infrastructure. `git clone --mirror` preserves them. They can be pushed/fetched like any other ref. They compose naturally with existing git workflows.

**Trade-off:** Notes don't survive `git rebase` without explicit handling. Rebase-aware attribution is deferred to a future phase.

### Why one binary, not per-agent tools?

A single `ai-attr` binary handles all agents. This gives users one thing to install and one set of commands to learn. Agent differences are isolated behind the `AgentAdapter` interface.

### Why `.git/ai-attr/`?

Storing state inside `.git/` means it's invisible to `git status` and never committed. The `ai-attr` prefix makes the purpose clear.
