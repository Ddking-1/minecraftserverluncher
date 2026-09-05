# PROJECT AGENTS / PERMANENT DEVELOPMENT RULES

## Purpose

This repository contains the complete development project for a private Minecraft network client ecosystem.

The project may contain:
- custom Minecraft launcher/client
- authentication backend
- Velocity integration/plugin
- Geyser/Floodgate integration
- Minecraft instance/update logic
- skin integration
- web/admin backend
- deployment tooling
- tests
- documentation

## Absolute repository rules

The default/integration branch of this repository is currently `master`.
Every rule below that refers to `main` applies to that default branch.

1. NEVER work directly on `main` (this repository: `master`).
2. ALWAYS create a dedicated task branch before modifying code.
3. NEVER force-push. Not `main`, not shared branches, not task branches.
4. NEVER delete another agent's branch.
5. NEVER rewrite unrelated subsystems during a task.
6. NEVER silently change an approved architecture decision.
7. NEVER copy third-party source code into the product without a documented license decision.
8. `reference/` is READ-ONLY reference material.
9. Third-party code must never be copied merely because it is convenient.
10. Every meaningful task must update documentation.
11. Every task must update `docs/AI_CONTEXT.md` when project state changes.
12. Every task must update `docs/PROGRESS.md`.
13. Every task must update `docs/WORK_STATE.md`.
14. Every completed task must update `docs/CHANGELOG.md` when appropriate.
15. Every meaningful task must end with a Git commit.
16. Completed work must be pushed to the remote branch.
17. Agents must report:
   - branch
   - commit
   - files changed
   - tests/checks
   - unresolved issues
   - manual user actions required
18. If the task is incomplete, document the exact next step.
19. Never pretend that a task is complete if it is not.
20. Never remove security checks merely to make a test pass.
21. ALWAYS inspect `git status` and the full `git diff` (staged and unstaged) before committing.
22. Commits must be focused: one logical task per commit, no unrelated files, no bulk reformatting.
23. Run the checks appropriate to the changed area before committing (build, tests, linters, schema/doc consistency). If no check exists for the change, say so explicitly in the report.
24. NEVER recursively scan, index, or bulk-read the third-party repositories under `reference/`. Open only the specific files needed, identified by a targeted search.
25. If a task is interrupted, CONTINUE ON THE EXISTING TASK BRANCH. Do not create a second branch for the same task and do not restart the work from scratch.
26. The AI agent is responsible for all Git/GitHub management for its task: branch creation, commits, push, and remote/repository handling. The user is not asked to perform these operations.

## Git / GitHub ownership rule

Repository management is agent work, not user work.

For every task the agent must:
1. verify the current branch and working tree state
2. create or resume the correct task branch
3. keep `reference/` and local tooling state out of version control
4. review the complete diff before committing
5. create focused commits with descriptive messages
6. push the task branch to `origin`
7. report branch, commit hash and pushed state truthfully

The agent must never ask the user to create branches, commit, or push.

The agent must report a push as successful ONLY if the push command actually succeeded. A failed or unauthenticated push must be reported as a blocker with the exact manual action required.

## Stop rule

Agents MUST stop when the requested task is complete.

They MUST NOT automatically continue to the next feature.

A task is complete only when:
- requested files/functions are finished
- relevant tests/checks are complete
- documentation is updated
- progress state is updated
- Git commit exists
- branch has been pushed

## Architecture rule

Requirements are authoritative unless a later approved architecture decision supersedes them.

If a coding agent discovers that the approved architecture is technically impossible, unsafe, or legally problematic:

STOP IMPLEMENTATION.

Document the problem and create:
`docs/DECISIONS.md`

Do not silently redesign the system.

## Security rule

The client is never the final security authority.

Server-side authorization is authoritative.

Never trust:
- username alone
- client brand
- arbitrary client metadata
- local files alone
- local configuration alone

## Third-party rule

The following are external reference repositories:

- reference/prism-launcher
- reference/geyser
- reference/floodgate
- reference/limboapi
- reference/limboauth
- reference/limbofilter
- reference/limboauth-clientmod
- reference/webauthn

They are READ-ONLY.

They are NOT tracked in this Git repository (see `.gitignore` and `docs/DECISIONS.md` DECISION-002). Each agent is expected to have them available locally; they must never be committed.

Do not scan them recursively. Use targeted searches and open only the specific files required.

Before copying or modifying third-party code:
1. identify exact repository
2. identify exact version/commit
3. identify license
4. determine whether copying is legally permitted
5. document the decision
6. preserve required notices
7. determine whether public source distribution is required

## GitHub rule

The existing private GitHub repository is the central development repository.

It may contain:
- launcher
- backend
- Velocity plugin/integration
- deployment
- tests
- schemas
- docs

If a third-party license or distribution model requires a separate public repository, the agent must document that requirement and create/manage the required public repository when the available GitHub integration permits it.

Do not create unnecessary repositories.

## Documentation rule

Documentation is part of the implementation.

A future AI agent must be able to understand the current project state by reading:
- AGENTS.md
- docs/AI_CONTEXT.md
- docs/WORK_STATE.md
- docs/PROGRESS.md
- docs/DECISIONS.md

## User interaction rule

Agents must explicitly tell the user when manual action is required, including:
- GitHub authorization
- branch merge
- DNS setup
- secrets
- environment variables
- server credentials
- certificate setup
- local software installation
- acceptance of a license
- publishing a public repository

Never imply that an external action succeeded unless it was actually verified.

## Development priorities

Prefer:
1. correctness
2. security
3. maintainability
4. legal/licensing compliance
5. low resource usage
6. implementation simplicity
7. performance
8. visual polish

Do not optimize prematurely.

## Resource constraint

The project must remain lightweight.

Avoid unnecessary:
- microservices
- distributed systems
- Redis
- message queues
- heavyweight web frameworks
- background polling
- unnecessary full-disk hashing
- unnecessary network requests

Unless architecture explicitly proves them necessary.

## AI efficiency

Agents must avoid reading the entire repository repeatedly when a focused file set is sufficient.

Prefer:
- relevant docs
- relevant source files
- focused diffs
- tests
- targeted reference source inspection

Before large-scale work, identify the smallest sufficient context.

## Final response format

Every agent task completion report should contain:

STATUS:
BRANCH:
COMMIT:
FILES:
TESTS:
SECURITY NOTES:
DOCUMENTATION:
MANUAL USER ACTIONS:
NEXT TASK:

Never omit the branch and commit.
