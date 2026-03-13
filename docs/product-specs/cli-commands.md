# CLI Commands

## `ai-attr checkpoint <agent>`

Called by agent hooks. Records a diff checkpoint.

```
ai-attr checkpoint claude --hook-input
ai-attr checkpoint copilot --hook-input
```

- `<agent>`: One of `claude`, `copilot`, `codex`, `cline`
- `--hook-input`: Read agent hook JSON from stdin

## `ai-attr commit`

Called by post-commit hook. Consolidates checkpoints into a git note.

```
ai-attr commit
```

No arguments. Reads checkpoints from `.git/ai-attr/working_logs/<previous-HEAD>/`, builds an `AuthorshipLog`, writes it as a git note on the new commit.

## `ai-attr install`

Sets up git hook and agent configurations.

```
ai-attr install --agents claude
ai-attr install --agents claude,copilot
```

- `--agents`: Comma-separated list of agents to configure

## `ai-attr show [commit]`

Displays the raw attribution note.

```
ai-attr show          # shows HEAD
ai-attr show abc123   # shows specific commit
```

## `ai-attr blame <file>`

Displays line-level AI attribution for a file, similar to `git blame`.

```
ai-attr blame src/main.rs
ai-attr blame src/main.rs --json
```

- `--json`: Output as JSON

## `ai-attr stats [range]`

Displays AI composition statistics.

```
ai-attr stats                    # HEAD only
ai-attr stats HEAD~10..HEAD      # range
ai-attr stats --json             # JSON output
```

- `--json`: Output as JSON

## `ai-attr uninstall`

Removes hooks and agent configurations.

```
ai-attr uninstall
```
