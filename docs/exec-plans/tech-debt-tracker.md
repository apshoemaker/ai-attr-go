# Tech Debt Tracker

Known gaps, shortcuts, and deferred work.

| Item | Severity | Phase | Notes |
|------|----------|-------|-------|
| No rebase attribution tracking | High | Future | Deferred entirely. Notes won't survive `git rebase`. |
| Copilot inline completions untracked | Medium | N/A | No hook system exists for tab completions. VS Code extension needed (post-MVP). |
| Codex attribution is turn-level, not edit-level | Low | 3 | Diffs entire working tree per turn. Human+AI edits between turns may be imprecise. |
| `cli/show` has no `--json` flag | Low | 2 | Currently outputs raw note text only. |
| No `.gitignore` for `.git/ai-attr/` | None | N/A | Directory is inside `.git/`, already invisible to git. |
