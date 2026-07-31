# Library Boundary Documentation Implementation Plan

Date: 2026-07-31

## Objective

Document the repository's existing stable package boundaries and record the evidence-based decision not to extract a new runtime package in this phase. Preserve all behavior and exported APIs.

## Task 1: Record the extraction decision

Files:

- Create `docs/superpowers/specs/2026-07-31-library-boundary-documentation-design.md`.
- Create `docs/superpowers/plans/2026-07-31-library-boundary-documentation.md`.

Work:

1. Record the evaluated candidates and the criteria they do not yet meet.
2. State the approved documentation-only scope.
3. Define dependency direction and validation expectations.

Verification:

- Review the documents against all eight extraction criteria.
- Confirm that no speculative package is proposed.

## Task 2: Clarify important Go package boundaries

Files:

- Modify `cmd/doc.go`.
- Modify `internal/di/doc.go`.
- Create `internal/config/v2/doc.go`.
- Create `internal/sops/doc.go`.

Work:

1. Describe `cmd` as the Cobra command, user interaction, and orchestration layer.
2. Describe typed application wiring as the canonical DI path and the reflection container as compatibility-only.
3. Describe the configuration pipeline and the ownership of typed schema behavior.
4. Describe the SOPS integration boundary and its order-sensitive file selection.
5. State where new behavior should and should not be added.

Verification:

- Run `gofmt` on the four files.
- Run targeted tests for `cmd`, `internal/di`, `internal/config/v2`, and `internal/sops`.

## Task 3: Add repository navigation documentation

Files:

- Create `docs/architecture.md`.
- Create `docs/llm-code-map.md`.

Work:

1. Document entrypoints, commands, package responsibilities, dependency direction, runtime and configuration flow, data flow, external integrations, embedded or generated boundaries, testing, and operations.
2. Document the application wiring location, key types and interfaces, common workflows, safe extension points, unsafe dumping grounds, dynamic references, and practical test commands.
3. Keep descriptions tied to current source paths and avoid describing removed command syntax.

Verification:

- Cross-check every named entrypoint and package against the source tree.
- Run the command documentation drift test.

## Task 4: Repair the stale audit reference

File:

- Modify `docs/refactor-audit-report.md`.

Work:

1. Replace the stale literal name of a removed schema-related subcommand with a behavior-level description.
2. Preserve the audit finding that the task no longer maps to a current command.

Verification:

- Run the command documentation drift test and confirm the baseline failure is resolved.

## Task 5: Write the extraction report

File:

- Create `docs/library-extraction-report.md`.

Work:

1. Include all eleven required report sections.
2. State that no package was extracted and explain the evidence.
3. Record package documentation changes, API compatibility, tests, validation results, risks, deferred candidates, and next steps.
4. Update validation results after executing the full checks.

Verification:

- Check the report headings and required content.
- Confirm it distinguishes baseline failures from final results.

## Task 6: Run final validation

Work:

1. Format changed Go files.
2. Run targeted package tests.
3. Run `go test ./...`.
4. Run `go vet ./...`.
5. Run standalone `staticcheck ./...` if installed.
6. Run `golangci-lint run` using the repository configuration when available.
7. Run `go mod tidy -diff`.
8. Run `git diff --check` and inspect the final diff and status.

Expected result:

- Runtime behavior and public APIs remain unchanged.
- All Go tests and vet pass after the stale documentation reference is repaired.
- Any remaining linter failures are identified as pre-existing debt rather than silently presented as success.
