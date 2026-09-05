# DEVELOPMENT WORKFLOW

## Core principle

The project is developed by multiple AI agents and therefore needs persistent, auditable state.

## Every task

1. Read AGENTS.md.
2. Read docs/AI_CONTEXT.md.
3. Read docs/WORK_STATE.md.
4. Read docs/PROGRESS.md.
5. Read relevant architecture documents.
6. Inspect git status.
7. Create a new branch.
8. Define a small task.
9. Implement only that task.
10. Run relevant tests/checks.
11. Update documentation.
12. Update WORK_STATE.
13. Update PROGRESS.
14. Commit.
15. Push.
16. Report results.
17. STOP.

## Branch naming

Preferred:
- architecture/*
- review/*
- feature/*
- fix/*
- security/*
- infra/*
- docs/*

## Commit size

Prefer small, logically isolated commits.

Do not combine:
- authentication
- launcher UI
- deployment
- updater

into one giant commit.

## Interruption recovery

If a task is interrupted:
- inspect git status
- inspect last commit
- read WORK_STATE
- read PROGRESS
- inspect current diff
- continue only the unfinished part

Never start the same task from scratch if existing work is recoverable.

## Handoff

Every agent must leave enough information for the next agent to continue without guessing.

## Manual action reporting

A task must explicitly say:
- no manual action required
OR
- exact action required

Examples:
"Merge branch X into main."
"Grant GitHub permission."
"Create DNS record."
"Install Java 24."
"Run command Y."
