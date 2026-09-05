﻿# WORK STATE

Current phase:
PHASE 1 — BACKEND SCAFFOLD / AUTH API

Current task:
Backend scaffold and core auth API (Go + SQLite)

Status:
BACKEND SCAFFOLD AND AUTH API IMPLEMENTED (branch backend/scaffold-and-auth)

Branch:
backend/scaffold-and-auth

Last completed step:
Created the repository scaffold directories (launcher/core, launcher/auth,
launcher/minecraft, launcher/updates, server/api, server/auth, server/sessions,
server/devices, server/database, velocity/plugin, tests, docs). Implemented the
Go + SQLite auth backend under server/: environment-driven config (no
hardcoded secrets), idempotent SQLite schema, account model with bcrypt
password hashing, server-side session model with expiry and revocation,
device identity mapping, and registration/login/refresh/logout/logout-all
endpoints with per-IP rate limiting. Added minimal tests (auth, database,
API integration). See server/README.md.

Current step:
None in progress. This task is finished and stopped.

Next step:
1. Owner: review and merge branch backend/scaffold-and-auth into master.
2. Owner: legal sign-off on DECISION-004 (manual action).
3. Resolve DECISION-005 (metadata source; self-hosted mirror recommended).
4. Launcher bootstrap task: pin the Prism upstream commit, create the public
   launcher fork repository, record vendored library licenses, produce the
   trim plan.
5. Launcher core design task (accounts/sessions/devices, launch
   authorization, server profiles, updater/integrity, RAM algorithm, UI
   concept).

Repository facts:
- default branch: master
- remote: origin (private GitHub repo Ddking-1/minecraftserverluncher)
- reference/ is local-only, never committed (DECISION-002)
- backend source exists under server/ (Go module minecraftserverluncher/server)
- launcher and velocity directories are scaffolded but empty (no code yet)

Last commit:
See docs/PROGRESS.md — backend scaffold and auth API commit on branch
backend/scaffold-and-auth.
