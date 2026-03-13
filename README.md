# ai-attr

**Line-level AI code attribution for your git history.**

ai-attr automatically tracks which lines of code were written by AI agents — Claude Code, GitHub Copilot, Codex, Cline — and stores that attribution as git notes. No manual comment markers. No PATH shimming. No binary interposition.

## The problem

Many organizations require developers to mark AI-generated code with comment blocks like `/*START GENAI*/` ... `/*END GENAI*/`. This is error-prone, noisy, and only works when developers remember to do it. Other tools solve this automatically, but they work by shimming the `git` binary on PATH — something enterprise security teams often flag as a supply chain risk.

## How ai-attr works

ai-attr uses each AI agent's **native hook system** to detect edits, and a standard **git post-commit hook** to record the results. No binary interposition required.

```
Developer uses Claude Code to edit src/main.rs
  ↓
Claude's PostToolUse hook fires
  ↓
ai-attr checkpoint claude --hook-input stdin
  → diffs the file against its last known state
  → writes a checkpoint to .git/ai-attr/
  ↓
Developer runs: git commit
  ↓
Post-commit hook fires
  ↓
ai-attr commit
  → consolidates checkpoints into an AuthorshipLog
  → stores it as a git note on refs/notes/ai-attribution
```

The result is a git note on every commit that records exactly which lines were AI-generated, by which agent, in which session.

## Quick start

```bash
# Build
make build

# Install hooks for Claude Code in the current repo
./ai-attr install --agents claude

# That's it. Use Claude Code normally.
# Attribution is recorded automatically on every commit.

# View attribution for the latest commit
./ai-attr show

# View attribution for a specific commit
./ai-attr show abc123
```

## Commands

| Command | Description |
|---------|-------------|
| `ai-attr install --agents <list>` | Set up hooks for specified agents |
| `ai-attr checkpoint <agent>` | Record a checkpoint (called by agent hooks, not manually) |
| `ai-attr commit` | Consolidate checkpoints into a git note (called by post-commit hook) |
| `ai-attr show [commit]` | Display the attribution note for a commit |
| `ai-attr blame <file>` | Line-level AI attribution, like `git blame` |
| `ai-attr stats [range]` | AI composition statistics across a commit range |
| `ai-attr uninstall` | Remove hooks and agent configurations |

## Supported agents

| Agent | Hook mechanism | Granularity | Status |
|-------|---------------|-------------|--------|
| **Claude Code** | `PreToolUse` / `PostToolUse` | Per-edit | Implemented |
| **GitHub Copilot** | `preToolUse` / `postToolUse` (agent mode) | Per-edit | Implemented |
| **Cline** | `PostToolUse` | Per-edit | Implemented |
| **Codex CLI** | `notify` (`agent-turn-complete`) | Per-turn | Implemented |

> **Note:** Copilot's inline tab completions cannot be tracked — there is no hook for accepting a completion. Only agent mode edits are captured.

## What the output looks like

`ai-attr show HEAD` displays the attribution note:

```
src/main.rs
  a1b2c3d4e5f6a7b8 1-10,15-20
src/lib.rs
  a1b2c3d4e5f6a7b8 1-50
  9f8e7d6c5b4a3210 55-60
---
{
  "schema_version": "authorship/3.0.0",
  "tool_version": "ai-attr/0.2.0",
  "prompts": {
    "a1b2c3d4e5f6a7b8": {
      "agent_id": {
        "tool": "claude",
        "model": "sonnet-4",
        "id": "session-abc123"
      },
      "human_author": "alice",
      "total_additions": 60,
      "total_deletions": 5,
      "accepted_lines": 55
    }
  }
}
```

Lines 1-10 and 15-20 of `src/main.rs` were written by Claude (sonnet-4) in session `session-abc123`, invoked by `alice`.

## Design principles

| Principle | Details |
|-----------|---------|
| **No PATH shimming** | Uses native agent hooks, not binary interposition |
| **Git-native storage** | Attribution stored as git notes — travels with the repo |
| **Agent-agnostic core** | One binary, agent differences isolated behind `AgentAdapter` interface |
| **Automatic** | No manual markers — hooks capture attribution transparently |

## Building from source

Requires Go 1.21+.

```bash
git clone <repo-url>
cd ai-attr
make build
# Binary at ./ai-attr
```

## Make targets

| Target | Description |
|--------|-------------|
| `make build` | Build the `ai-attr` binary |
| `make test` | Run all tests |
| `make test-v` | Run all tests with verbose output |
| `make vet` | Run `go vet` |
| `make clean` | Remove the built binary |
| `make install` | Build and copy binary to `$GOPATH/bin` |

## Running tests

```bash
make test
```

Tests create temporary git repos and don't require network access.

## Project status

ai-attr is under active development. All core functionality is implemented with 96 passing tests: checkpoint recording, commit consolidation, install/uninstall, blame, stats, and adapters for Claude Code, GitHub Copilot, Cline, Codex, and a generic adapter.

See [docs/PLANS.md](./docs/PLANS.md) for the full roadmap.

## License

[MIT](./LICENSE)
