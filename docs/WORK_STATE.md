﻿# WORK STATE

Current phase:
PHASE 1 — BACKEND SCAFFOLD / AUTH API + LAUNCHER CORE SCAFFOLD

Current task:
Launcher core scaffold (C++ / Qt): minimal instance management, launch-task
framework, and backend auth client.

Status:
LAUNCHER CORE SCAFFOLD IMPLEMENTED (branch launcher/core-scaffold).
Compilation BLOCKED: no Qt 6 toolchain is installed in this environment.

Branch:
launcher/core-scaffold

Last completed step:
Implemented the minimal C++ launcher core under launcher/: core/ (BaseInstance,
InstanceList as a QAbstractListModel, InstancePersistence for instance.json /
launcher.json / version.json), minecraft/ (MinecraftInstance, PackProfile),
auth/ (BackendAuthClient matching server/api/handlers.go, TokenStorage,
SessionManager — replaces Prism's Mojang/MSA flow), and launch/ (LaunchTask /
LaunchStep framework with ClaimAuthSession, ResolveLibraries, ResolveAssets,
ExtractNatives, VerifyJava, LauncherPartLaunch steps and Java version helpers).
Added launcher/CMakeLists.txt (Qt6), a minimal main.cpp, and launcher/README.md.
The minecraft and launch subsystems are built as one library because
MinecraftInstance::createLaunchTask assembles the launch steps and the steps
operate on MinecraftInstance (mutual dependency).

Verification performed: CMake configure dry-run reaches the Qt find_package
step cleanly (CMake structure and MSVC toolchain detection valid; only failure
is the missing Qt 6); the C++ auth client request/response contract was
verified line-by-line against server/api/handlers.go; all project-header
#include paths verified to resolve; manual review for incomplete-type and
upcast issues. The C++ code was NOT compiled because no Qt toolchain is
installed.

Current step:
None in progress. This task is finished and stopped.

Next step:
1. Install Qt 6 (Core/Gui/Widgets/Network) and compile/verify the launcher
   core (manual action).
2. Owner: review and merge branch launcher/core-scaffold into master.
3. Owner: legal sign-off on DECISION-004 (manual action).
4. Resolve DECISION-005 (metadata source; self-hosted mirror recommended) to
   implement library/asset download (currently stubs).
5. Implement OS-protected token storage (Windows DPAPI) per docs/AUTH.md
   (currently first-pass plaintext).
6. Launcher UI and remaining launcher core design tasks.

Repository facts:
- default branch: master
- remote: origin (private GitHub repo Ddking-1/minecraftserverluncher)
- reference/ is local-only, never committed (DECISION-002)
- backend source exists under server/ (Go module minecraftserverluncher/server)
- launcher core scaffold exists under launcher/ (C++ / Qt, not yet compiled)
- velocity directory is scaffolded but empty (no code yet)

Last commit:
See docs/PROGRESS.md — launcher core scaffold commit on branch
launcher/core-scaffold.
