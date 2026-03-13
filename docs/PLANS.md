# Plans

Execution plans live in `exec-plans/active/` while in progress and move to `exec-plans/completed/` when done.

## Current phase: Complete through Phase 4

All core phases are implemented.

## Phase overview

| Phase | Scope | Plan | Status |
|-------|-------|------|--------|
| 0 | Scaffold, storage, serialization, git notes | [completed/phase-0-scaffold.md](./exec-plans/completed/phase-0-scaffold.md) | Done |
| 1 | Claude Code end-to-end | [completed/phase-1-claude-e2e.md](./exec-plans/completed/phase-1-claude-e2e.md) | Done |
| 2 | Query commands + Copilot + Cline | [completed/phase-2-query-copilot-cline.md](./exec-plans/completed/phase-2-query-copilot-cline.md) | Done |
| 3 | Codex + Generic adapter | [completed/phase-3-codex-generic.md](./exec-plans/completed/phase-3-codex-generic.md) | Done |
| 4 | Uninstall | [completed/phase-4-uninstall.md](./exec-plans/completed/phase-4-uninstall.md) | Done |

## Planning conventions

- One plan per phase, named `phase-N-{slug}.md`
- Plans include: goal, deliverables, acceptance criteria, decision log
- Move to `completed/` when all acceptance criteria pass
- Tech debt discovered during execution goes in [tech-debt-tracker.md](./exec-plans/tech-debt-tracker.md)
