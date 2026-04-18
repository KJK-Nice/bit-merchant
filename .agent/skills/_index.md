# Skill Registry

Read this file first. Full `SKILL.md` contents load only when a skill's
triggers match the current task. Machine-readable equivalent:
`skills/_manifest.jsonl`.

## skillforge
Creates new skills from observed patterns and recurring tasks.
Triggers: "create skill", "new skill", "I keep doing this manually"

## memory-manager
Reads, scores, and consolidates memory. Runs reflection cycles.
Triggers: "reflect", "what did I learn", "compress memory"

## deploy-checklist
Pre-deployment verification against a structured checklist.
Triggers: "deploy", "ship", "release", "go live"
Constraints: all tests passing, no unresolved TODOs in diff,
requires human approval for production.

---

## Project-specific skills (loaded by Claude Code)

These live under `.claude/skills/` and are authoritative for this project:

- **gitnexus/gitnexus-exploring** — understand architecture, trace flows
- **gitnexus/gitnexus-impact-analysis** — blast radius before edits
- **gitnexus/gitnexus-debugging** — trace bugs via the call graph
- **gitnexus/gitnexus-refactoring** — safe rename/extract/split
- **gitnexus/gitnexus-pr-review** — review PRs with graph context
- **gitnexus/gitnexus-cli** — index, status, clean, wiki CLI
- **gitnexus/gitnexus-guide** — tools/resources/schema reference
- **threedotslab** — Go CQRS/DDD/Clean Architecture audit & scaffold

Git operations: Claude Code's built-in git safety rules apply — never
force-push, stage specific files, prefer new commits over amend.
