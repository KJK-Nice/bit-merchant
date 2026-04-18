# Project Instructions (Claude Code)

This project uses the **agentic-stack** portable brain. All memory, skills,
and protocols live in `.agent/`.

## Before doing anything
1. Read `.agent/AGENTS.md` — it's the map.
2. Read `.agent/memory/personal/PREFERENCES.md` — how the user works.
3. Read `.agent/memory/semantic/LESSONS.md` — what we've learned.
4. Read `.agent/protocols/permissions.md` — what you can and cannot do.

## While working
- Consult `.agent/skills/_index.md` and load the full `SKILL.md` for any
  skill whose triggers match the task.
- Update `.agent/memory/working/WORKSPACE.md` as the task evolves.
- Log significant actions to `.agent/memory/episodic/AGENT_LEARNINGS.jsonl`
  via `.agent/tools/memory_reflect.py`.

## Rules that override defaults
- Never force push to `main`, `production`, or `staging`.
- Never delete episodic or semantic memory entries — archive them.
- Never modify `.agent/protocols/permissions.md`.

---

<!-- gitnexus:start -->
## GitNexus — Code Intelligence

This repo is indexed by GitNexus (bit-merchant). Use the GitNexus MCP tools
to navigate safely. If any tool warns the index is stale, run
`npx gitnexus analyze`.

**Must do:**
- Before editing a function/class/method, run
  `gitnexus_impact({target: "X", direction: "upstream"})` and report the
  blast radius. Warn on HIGH/CRITICAL risk before proceeding.
- Before committing, run `gitnexus_detect_changes()` to verify scope.
- For renames, use `gitnexus_rename({symbol_name, new_name, dry_run: true})`
  first — never find-and-replace.

**Must not do:**
- Never edit a symbol without first running `gitnexus_impact`.
- Never ignore HIGH/CRITICAL risk warnings.
- Never commit without `gitnexus_detect_changes()`.

**Tool cheatsheet** — `query` (find by concept), `context` (360° view of a
symbol), `impact` (blast radius), `detect_changes` (pre-commit scope check),
`rename` (safe multi-file rename), `cypher` (custom graph queries).

**Risk depth** — `d=1` WILL BREAK (must update), `d=2` LIKELY AFFECTED
(test), `d=3` MAY NEED TESTING (test if critical path).

**Skills** — detailed workflows live in `.claude/skills/gitnexus/*/SKILL.md`
(exploring, impact-analysis, debugging, refactoring, pr-review, cli, guide).

> A user-level PostToolUse hook re-runs `npx gitnexus analyze` after
> `git commit` / `git merge`. Preserve embeddings with `--embeddings` when
> running manually (see `.gitnexus/meta.json` → `stats.embeddings`).
<!-- gitnexus:end -->
