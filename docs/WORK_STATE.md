# WORK STATE

Current phase:
PHASE 1 — LAUNCHER BUILD BASELINE (CMake/Qt6) GREEN

Current task:
Fix the launcher CMake/Qt6 build so the C++ launcher core compiles with the
installed Qt 6.6.3 MSVC 2022 toolchain, and remove the OpenSSL dependency from
the HWID helper.

Status:
BUILD BASELINE GREEN (branch fix/launcher-cmake-qt6). CMake configure and
`cmake --build ... --config Release` both succeed and produce
MinecraftLauncher.exe.

Branch:
fix/launcher-cmake-qt6

Last completed step:
- launcher/CMakeLists.txt rewritten: find_package(Qt6 Core Network),
  CMAKE_AUTOMOC ON, C++17, full source list (core/minecraft/launch/steps),
  include dirs, Qt6::Core + Qt6::Network, Windows libs (bcrypt, iphlpapi,
  wbemuuid, ole32, oleaut32).
- Root CMakeLists.txt fixed: removed duplicate add_subdirectory(launcher) and
  the Qt5-first lookup.
- launcher/core/HWIDExtractor.cpp: OpenSSL sha.h replaced by Windows CNG
  (bcrypt.h) SHA-256; input = MachineGuid + CPUID + primary MAC.
- launcher/game/MinecraftLauncher.cpp: invalid `if (!m_process->start())`
  (void return in Qt6) replaced by start() + waitForStarted(5000).

Verification performed: real build on Qt 6.6.3 msvc2019_64, CMake 4.1.1,
MSVC 19.44, generator "Visual Studio 17 2022" x64 — configure OK, Release
build OK, MinecraftLauncher.exe produced.

Current step:
None in progress. This task is finished and stopped.

Next step:
1. Owner: review and merge branch fix/launcher-cmake-qt6 into master.
2. Continue the approved 15-task plan (next: HWID verification / remaining
   launcher build baseline items).
3. Owner: legal sign-off on DECISION-004 (manual action).
4. Resolve DECISION-005 (metadata source; self-hosted mirror recommended) to
   implement library/asset download (currently stubs).
5. Implement OS-protected token storage (Windows DPAPI) per docs/AUTH.md
   (currently first-pass plaintext).

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
