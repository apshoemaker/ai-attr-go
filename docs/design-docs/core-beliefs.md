# Core Beliefs

Foundational principles that guide ai-attr's design. When in doubt, defer to these.

## 1. Attribution must be automatic

Manual markers (`/*START GENAI*/`) fail because they depend on developer discipline. ai-attr exists to remove humans from the attribution loop. If a developer has to remember to do something for attribution to work, we've failed.

## 2. No supply chain risk

Enterprise security teams must be able to adopt this tool without exception requests. That means:
- No PATH manipulation
- No binary interposition
- No symlinks posing as system tools
- Auditable single-binary distribution

The moment ai-attr requires a security exception, it loses to "just use comment markers."

## 3. Git-native storage

Attribution lives in git notes, not databases, not SaaS, not sidecar files. This means:
- Works offline
- Travels with the repo (`git push`/`fetch` notes refs)
- Composes with existing git workflows (blame, log, bisect)
- No infrastructure to provision or maintain

## 4. Agent-agnostic core, agent-specific edges

The core attribution engine (diff, track, serialize) is identical for all agents. Agent-specific code is confined to two thin layers: adapters (parse hook JSON) and installers (write config files). Adding a new agent should never require changing core logic.

## 5. Stable, well-specified format

ai-attr uses the `authorship/3.0.0` note format — a documented, stable schema stored as plain git notes. This makes attribution data readable by any tool without ai-attr installed.

## 6. Incremental correctness over completeness

It's better to accurately attribute 80% of AI-generated code than to guess at 100%. Known gaps (Copilot tab completions, Codex inter-turn edits) are documented as limitations rather than papered over with heuristics.
