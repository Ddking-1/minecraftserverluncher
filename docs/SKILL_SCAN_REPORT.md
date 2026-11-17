# Skill scan report (NVIDIA SkillSpector)

Scanner: NVIDIA SkillSpector v2.12.0 (installed via `uv tool install git+https://github.com/NVIDIA/SkillSpector.git`).
Agent skill installed: `skill-inspector` from `nvidia/skillspector`.
Scan mode: static only (`--no-llm`), recursive over `.agents/skills/`.
Scanned: 2026-11-17.

## Summary

| Skill | Score | Severity | Recommendation |
|---|---|---|---|
| autonomous-execution | 17 | LOW | SAFE |
| autonomous-loops | 39 | MEDIUM | CAUTION |
| ci-cd-and-automation | 27 | MEDIUM | CAUTION |
| cpp-api-documentation | 0 | LOW | SAFE |
| cpp-benchmark-testing | 17 | LOW | SAFE |
| cpp-fuzz-testing | 17 | LOW | SAFE |
| cpp-mock-testing | 0 | LOW | SAFE |
| cpp-unit-testing | 30 | MEDIUM | CAUTION |
| github-actions-efficiency | 0 | LOW | SAFE |
| go-api-documentation | 0 | LOW | SAFE |
| go-benchmark-testing | 0 | LOW | SAFE |
| go-fuzz-testing | 0 | LOW | SAFE |
| go-unit-testing | 20 | LOW | SAFE |
| golang-continuous-integration | 94 | CRITICAL | DO NOT INSTALL |
| golang-testing | 51 | HIGH | DO NOT INSTALL |
| skill-inspector | 0 | LOW | SAFE |

## Flagged skills (recommended removal)

- golang-continuous-integration (94/100 CRITICAL): hidden instructions in assets/copilot-review-instructions.md (P2), credential access in references/repo-security.md (PE3), unpinned npx references (RP1), skill enumeration (AS3), autonomous decision making (EA2).
- golang-testing (51/100 HIGH): credential access in evals/evals.json (PE3 x3), direct prompt extraction in references/examples.md (P6).

## Action taken

No skills were removed automatically. The two flagged skills remain installed but are documented here as unsafe; user decision required to delete them.
"