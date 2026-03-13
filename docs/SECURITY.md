# Security

## Threat model

ai-attr runs as a local CLI tool invoked by git hooks and agent hooks. It has no network access, no daemon process, and no privileged operations.

### What ai-attr does NOT do

- Modify PATH or create symlinks posing as `git`
- Intercept, wrap, or proxy git commands
- Send data to any remote service
- Run with elevated privileges
- Access files outside the repository working tree and `.git/`

### Attack surface

| Surface | Risk | Mitigation |
|---------|------|------------|
| Hook input (stdin JSON) | Malformed JSON from a compromised agent | Parsed with `encoding/json`; invalid input returns an error, no execution |
| Git notes content | Tampered attribution data | Notes are signed by the same trust model as git commits. ai-attr does not add cryptographic signatures (future consideration) |
| `.git/ai-attr/` working logs | Local file manipulation | Lives inside `.git/`, not accessible to unprivileged users who can't already access the repo |
| Post-commit hook | Execution as current user | Standard git hook execution model. No privilege escalation |
| Installer writes | Config file modification | Append-only. Installers do not delete existing configuration |

### Supply chain

ai-attr is distributed as a single Go binary. Dependencies are locked via `go.sum`. No runtime downloads, no auto-updates, no telemetry. The only external dependency is `github.com/spf13/cobra` for CLI parsing.

### Future considerations

- Cryptographic signing of attribution notes (GPG or SSH signatures)
- Verification of note integrity on `ai-attr show`/`ai-attr blame`
- CI integration to validate attribution completeness
