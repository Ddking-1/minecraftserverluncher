# LICENSE AUDIT

## Purpose

This document tracks the licensing status of every third-party component considered for use.

No component is approved for production use merely because it is open source.

## Required record

For every dependency record:

- name
- repository URL
- exact version/commit inspected
- license
- intended use
- copied code? yes/no
- linked code? yes/no
- external runtime? yes/no
- source distribution obligations
- attribution obligations
- notice requirements
- public repository requirements
- legal review required? yes/no

## Critical project constraints

The central development repository is private.

The launcher will eventually be distributed to players.

The backend may remain private where legally permissible.

A required third-party public source repository must not be hidden merely because the central project repository is private.

## Prism Launcher

Status:
ANALYSIS COMPLETE (architecture level, 2026-09-05) — resolution recorded in
`docs/PRISM_ARCHITECTURE_REVIEW.md` and DECISION-004.
Final owner legal sign-off: still required.

License: GPL-3.0, with incorporated works recorded in upstream `COPYING.md`
(PolyMC GPL-3.0; MultiMC Apache-2.0; `libraries/launcher` GPL-3.0 with
linking exception; libnbt++ LGPL-3.0-or-later; rainbow LGPL-2.1-or-later;
cmark BSD-2; MinGW-w64 runtime notice; Qt 6 LGPL-3; icon/font licenses —
details in `docs/PRISM_ARCHITECTURE_REVIEW.md` section 2).

Resolved conclusions:
- modified fork + project launcher core in one executable = one combined
  GPLv3 work; complete corresponding source must be offered to every
  recipient of the binaries
- distribution model: public launcher source repository; the central
  repository stays private (DECISION-004)
- backend, Velocity plugin and server-side components are outside the
  launcher GPL boundary
- per-release obligations checklist: `docs/PRISM_ARCHITECTURE_REVIEW.md`
  section 3.4

Open audit items:
- exact upstream commit NOT pinned: the local snapshot has no VCS metadata;
  pin a commit before any code copying and record it here
- record individual licenses of vendored `javacheck`, `LocalPeer`,
  `murmur2`, `qdcss` before the first launcher build
- verify presence and license of cmark in the pinned clone

## AGPL components

LimboAPI, LimboAuth and LimboFilter require separate careful licensing analysis.

Do not copy their source into proprietary components casually.

Prefer integration across a clean component boundary when legally and technically appropriate.

## MIT components

Geyser/Floodgate may be used according to their licenses, preserving required copyright/license notices.

## BSD components

WebAuthn libraries with permissive licenses may be considered.

## Decision rule

When legal status is uncertain:
STOP.

Mark:
LEGAL REVIEW REQUIRED

Do not improvise.
