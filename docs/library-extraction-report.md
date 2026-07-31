# Library Extraction Report

Date: 2026-07-31

## 1. Summary

No runtime package was extracted in this phase. The repository already has specific domain and infrastructure packages under `internal/`, while every remaining extraction candidate still mixes distinct policy, lifecycle, compatibility, or ordering concerns. Creating another package would make those differences less visible and would fail the required stable-responsibility test.

The implemented work is documentation-only: important package boundaries were clarified, repository architecture and navigation maps were added, and a stale reference in the audit report was rewritten so the command documentation-drift guard can evaluate the post-deduplication tree. Runtime behavior, imports, dependencies, and exported APIs are unchanged.

## 2. Current test baseline

The baseline was captured from the post-Prompt-3 `main` worktree with Go 1.26.5.

| Check | Baseline result |
| --- | --- |
| `go test ./...` | Failed only in `cmd`: `TestDocsDoNotUseRemovedGACommands` found a stale literal reference in `docs/refactor-audit-report.md`. All other default packages passed. |
| `go vet ./...` | Passed. |
| Standalone `staticcheck` | Not installed. |
| `golangci-lint run` | Failed with 102 existing findings: 50 `errcheck`, 2 `ineffassign`, and 50 `staticcheck`. |
| `go mod tidy -diff` | Passed with no module-file diff. |

Repository state matched the expected post-Prompt-3 phase: the latest semantic-deduplication change centralized SOPS overlay file selection inside `internal/sops`, and the audit, dead-code, and deduplication reports were present. The existing worktree was already ahead of its upstream branch; this phase did not rewrite or commit that history.

## 3. Package extraction plan

The approved plan was to evaluate the strongest remaining candidates and proceed only if all required package criteria were met.

| Candidate | Evidence for reuse | Failed criteria and decision |
| --- | --- | --- |
| SOPS command workflow | Several commands coordinate encryption-related steps. | Prompt, output, key-source, file replacement, and path policies differ. A small stable API and independent behavior contract are not yet available. Deferred. |
| Drift orchestration | Command and operations code both coordinate detection and reporting. | Goroutine lifecycle, callback sequence, persistence shape, and provider behavior are not sufficiently characterized. Deferred. |
| Atomic file writing | Several packages write or replace files. | Mode preservation, durability, rollback, and error policy differ. A common API would need policy switches and reduce call-site clarity. Deferred. |
| Retry classification | Multiple integrations recognize transient failures. | Classification is integration-specific and has no common lower-level policy. Deferred. |

The resulting plan was intentionally limited to package comments and repository documentation. The design is recorded in `docs/superpowers/specs/2026-07-31-library-boundary-documentation-design.md`, and the implementation plan is in `docs/superpowers/plans/2026-07-31-library-boundary-documentation.md`.

## 4. Implemented package changes

No package extraction or source relocation was implemented.

Boundary documentation changed as follows:

| Location | Change | Maintainability benefit |
| --- | --- | --- |
| `cmd/doc.go` | Defines commands as the Cobra, user-interaction, presentation, and orchestration layer. | Future changes have a clear rule for keeping reusable behavior out of commands. |
| `internal/di/doc.go` | Identifies `App` and `NewApp` as canonical, with reflection wiring retained only for compatibility. | New dependencies have one discoverable wiring path and lower layers are warned not to import DI. |
| `internal/config/v2/doc.go` | Documents ownership of typed schema processing and its order-sensitive stages. | Configuration changes can be placed and reviewed against their compatibility boundary. |
| `internal/sops/doc.go` | Documents terminal-free SOPS responsibilities, command separation, and observable file ordering. | Agents are less likely to merge distinct key, prompt, and file policies into an unsafe abstraction. |

Because there was no extraction, there are no old-to-new source locations, dependency changes, or new package-level tests to enumerate. This is the preferred outcome under the final rule: no extraction is safer than a policy-heavy package with an unstable contract.

## 5. Package APIs added or changed

None. No exported or private runtime API was added, removed, renamed, or moved. No import path changed. The new Go files contain package comments only.

## 6. Documentation added or updated

- Added `docs/architecture.md` with the system overview, commands, packages, dependency direction, runtime and configuration flows, data flow, integrations, generated boundaries, test strategy, and operational concerns.
- Added `docs/llm-code-map.md` with entrypoints, wiring, package ownership, key types, workflows, extension guidance, generated and dynamic-reference warnings, test commands, and safe-refactor rules.
- Added the approved design and detailed implementation plan under `docs/superpowers/`.
- Updated important Go package comments in `cmd`, `internal/di`, `internal/config/v2`, and `internal/sops`.
- Reworded one stale removed-command reference in `docs/refactor-audit-report.md` without changing the audit conclusion.
- Added this report as `docs/library-extraction-report.md`.

## 7. Tests added or updated

No test source was added or changed because runtime behavior and API shape did not change. Existing tests provide the relevant checks:

- Command tests validate root wiring and documentation drift.
- Package tests exercise DI, v2 configuration, and SOPS behavior.
- The complete default suite detects cross-package compile or behavior regressions.

The documentation-drift test is specifically relevant because it scans Markdown for removed CLI syntax; repairing the stale audit reference turns the only baseline test failure into a valid post-change check.

## 8. Test and lint results

Validation used Go 1.26.5, matching the baseline toolchain.

| Check | Final result |
| --- | --- |
| `gofmt` on changed Go files | Passed; no formatting changes remain. |
| `git diff --check` | Passed. |
| `go test ./cmd ./internal/di ./internal/config/v2 ./internal/sops` | Passed. This includes the command documentation-drift guard. |
| `go test ./...` | Passed for every default package. |
| `go vet ./...` | Passed. |
| Standalone `staticcheck ./...` | Not run because the binary is not installed. |
| `golangci-lint run` | Failed with the unchanged baseline of 102 pre-existing findings: 50 `errcheck`, 2 `ineffassign`, and 50 `staticcheck`. None points to a file changed by this phase. |
| `go mod tidy -diff` | Passed with no module-file diff. |

The sole baseline test failure was resolved by removing the stale removed-command literal from the audit report. No production behavior was changed to make the tests pass.

## 9. Public API compatibility notes

External behavior is unchanged. This phase did not modify:

- Commands, arguments, flags, prompts, output, error mapping, or exit codes.
- Exported Go identifiers or package import paths.
- YAML or JSON configuration shape, defaults, normalization, references, or validation.
- Filesystem layout, generated output, embedded asset names, or file modes.
- Plugin discovery, executable conventions, metadata, environment, or checksums.
- Provider requests, secret handling, retries, concurrency, or persistence.

Although Go packages under `internal` are not importable outside the module's parent tree, their serialized data and runtime behavior still participate in the CLI's effective public API.

## 10. Risks and intentionally deferred work

- SOPS workflow extraction remains risky until prompt behavior, output, key-source precedence, replacement semantics, permissions, and exact file order have characterization coverage.
- Drift orchestration remains deferred until cancellation, goroutine ownership, callback order, provider boundaries, and persisted report shape are explicit contracts.
- Atomic-write consolidation remains deferred because callers do not yet share durability, rollback, and file-mode policy.
- Retry consolidation remains deferred because providers classify transient failures differently.
- Typed and reflection-based DI coexist. The typed graph is canonical, but deleting compatibility wiring requires resolving dynamic names and callers first.
- Derived schema and command-reference generation are not fully trustworthy: some `.mise.toml` generation paths are stale or missing, and command-reference output can depend on host plugins.
- The repository has substantial pre-existing lint debt. This documentation-only phase does not broaden scope into unrelated production-code cleanup.

## 11. Recommended next steps

1. Add characterization tests around SOPS command prompts, key sources, replacement behavior, permissions, output, and ordered file selection before reconsidering an orchestration boundary.
2. Add lifecycle and cancellation tests around drift detection, scheduled execution, callbacks, and persistence before moving orchestration ownership.
3. Repair and verify schema, fixture, and command-reference generation tasks; document reproducible inputs and expected diffs.
4. Continue migration from reflection-based DI only when each named compatibility consumer has a typed replacement.
5. Address lint debt in small, behavior-preserving batches with separate baselines and focused tests.
6. Add package documentation to other important packages when their responsibility is non-obvious, without using documentation work as a reason to split packages.
