# Deduplication Report

## 1. Summary

This phase implemented one low-risk semantic deduplication in `internal/sops`: the ordered selection of overlay files that require encryption. `DefaultSOPSManager` and `GitIntegrator` contained byte-identical private methods with the same domain meaning and caller expectations. They now call one private, typed package function.

Nine other candidate groups were re-evaluated and left unchanged. Their differences involve path precedence, whitespace policy, filesystem durability, retry safety, reflection semantics, compatibility layering, CLI ownership, or architectural scope. Unifying them in this phase would either change behavior or create an abstraction that is less clear than the duplication.

No additional dead code was removed. No exported API, package dependency, error text, log message, metric, label, configuration key, route, serialization format, or CLI output changed.

## 2. Current test baseline

Preflight was performed from commit `3ecb3ae` (`chore: remove unused test`) on `main`, which is the post-Prompt-2 state described by `docs/dead-code-removal-report.md`. The Prompt 1 and Prompt 2 reports were present as untracked documentation inputs, and the Prompt 2 test change was already committed.

Validation used Go 1.26.5 with an explicit matching `GOROOT` and `PATH`.

| Baseline check | Result before deduplication |
|---|---|
| `go test ./...` | Failed only in `cmd` at `TestDocsDoNotUseRemovedGACommands`. The required Prompt 1 audit report contains historical retired-command wording at line 288. All other packages passed. |
| `go vet ./...` | Passed. |
| `staticcheck ./...` | Standalone executable unavailable. |
| `golangci-lint run` | Reported 102 existing findings: 50 `errcheck`, 2 `ineffassign`, and 50 `staticcheck`. |
| `go mod tidy -diff` | Clean in the Prompt 2 report. |

The duplicate candidates were re-checked against the post-Prompt-2 source rather than accepted mechanically from the audit. Dead-code removal did not make the audit-listed broad candidates safer to merge. It did expose a smaller, stronger package-local candidate: the two identical SOPS overlay file-selection methods.

## 3. Deduplication candidate matrix

| Candidate name | Current locations | Domain concept | Differences between implementations | Proposed shared location | Proposed API shape | Why the abstraction is better | Risks | Tests to add or update | Decision |
|---|---|---|---|---|---|---|---|---|---|
| SOPS overlay encryption file selection | Before: `internal/sops/manager.go:438-464` and `internal/sops/git.go:546-572`; callers in both files | Select the ordered overlay-relative files that SOPS should encrypt for a provider | None in returned values, ordering, or provider branching. Both accepted an unused overlay-path argument. | `internal/sops/overlay_files.go` | Private typed function `overlayFilesToEncrypt(*v2.Config) []string` | One explicit domain name, one package owner, no options or flags, and simpler call sites | A future caller could accidentally change ordering or provider spelling semantics | Replace the weak single-provider method test with an exact-order table test covering default, OpenStack, and vSphere branches | **Implement now** |
| Canonical config paths versus manual secrets paths | `internal/config/persistence`; manual construction in `internal/secrets/multi_cluster.go`, `rotation.go`, and `manager.go` | Locate OpenCenter configuration/state | Canonical resolution honors configured, environment, XDG/platform, and legacy behavior; secrets code currently does not | Existing configuration path owner, after policy characterization | A narrow existing resolver or explicitly named path function | Could remove real inconsistency, but only after migration behavior is defined | Medium-high: existing installations may depend on current paths | Characterize explicit directory, environment override, platform default, and legacy fallback | **Defer** |
| State-directory resolution | `internal/config` and `internal/config/persistence` | Resolve state directory | Same name but different starting points and fallback contracts; the persistence helper was audited as dead | None | None | A shared function would erase meaningful policy differences | High: silent path changes | Retain existing path tests; document the live contract | **Leave duplicated** |
| SOPS age-key file loading | `internal/sops/manager.go` and `internal/sops/git.go` | Expand a key path and load an age key | One preserves file whitespace while the other trims it; empty-path handling also needs an explicit policy | Potentially `internal/sops`, only after policy is decided | A SOPS-specific loader with an explicit content contract | Could centralize security-sensitive parsing only if behavior is intentionally standardized | Medium: whitespace changes are observable and security-sensitive; `GitIntegrator` support status is unresolved | Add empty path, home expansion, newline, and whitespace characterization | **Defer** |
| Atomic file replacement | `internal/util/files`, `internal/util/fs`, `internal/gitops/atomic.go`, OpenStack sync, and `cmd/secrets_sops_helpers.go` | Avoid partial file replacement | Retry, permissions, temporary naming, durability, staged transactions, rollback, errors, and external-process behavior differ | No shared location now; possibly a future narrow `internal/atomicfile` | One replacement primitive only after its contract is specified | No current abstraction is both smaller and behavior-preserving | High: durability, permission, cleanup, and rollback regressions | Characterize mode preservation, sync behavior, overwrite, cleanup, platforms, and failure rollback | **Leave duplicated** |
| Retry classification | `internal/util/errors/error_handler.go` and `internal/config/v2/errors.go` | Decide whether an operation/error is retryable | General string-pattern classification versus timeout and `Temporary()` semantics | None | None; a future design would use typed retry reasons | Related syntax does not establish the same operational contract | Medium: retries can repeat destructive work | Preserve existing error tests; require operation-level safety tests before any redesign | **Leave duplicated** |
| Reflection traversal | `internal/config/flags/integration.go` and `internal/cluster/configure_service.go` | Walk configuration fields | CLI precedence/security semantics versus cluster mutation/validation semantics | None | None | A generic traversal would hide the behavior each domain owns | High: reflective and validation behavior is difficult to trace | Existing flags and cluster tests are insufficient to prove one shared contract | **Leave duplicated** |
| Path parsing wrappers | `internal/config`, `internal/config/persistence`, and `internal/core/paths` | Parse and expose configuration paths across layers | Low-level parsing, persistence policy, and compatibility-facing APIs have distinct owners | None pending a deprecation plan | None | Local wrappers preserve dependency direction and compatibility intent | High: flattening can break compatibility or invert dependencies | Retain path/config coverage; add deprecation tests only with a future migration | **Leave duplicated** |
| Command-level SOPS workflows | `cmd/secrets_sops_helpers.go`, `internal/sops`, and `internal/secrets` | Discover and transform SOPS-managed files | Command code also owns prompts, console rendering, argument policy, replacement, and exit behavior | Potential future narrow service in `internal/sops` or `internal/secrets` | Explicit workflow methods with Cobra kept in `cmd` | Some domain behavior may be reusable, but extraction is not a small deduplication | Medium-high: output and filesystem behavior are externally observable | Characterize discovery, encryption/decryption, status, replacement, prompts, and output first | **Defer** |
| Drift scheduling/orchestration | `cmd/cluster_drift.go` and `internal/operations` | Schedule and run drift operations | Detection, persistence, scheduling, Cobra binding, and presentation responsibilities are interleaved | Potential future `internal/operations/drift` service | Explicit orchestration service with rendering retained in `cmd` | The boundary may be valuable, but it is architectural extraction rather than a narrow duplicate | Medium: scheduling and persistence behavior | Add command/service characterization around lifecycle and persisted state | **Defer** |

## 4. Implemented deduplications

### SOPS overlay encryption file selection

- **Before locations:** `DefaultSOPSManager.getFilesToEncrypt` in `internal/sops/manager.go:438-464` and `GitIntegrator.getFilesToEncrypt` in `internal/sops/git.go:546-572`. The manager had one caller; the Git integrator had two.
- **New shared location:** `internal/sops/overlay_files.go`.
- **New function/type names:** private function `overlayFilesToEncrypt(cfg *v2.Config) []string`; no new type was needed.
- **Why the semantic intent is the same:** both methods returned the same ordered default paths and the same provider-specific additions. Both supplied files to the SOPS encryption workflow, and their overlay-path parameters were unused. The implementations were byte-identical apart from receiver type.
- **Why this location owns the concept:** selection is SOPS package policy shared by two SOPS implementations. Keeping it private avoids a new package, dependency, or public contract.
- **Tests added or updated:** added `TestOverlayFilesToEncrypt`, a table-driven exact-slice test for a provider with no additions, OpenStack, and vSphere. Removed the superseded `TestGitIntegrator_getFilesToEncrypt`, which only checked unordered membership for OpenStack and was tied to the deleted receiver method.
- **Validation results:** the test was first observed failing to compile because `overlayFilesToEncrypt` did not exist. After the minimal implementation, the focused test and the complete `internal/sops` package passed. The full suite retained only its documented baseline documentation-policy failure.

## 5. Candidates intentionally left duplicated

The following were deliberately not unified:

- Configuration and secrets paths: deferred until path precedence and legacy installation behavior are characterized.
- State-directory resolvers: left separate because their contracts differ despite similar names.
- Age-key loaders: deferred because trimming versus preserving whitespace is observable, security-sensitive behavior.
- Atomic writers: left separate because they provide different durability, permission, transaction, rollback, and process-execution guarantees.
- Retry classifiers: left separate because they answer different retry questions and retries may repeat destructive operations.
- Reflection traversals: left separate because CLI precedence and cluster mutation are different domains.
- Path wrappers: retained as intentional layering and compatibility boundaries.
- Command-level SOPS workflows: deferred because a safe change requires broader CLI/filesystem characterization and service extraction.
- Drift orchestration: deferred because the proposed change is architectural extraction, not a small proven deduplication.

No candidate was merged solely for line-count reduction. In each deferred or retained case, the available shared API would either be vague, conceal behavior, require policy decisions, or broaden the phase beyond the smallest safe patch.

## 6. New or changed tests

- Added `internal/sops/overlay_files_test.go` with exact ordered expectations for all existing selection branches.
- Replaced the older OpenStack-only `GitIntegrator` membership test. The new test directly covers the shared policy and detects ordering changes, missing paths, additional paths, and provider-branch regressions.
- Recorded the test-first failure: the focused test initially failed with `undefined: overlayFilesToEncrypt`.
- After implementation, `go test ./internal/sops -run '^TestOverlayFilesToEncrypt$' -count=1` passed.
- The complete `internal/sops` package test suite passed after call-site migration.

## 7. Test and lint results

| Check | Result after deduplication |
|---|---|
| `gofmt` on changed Go files | Passed; `gofmt -d` produced no output. |
| Focused `TestOverlayFilesToEncrypt` | Passed. |
| `go test ./internal/sops` | Passed. |
| `go test ./...` | Failed only at the same baseline `cmd/TestDocsDoNotUseRemovedGACommands` documentation-policy test. All other packages passed, including `internal/sops`. |
| `go vet ./...` | Passed. |
| `staticcheck ./...` | Not run because the standalone executable is unavailable. Staticcheck-backed golangci findings remain included in the lint baseline below. |
| `golangci-lint run` | Reported the same 102 baseline findings: 50 `errcheck`, 2 `ineffassign`, and 50 `staticcheck`. No finding names a changed `internal/sops` file. |
| `go mod tidy -diff` | Passed with no output; `go.mod` and `go.sum` are unchanged. |
| `git diff --check` | Passed. |

The known full-suite and lint failures are pre-existing and unchanged in category, count, and cause. No new test, vet, lint, formatting, or module-tidiness regression was introduced by this patch.

## 8. Behavior compatibility notes

- The exact returned path strings and their order are unchanged.
- Provider matching remains exact and case-sensitive. In particular, the existing `"vsphere"` branch was preserved rather than normalized or broadened.
- The default behavior for providers without an explicit branch remains the same two standard files.
- Callers still join returned relative paths with their existing overlay/repository paths. Removing the unused overlay-path argument therefore changes no path computation.
- Existing logging, error propagation, context usage, encryption behavior, Git staging behavior, and file existence checks were not modified.
- The new function and files are private to `internal/sops`; no public API or package dependency changed.

## 9. Risks and follow-up work

- Configuration code may canonicalize provider names differently from the historical exact `"vsphere"` branch. This phase deliberately preserves current behavior. Any normalization change needs separate compatibility evidence.
- The full suite remains red because the required audit document contains historical wording rejected by a command documentation-policy test. Resolve the policy/report conflict separately; it is not a SOPS regression.
- The 102-finding golangci-lint baseline remains unchanged and should be handled as dedicated lint work rather than mixed into semantic refactoring.
- The secrets path, age-key whitespace, command-level SOPS, and drift candidates need characterization and explicit policy decisions before reconsideration.
- If future provider branches are added, update `TestOverlayFilesToEncrypt` with exact ordered expectations so both manager and Git workflows remain aligned.

## 10. Handoff instructions for Prompt 4

1. Start from the current post-Prompt-3 worktree and read `docs/refactor-audit-report.md`, `docs/dead-code-removal-report.md`, and this report in full.
2. Preserve `overlayFilesToEncrypt` as the single private owner of SOPS overlay file selection. Do not restore receiver-specific copies or change provider/path semantics without explicit compatibility tests.
3. Treat every deferred candidate in section 3 as unresolved, not pre-approved. In particular, path precedence, age-key whitespace, CLI output, atomic durability, retries, and reflection contracts require characterization before change.
4. Preserve the existing full-test baseline distinction: the documentation-policy failure predates this phase, vet passes, standalone staticcheck is unavailable, and golangci-lint has 102 existing findings.
5. Keep Prompt 4 changes separate from the untracked required reports and the scoped Prompt 3 files. Inspect `git status` before editing and do not discard user-owned work.
6. Run focused tests for every touched package, then `gofmt`, `go test ./...`, `go vet ./...`, available lint checks, `go mod tidy -diff`, and `git diff --check`. Report pre-existing failures separately from regressions.
