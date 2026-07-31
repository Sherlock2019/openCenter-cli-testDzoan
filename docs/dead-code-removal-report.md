# Dead-Code Removal Report

## 1. Summary

This phase removed one test-only private helper that the audit explicitly classified as **Safe to remove**:

- `NewLegacyConfigurationMerger` in `internal/config/flags/backward_compatibility_property_test.go`

The other safe-listed item was intentionally not removed because the expression and symbol context recorded in the audit do not match the current source. No production runtime code, exported API, generated code, build-tag-specific code, package structure, dependencies, or behavior was changed.

## 2. Baseline status before edits

The working tree contained one pre-existing untracked file:

- `docs/refactor-audit-report.md`

No other local changes were present before this phase.

The repository was validated with Go 1.26.5. The inherited environment initially pointed `GOROOT` at Go 1.26.4, so validation commands were rerun with `GOROOT` and `PATH` explicitly set to the installed Go 1.26.5 toolchain.

| Check | Baseline result |
|---|---|
| `go test ./internal/importer ./internal/config/flags` | Passed |
| `go test ./...` | Failed only in `cmd` at `TestDocsDoNotUseRemovedGACommands`; the required audit report contains prohibited deprecated-command terminology at line 288. All other package results passed. |
| `go vet ./...` | Passed |
| `staticcheck ./...` | Unavailable as a standalone executable; it was not installed |
| `golangci-lint run` | Failed with 102 pre-existing findings: 50 `errcheck`, 2 `ineffassign`, and 50 `staticcheck` findings |

The full test, vet, and lint commands were rerun outside the filesystem sandbox after initial attempts encountered Go build-cache permission errors.

## 3. Removed code

### `NewLegacyConfigurationMerger`

- **Symbol:** `NewLegacyConfigurationMerger`
- **Package path:** `github.com/opencenter-cloud/opencenter-cli/internal/config/flags`
- **Location before removal:** `internal/config/flags/backward_compatibility_property_test.go:315`
- **Audit evidence:** The dead-code candidate matrix explicitly classified the private, test-only helper as **Safe to remove**. Repository-wide exact-name inspection found its declaration but no callers, so it was unreachable from the active test suite.
- **Validation performed:** The affected package passed before and after removal; the exact symbol name is absent after removal; `gofmt` was run on the changed Go file, with unrelated pre-existing whitespace restored to keep the diff scoped to the audited removal; the full test suite retained only its documented baseline failure; `go vet ./...` passed; `golangci-lint run` retained the same 102 findings and category counts; and `go mod tidy -diff` produced no changes.

No tests were added or rewritten because the removed item was an unreachable test helper and the existing package tests continued to pass.

## 4. Candidates intentionally not removed

### Audited importer self-assignment

- **Audit entry:** `result.Confidence = result.Confidence`
- **Package path:** `github.com/opencenter-cloud/opencenter-cli/internal/importer`
- **Audit location:** `internal/importer/scanner.go:411`
- **Reason not removed:** The exact audited expression is absent from the current repository. The current source at that location contains a different self-assignment, `info[currentCluster] = info[currentCluster]`. Although lint identifies the current statement as a no-op, deleting it would require interpreting or correcting the audit rather than implementing the exact safe-listed candidate.
- **Recommended action:** Reconcile the audit entry with the current source and repeat dynamic-reference and behavior checks before authorizing removal.

Every item classified as **Probably safe but requires confirmation**, **Unsafe to remove**, or **Keep** in the audit report was left untouched.

## 5. Reverted removals, if any

None. No implemented removal caused a test, build, vet, or lint regression.

The importer candidate described above was never removed, so it is not a reverted change.

## 6. Test and lint results

| Check | Before edits | After edits |
|---|---|---|
| `go test ./internal/config/flags` | Passed as part of focused baseline | Passed |
| `go test ./...` | Failed only at the documented `cmd` documentation-policy test | Failed only at the same documented test |
| `go vet ./...` | Passed | Passed |
| `staticcheck ./...` | Standalone executable unavailable | Standalone executable unavailable |
| `golangci-lint run` | 102 findings: 50 `errcheck`, 2 `ineffassign`, 50 `staticcheck` | Same 102 findings and category counts |
| `gofmt` on changed Go files | Not applicable | Run; unrelated pre-existing whitespace was retained to avoid broad formatting |
| `go mod tidy -diff` | Not applicable | Passed with no output |
| `git diff --check` | Not applicable | Passed |

The pre-existing documentation-policy failure is caused by terminology in the required audit report, not by the Go code removal. The lint baseline also remains unchanged.

## 7. go.mod/go.sum changes, if any

There are no changes to `go.mod` or `go.sum`. `go mod tidy -diff` completed successfully without producing a diff.

## 8. Risk assessment

The implemented change is low risk:

- It removes one unexported helper from a `_test.go` file.
- Exact-name inspection found no callers.
- Focused package tests pass.
- Full-suite, vet, lint, and module-tidiness results did not regress from baseline.
- No public API, dynamic registration point, generated file, platform-specific file, configuration surface, migration, route, handler, or production dependency was changed.

Residual issues are explicitly contained: the mismatched importer candidate remains in place, the known documentation-policy test still fails on the audit report, and the existing lint backlog is unchanged.

## 9. Handoff instructions for Prompt 3

Use the following instructions verbatim as the operational handoff:

> Use `docs/refactor-audit-report.md` and `dead-code-removal-report.md` as authoritative inputs. Preserve the Prompt 2 diff and do not repeat dead-code removal.
>
> Limit Prompt 3 to duplicate-logic candidates that the audit explicitly assigns to Prompt 3. Do not perform package extraction, unrelated architecture cleanup, public API changes, generated-code edits, broad renaming, or opportunistic formatting.
>
> For each candidate, confirm semantic intent, domain ownership, caller behavior, dynamic-reference risks, and existing test coverage before editing. Do not merge code solely because it is syntactically similar. Add or strengthen characterization tests before changing shared behavior where coverage is insufficient.
>
> Work one semantic group at a time. After each group, run `gofmt` on changed Go files, focused package tests, and broader validation when practical. Revert any refactor that changes behavior or causes a new test, build, vet, or lint failure.
>
> Do not remove the importer statement discussed in section 4 unless a corrected audit explicitly authorizes it.
>
> Record the working-tree state before edits and preserve the required audit and Prompt 2 report files. The known baseline is: `go test ./...` has one documentation-policy failure caused by line 288 of the audit report; `go vet ./...` passes; `golangci-lint run` reports 102 existing findings; and standalone `staticcheck` is unavailable.
>
> Finish with `go test ./...`, `go vet ./...`, `golangci-lint run`, and `go mod tidy -diff`, clearly separating pre-existing failures from regressions.
