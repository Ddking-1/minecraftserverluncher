# CHANGELOG

## Unreleased

### Project initialization

- Defined requirements for a Windows Minecraft launcher/client ecosystem.
- Defined common account concept.
- Defined separate Java and Bedrock identity paths.
- Defined Velocity as central network gateway.
- Defined trusted launcher and normal Java authentication paths.
- Defined Geyser/Floodgate Bedrock path.
- Defined update, integrity, skin and deployment goals.
- Added persistent AI development governance documents.

No production functionality is claimed at this stage.

### Repository governance baseline

- Added `.gitignore`; `reference/` and local tooling state are excluded from version control (DECISION-002).
- Extended `AGENTS.md` with permanent agent rules: task branches, no force-push anywhere, focused commits, mandatory `git diff` review before commit, mandatory checks, no recursive scanning of `reference/`, interrupted tasks continue on the existing task branch, and full agent ownership of Git/GitHub operations.
- Recorded DECISION-001: the launcher direction is "required Prism core + own launcher core + own backend", not "full Prism Launcher + custom layer".
- Documented component ownership (project-owned launcher core, Go + SQLite backend, project-owned Velocity plugin; Geyser/Floodgate/Limbo as external integrations) in `docs/ARCHITECTURE.md`.
- Synchronized `docs/AI_CONTEXT.md`, `docs/WORK_STATE.md` and `docs/PROGRESS.md` with the real repository state.
