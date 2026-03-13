# Architecture

This document describes the high-level architecture of ai-attr. If you want to familiarize yourself with the codebase, you are in the right place.

See also [docs/DESIGN.md](./docs/DESIGN.md) for format specs and design decisions, and [docs/references/agent-hooks.md](./docs/references/agent-hooks.md) for per-agent hook details.

## Bird's Eye View

ai-attr records which lines of code were written by AI agents and stores that attribution as git notes. It does this without intercepting or shimming the `git` binary -- instead it uses each agent's native hook system to detect edits, and a standard git post-commit hook to consolidate the results.

The system has two phases of operation:

**Edit time** (checkpoint): An AI agent edits a file. The agent's hook fires and calls `ai-attr checkpoint <agent>`. ai-attr diffs the working tree against its last snapshot, computes line-level attribution, and writes the result to `.git/ai-attr/working_logs/<HEAD>/`.

**Commit time** (post-commit): The git post-commit hook calls `ai-attr commit`. ai-attr reads all checkpoints accumulated since the last commit, merges them into an `AuthorshipLog`, and stores it as a git note on `refs/notes/ai-attribution`.

```
Agent edits file
  -> Agent's native hook fires
  -> ai-attr checkpoint <agent> --hook-input stdin
  -> Diffs working tree against last snapshot
  -> Writes checkpoint to .git/ai-attr/working_logs/<HEAD>/entry-N.json

git commit
  -> post-commit hook fires
  -> ai-attr commit
  -> Reads checkpoints, builds AuthorshipLog
  -> Stores as git note on refs/notes/ai-attribution
```

**Architecture Invariant:** ai-attr never interposes on the `git` binary. It never modifies PATH, creates symlinks to itself as `git`, or wraps git commands. All git interaction is through `os/exec.Command` calling the real `git`.

## Layers

```
pkg/core/          Pure types and algorithms (no I/O)
  |
pkg/storage/       Filesystem + git operations (I/O boundary)
  |
pkg/adapters/      Agent-specific hook payload parsing (stateless)
  |
pkg/cli/           Subcommand handlers (orchestration)
  |
cmd/ai-attr/       Cobra dispatch (no logic)
```

**Architecture Invariant:** Dependencies flow downward only. `core/` imports nothing from other layers. `storage/` imports from `core/` but not `cli/` or `adapters/`. `cli/` may import from all lower layers.

## Code Map

### `cmd/ai-attr/main.go`

Cobra CLI entry point. Defines subcommands (`checkpoint`, `commit`, `install`, `blame`, `show`, `stats`, `uninstall`) and dispatches to `pkg/cli/` handlers. Contains no business logic.

### `pkg/errors/errors.go`

`AiAttrError` struct with `ErrorKind`. All fallible functions return `error`.

### `pkg/core/`

Pure data types and algorithms. No I/O, no git commands, no filesystem access.

#### `pkg/core/diff.go`

Line-level diffing using a Histogram diff algorithm. `ChangedLines(old, new)` returns a `DiffResult` containing 1-indexed line numbers of added/modified lines, plus total addition/deletion counts.

#### `pkg/core/workinglog.go`

`CheckpointEntry` struct (serializable to JSON) recording which lines an AI agent touched in a single file edit. `WriteEntry(logDir, entry)` writes auto-incrementing `entry-N.json` files. `ReadEntries(logDir)` reads all entries sorted by index.

#### `pkg/core/linerange.go`

The `LineRange` struct: `{Start, End uint32}` (1-indexed, inclusive). Start==End for single lines. Provides `Contains`, `Overlaps`, `Remove`, `CompressLines` (sorted `[]uint32` -> minimal `[]LineRange`), and `Expand`.

This is the fundamental unit of attribution: "lines 10-20 were written by agent X."

**Architecture Invariant:** Line numbers are 1-indexed and inclusive on both ends. `LineRange{5, 5}` represents a single line.

#### `pkg/core/serialization.go`

The `authorship/3.0.0` git note format. Key types:

- `AuthorshipLog` -- top-level container: attestations (file -> hash -> line ranges) + metadata (JSON with prompts map)
- `AuthorshipMetadata` -- schema version, tool version, base commit, prompts map
- `FileAttestation` -- per-file list of `AttestationEntry` (hash + line ranges)
- `PromptRecord` -- agent identity, human author, line stats
- `AgentId` -- tool name, session id, model name
- `GenerateShortHash(agentID, tool)` -- SHA256-based 16-char identifier

Serialization: text attestation section + `---` divider + JSON metadata. Full spec in [docs/DESIGN.md](./docs/DESIGN.md).

**Architecture Invariant:** `GenerateShortHash` uses the format `"{tool}:{agent_id}"` hashed with SHA256 and truncated to 16 hex characters. This must remain stable across versions.

### `pkg/storage/`

Filesystem and git operations. This is the I/O boundary.

#### `pkg/storage/repostorage.go`

Manages the `.git/ai-attr/` directory:

```
.git/ai-attr/
  working_logs/
    <commit-sha>/
      entry-0.json
      entry-1.json
      ...
  snapshots/
    <sha256-of-relative-path>   (pre-edit file content)
  post_snapshots/
    <sha256-of-relative-path>   (post-AI-edit file content)
```

`NewRepoStorage(gitDir, workdir)` creates the directory structure. `WorkingLogDir(sha)` returns the path for a specific commit's checkpoints. `SaveSnapshot` / `LoadSnapshot` / `DeleteSnapshot` manage pre-edit file content for diffing. `ListWorkingLogDirs()` returns all working log directories for commit consolidation.

Also provides `FindGitDir(start)` which walks up from a path to find `.git/`.

**Architecture Invariant:** The `.git/ai-attr/` directory is inside `.git/`, so it is never committed and never visible to `git status`. Working logs are keyed by the HEAD commit SHA at the time of the checkpoint.

#### `pkg/storage/gitnotes.go`

Read/write git notes on `refs/notes/ai-attribution`:

- `NotesAdd(repoDir, sha, content)` -- writes a note (force-overwrites)
- `NotesShow(repoDir, sha)` -- reads a note, returns `(content, ok, err)`
- `GetAuthorship(repoDir, sha)` -- reads + deserializes to `*AuthorshipLog`
- `NotesRemove(repoDir, sha)` -- removes a note (ignores errors)
- `GetHeadSHA(repoDir)`, `GetUserName(repoDir)`, `FindRepoRoot(start)`

All functions take `repoDir string` and use `exec.Command("git").Dir = repoDir`.

**Architecture Invariant:** Git commands never rely on the process working directory. Every git invocation sets `.Dir` explicitly. This makes the code safe for concurrent test execution.

### `pkg/cli/`

One file per subcommand. Each exports a `Run...()` function called from `main.go`.

- `checkpoint.go` -- parse hook input, snapshot/diff, write checkpoint entry
- `commit.go` -- consolidate checkpoints into git note on HEAD
- `install.go` -- set up post-commit hook and agent configs
- `show.go` -- display raw attribution note
- `blame.go` -- line-level file attribution (text and JSON output)
- `stats.go` -- AI composition statistics across commit ranges
- `uninstall.go` -- remove hooks, agent configs, and `.git/ai-attr/`

### `pkg/adapters/`

One file per supported agent. Each implements the `AgentAdapter` interface:

```go
type AgentAdapter interface {
    ParseHookInput(input string) (*AgentContext, error)
    Name() string
}
```

The interface converts agent-specific hook JSON into a common `AgentContext` struct.

- `claude.go` -- Claude Code `PreToolUse`/`PostToolUse` hooks
- `copilot.go` -- GitHub Copilot agent mode hooks
- `codex.go` -- Codex CLI `notify` events (working-tree diff strategy)
- `cline.go` -- Cline `PostToolUse` hooks
- `generic.go` -- fallback for unknown agents

**Architecture Invariant:** Adapters are stateless. They parse a JSON string and return structured data. They do not touch the filesystem, call git, or maintain state between invocations.

### `pkg/installers/`

One file per agent + `githooks.go`. Installers write configuration files to enable hook integration:

- `claude.go` -- writes `.claude/settings.json` hook entries
- `copilot.go` -- writes `.github/hooks/ai-attr.json`
- `codex.go` -- writes `~/.codex/config.toml` notify entry
- `cline.go` -- writes `.clinerules/hooks/` configuration
- `githooks.go` -- writes `.git/hooks/post-commit`

**Architecture Invariant:** Installers are append-only. They add ai-attr configuration without removing existing hook entries. Uninstallation is a separate operation.

## Cross-Cutting Concerns

### Testing

Tests live in `_test.go` files alongside implementation. Integration tests that need a real git repo create one via `t.TempDir()` and initialize it with `git init` + an initial commit.

All git-interacting tests pass `repoDir` explicitly. No test changes the process working directory.

```bash
go test ./...                  # everything
go test ./pkg/core/...         # pure logic tests (fast, no I/O)
go test ./pkg/storage/...      # git integration tests
```

### Error handling

Public functions return `error`. The `AiAttrError` type provides structured error kinds. CLI handlers in `main.go` print the error to stderr and exit with code 1.

### Git notes ref

ai-attr stores attribution on `refs/notes/ai-attribution`. The note format is `authorship/3.0.0`.
