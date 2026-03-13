# AGENTS.md

Read [ARCHITECTURE.md](./ARCHITECTURE.md) before making structural changes.

## Rules

1. **TDD.** Write a failing test before implementation. Run `go test ./...` after every change.
2. **No go-git.** All git ops use `os/exec.Command` with `.Dir = repoDir`. Never change process cwd.
3. **No stubs in released code.** All phases (0-4) are implemented. Future extensions should follow the same TDD pattern.
4. **Format compatibility.** Git notes must round-trip with `authorship/3.0.0`. See [docs/DESIGN.md](./docs/DESIGN.md).
5. **One adapter per agent.** `pkg/adapters/{agent}.go` implements `AgentAdapter`. Do not merge adapters.
6. **Installers separate from adapters.** `pkg/installers/` writes config. `pkg/adapters/` parses payloads.
7. **Conventional commits.** `feat:`, `fix:`, `test:`, `docs:`, `refactor:`.

## Progressive disclosure

| Question | Location |
|----------|----------|
| How does the codebase fit together? | [ARCHITECTURE.md](./ARCHITECTURE.md) |
| Design decisions and format spec | [docs/DESIGN.md](./docs/DESIGN.md) |
| Core beliefs / why this tool exists | [docs/design-docs/core-beliefs.md](./docs/design-docs/core-beliefs.md) |
| How does a specific agent hook work? | [docs/references/agent-hooks.md](./docs/references/agent-hooks.md) |
| Execution plans (all phases complete) | [docs/PLANS.md](./docs/PLANS.md) |
| Completed plans | [docs/exec-plans/completed/](./docs/exec-plans/completed/) |
| Tech debt and known gaps | [docs/exec-plans/tech-debt-tracker.md](./docs/exec-plans/tech-debt-tracker.md) |
| Quality grades by module | [docs/QUALITY_SCORE.md](./docs/QUALITY_SCORE.md) |
| Security constraints | [docs/SECURITY.md](./docs/SECURITY.md) |
| Product specs | [docs/product-specs/index.md](./docs/product-specs/index.md) |

## Testing

```bash
go test ./...                  # everything
go test ./pkg/core/...         # pure logic (fast, no I/O)
go test ./pkg/storage/...      # git integration tests
```

Git notes tests create real repos via `t.TempDir()`. No network access required.
