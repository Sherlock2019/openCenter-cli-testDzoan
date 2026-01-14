I couldn’t find an `AGENTS.md` in `/Users/victor.palma/projects/rackspace/openCenter-cli`, so I’m proceeding without repo-specific instructions.

Here are design ideas to make bootstrap resumable, non-redundant, and controllable:

**State + Idempotency**
- Track step state in a durable store (local file, cluster metadata, or both): `pending | running | success | failed | skipped`, with timestamps and versions.
- Make each step idempotent and self-validating (precheck determines if it’s already done), so “resume” is safe.

**Step Registry**
- Define a step registry with IDs, dependencies, and optional “precheck” and “rollback” hooks.
- Let the runner compute an execution plan based on registry + saved state.

**CLI Controls**
- `--resume` (default): skip completed steps based on state/precheck.
- `--restart` or `--force` : ignore saved state, rerun all steps.
- `--step <id>`: execute a single step (and optionally its prerequisites).
- `--from-step <id>`: rerun from a step onward, marking subsequent steps as pending.
- `--until <id>`: stop after a step.
- `--reset-to <id>`: clear state for step and downstream steps.

**Plan + Preview**
- `--plan` or `--dry-run`: show ordered steps with “run/skip” decisions and reasons.
- Emit a “why skipped” message (precheck matched or state says success).

**Failure Handling**
- On failure, persist state and error details; `--resume` continues from the failing step.
- Add `--retry <n>` and `--backoff` for transient failures.

**Data Model Sketch**
- `bootstrap-state.json` (or cluster annotation) with:
  - `version` of bootstrap schema
  - `steps`: map of `id -> status + last_run + result_hash`
  - `inputs_hash` to detect config changes and invalidate steps

**Dependency-aware Execution**
- If a step’s inputs change (config hash, cluster version), mark it and downstream steps as stale.
- Use a DAG to avoid redoing unrelated steps.

**UX Details**
- Show a “checkpoint” summary at the end with next suggested command: `bootstrap --resume`
- Provide `bootstrap status` to show progress and failures.

Questions to refine:
- Where should state live (local file vs cluster metadata)?
- Do you need multi-host coordination (e.g., multiple operators can run bootstrap)?
- Any strict ordering or parallelism requirements?

If you want, I can draft the CLI flags + step registry interface based on your current command layout.
