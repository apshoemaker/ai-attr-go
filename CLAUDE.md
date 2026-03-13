# ai-attr

Line-level AI code attribution via native agent hooks. No PATH shimming, no binary interposition.

## Quick start

```bash
go build -o ai-attr ./cmd/ai-attr
go test ./...
```

## Orientation

- [AGENTS.md](./AGENTS.md) -- behavioral rules and progressive disclosure map
- [ARCHITECTURE.md](./ARCHITECTURE.md) -- package structure, data flow, invariants

## Conventions

- TDD: write tests first, implement to pass
- No go-git -- shell out via `os/exec.Command` with `.Dir`
- Git note ref: `refs/notes/ai-attribution`
- Note format: `authorship/3.0.0`
- Go 1.21+
