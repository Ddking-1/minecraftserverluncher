﻿# WORK STATE

Current phase:
PHASE 0 — PROJECT GOVERNANCE / ARCHITECTURE

Current task:
Repository/project-state synchronization (governance only)

Status:
GOVERNANCE BASELINE COMMITTED

Branch:
architecture/project-state

Last completed step:
Governance and project-state documentation synchronized with the approved
direction ("required Prism core + own launcher core + own backend"),
permanent agent rules extended in AGENTS.md, `reference/` excluded from
version control, DECISION-001 and DECISION-002 recorded.

Current step:
None in progress. This task is finished and stopped.

Next step:
Architecture/licensing checkpoint:
1. resolve the Prism Launcher GPL analysis in docs/LICENSE_AUDIT.md
   (distribution model, modified-source publication, public repository need)
2. produce the inventory of Prism functionality that must be RETAINED
3. define the launcher core boundary (own layer vs forked Prism code)
4. record the results as new decisions in docs/DECISIONS.md

No production implementation is approved until that checkpoint is decided.

Repository facts:
- default branch: master
- remote: origin (private GitHub repo Ddking-1/minecraftserverluncher)
- reference/ is local-only, never committed (DECISION-002)
- no production source code exists yet

Last commit:
See docs/PROGRESS.md — governance baseline commit on branch
architecture/project-state.
