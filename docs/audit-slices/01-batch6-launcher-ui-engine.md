# SLICE 1 — BATCH 6: Launcher C++ / Qt6 UI & Modpack Engine

Security posture for this slice: treat all local config, downloaded manifest,
skin, modpack and asset data as UNTRUSTED input until validated. Memory safety
and RAII are mandatory. No code is "safe" merely because it compiles or tests
green.

## Goals (each item ends with a green build/test + commit)

- B6.1 `EditorConfig` C++ class wired to Qt UI forms — settings panel
  (RAM / resolution) saves to the local config file.
- B6.2 `AuthClient` C++ Qt signal/slot wired to a login dialog.
- B6.3 `SkinClient` C++ Qt widget — async skin preview + upload button.
- B6.4 Async asset downloader: SHA-256 verifying download class.
- B6.5 Async asset downloader: buffered / multi-threaded HTTP worker pool
  (QtNetwork; no new heavy deps).
- B6.6 Async asset downloader: progress event callbacks + retry logic.
- B6.7 Game launcher engine: JRE/JDK auto-detection module.
- B6.8 Game launcher engine: classpath builder + native library extractor.
- B6.9 Game launcher engine: process launcher (QProcess, RAM allocation,
  session token passing — token never on an untrusted/visible path).
- B6.10 CMake / CTest integration for the full B6 module; build + CTest green.

## Hard rules (non-negotiable)

1. Micro-delegate isolated chunks to `lmstudio5`
   (http://localhost:1234/v1/chat/completions, model `qwen2.5-coder-7b-instruct`).
2. 8-attempt local auto-fix loop with `lmstudio5`; Architect fallback after 8.
3. Never weaken a security control to make a test pass (NO TEST CHEATING).
4. `EditorConfig`, `SkinClient`, `AuthClient`, `FileIntegrity` already exist in
   `launcher/core` and `launcher/auth` — reuse, do not duplicate.
5. Prefer `std::unique_ptr`, stack ownership, RAII. `shared_ptr` only when
   ownership is genuinely shared. No raw owning pointers.
6. Classpath and native extraction: prevent path traversal, symlink escape,
   and arbitrary writes. Verify SHA-256 before use.
7. Game launcher: do NOT let user-controlled input lead to arbitrary local
   program execution beyond the intended Java launch.
8. Session token to the JVM must not be readable via process command line in a
   way that leaks it (consider env or protected args).
9. Verify `go vet` (if Go touched), CMake build, and `ctest` green before each
   commit; add a CTest target for the new module.

## Verification gate for the slice

- CMake configure + Release build clean.
- `ctest --test-dir launcher/build -C Release --output-on-failure` 100% green.
- No new dependency added without documenting license/version.
- No hardcoded secret anywhere in committed code.
