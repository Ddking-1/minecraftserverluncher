﻿# WORK STATE

Current phase:
PHASE 0 — PROJECT GOVERNANCE / ARCHITECTURE

Current task:
Prism minimal-core architecture and licensing boundary resolution

Status:
PRISM MINIMAL-CORE CHECKPOINT RESOLVED (architecture level)

Branch:
architecture/prism-minimal-core

Last completed step:
Targeted analysis of reference/prism-launcher (instance/version handling,
Fabric/loader handling, Java selection, libraries/assets download, launch
process). Component-level KEEP / KEEP BUT ADAPT / REPLACE / NOT NEEDED
inventory produced in docs/PRISM_MINIMAL_CORE.md. Licensing boundary resolved
in docs/PRISM_ARCHITECTURE_REVIEW.md. DECISION-003 (minimal-core boundary,
approved), DECISION-004 (GPLv3 distribution / public launcher repository,
approved direction pending legal sign-off) and DECISION-005 (metadata source,
PENDING) recorded.

Current step:
None in progress. This task is finished and stopped.

Next step:
1. Owner: legal sign-off on DECISION-004 (manual action).
2. Resolve DECISION-005 (metadata source; self-hosted mirror recommended).
3. Launcher bootstrap task: pin the Prism upstream commit (the local snapshot
   has no VCS metadata), create the public launcher fork repository, record
   vendored library licenses in docs/LICENSE_AUDIT.md, produce the trim plan.
4. Launcher core design task (accounts/sessions/devices, launch
   authorization, server profiles, updater/integrity, RAM algorithm, UI
   concept).

Production implementation stays blocked until DECISION-004 sign-off and
DECISION-005 resolution.

Repository facts:
- default branch: master
- remote: origin (private GitHub repo Ddking-1/minecraftserverluncher)
- reference/ is local-only, never committed (DECISION-002)
- no production source code exists yet

Last commit:
See docs/PROGRESS.md — Prism minimal-core architecture commit on branch
architecture/prism-minimal-core.
