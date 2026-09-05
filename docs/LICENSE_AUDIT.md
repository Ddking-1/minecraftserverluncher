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
LEGAL/ARCHITECTURAL REVIEW REQUIRED

Reason:
Prism Launcher is GPL-licensed and may impose source-distribution requirements if modified and distributed.

Do not assume:
"private GitHub repo" = "no GPL obligations".

Before using Prism as a distributed launcher foundation:
- inspect exact license
- determine whether modified source must be published
- determine whether proprietary parts would form a combined work
- determine whether a separate/public source repository is needed
- determine required notices
- determine required source availability

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
