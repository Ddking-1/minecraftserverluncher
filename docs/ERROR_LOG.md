# ERROR LOG

This file records tasks that could not be completed within the retry budget
(8 Worker attempts). A blocked task is recorded here, marked as blocked in
`docs/MASTER_PLAN_EXTENDED.md`, and the marathon continues with the next task.

Format:

## YYYY-MM-DD — TASK ID / TITLE

- Branch:
- Symptom:
- Build/check output (excerpt):
- Worker attempts: N
- Suspected cause:
- Next action required:

---

## 2026-10-04 — WORKER TOOL-USE LIMITATION (not a product blocker)

- Branch: n/a
- Symptom: The Worker model
  (`lm-studio/deepseek-coder-v2-lite-instruct`) cannot invoke tools. When asked
  to write files it emits pseudo-JSON / markdown code fences as text instead of
  calling the `write` tool. It also frequently invents non-existent APIs
  (e.g. `Load().Validate()`, `database.Database`) and emits truncated output.
- Build/check output (excerpt): repeated `not enough arguments`, `use of
  undefined type`, `redeclared` compile errors from Worker output; each fixed
  by the Architect restating the exact signature/constraint.
- Worker attempts: n/a (ongoing characteristic)
- Suspected cause: model capability; the role split assigns file writing, build
  and Git to the Architect.
- Next action required: none. Architect transcribes Worker output verbatim
  (after review) and runs builds/tests. This is expected workflow, not a
  blocked task.

---

## 2026-10-05 — WORKER MODEL BENCHMARKS (lmstudio3 14B vs lmstudio5 7B)

- Branch: n/a
- Symptom / result: Two candidate local worker models were benchmarked over
  the LM Studio REST API (`http://localhost:1234/v1/chat/completions`).
- `lmstudio3` (`qwen2.5-coder-14b-instruct`, IQ4_XS): first-token latency
  11–20s even for trivial prompts; a 40-token JSON completion took ~283s; a
  200-token Go helper and a 250-token HTML component both exceeded a 300s
  timeout. It followed a JSON *shape* but not the exact value contract
  (`model:"test"` → returned `"Qwen"`). **REJECTED** for the worker role.
- `lmstudio5` (`qwen2.5-coder-7b-instruct-abliterated`, Q6_K_L): ping 2.56s;
  JSON contract `{"ok":true,"model":"test"}` exact (3.25s, 78 tokens); Go
  helper correct (10.0s, 105 completion tokens, ~10.5 tok/s); HTML button
  correct (13.0s, 140 completion tokens, ~10.8 tok/s). No timeouts. **ACCEPTED.**
- Build/check output (excerpt): both code generations were syntactically valid
  and matched the requested contract.
- Worker attempts: n/a (benchmark, not a task retry)
- Suspected cause: the 14B IQ4_XS quantization is too slow for this runtime;
  the 7B Q6_K_L is a good latency/size trade-off.
- Next action required: `lmstudio5` (`qwen2.5-coder-7b-instruct`) is now the
  local worker model for grunt-work (HTML/CSS/JSON/boilerplate). The previous
  `deepseek-coder-v2-lite-instruct` remains usable as a fallback.

---

(no blocked product tasks recorded yet)
