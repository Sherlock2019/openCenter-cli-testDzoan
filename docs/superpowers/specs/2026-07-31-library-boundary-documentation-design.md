# Library Boundary Documentation Design

Date: 2026-07-31

## Context

This phase follows the repository audit, dead-code removal, and semantic deduplication work. The three reports show that the repository already has specific domain and infrastructure packages under `internal/`, and that the remaining repeated-looking code has meaningful differences in compatibility, ordering, durability, or command behavior.

The goal is therefore to improve package comprehension without forcing code into a new abstraction.

## Extraction decision

No runtime package extraction is approved for this phase.

The strongest candidates were evaluated against the required extraction criteria:

| Candidate | Why it looks reusable | Why it is not extracted now |
| --- | --- | --- |
| SOPS command workflow | Several commands coordinate key loading, file selection, encryption, and output | Prompting, output, path policy, key handling, and file replacement semantics are still coupled to individual commands. Characterization tests are needed before an API can be both small and behavior-preserving. |
| Drift orchestration | Commands and operations code coordinate detection, persistence, and reporting | Goroutine lifecycle, callback sequencing, stored report shape, and provider-specific behavior are not yet captured as a stable integration contract. |
| Atomic file writing | Multiple areas write or replace files | Mode preservation, durability, rollback, and error handling differ by caller. A shared API would either erase those differences or expose too many policy switches. |
| Retry classification | Several integrations classify transient failures | The classifications are integration-specific and do not yet form a stable lower-level policy. |

Each candidate currently fails at least one of the small-API, minimal-dependency, independently testable, or call-site-readability criteria. Keeping the code in its current packages is safer than introducing a bridge package or a policy-heavy abstraction.

## Approved changes

The implementation is documentation-only:

1. Rewrite `internal/di` package documentation to identify the typed `App` graph as canonical and the reflection container as a compatibility boundary.
2. Add package documentation for `internal/config/v2`, including its processing stages and ownership rules.
3. Add package documentation for `internal/sops`, including command/package separation and order-sensitive overlay behavior.
4. Update `cmd` package documentation to describe command orchestration, application wiring, and where domain logic belongs.
5. Add repository-level architecture and LLM navigation documents.
6. Add a library extraction report that records the no-extraction result and validation evidence.
7. Repair one stale audit-report reference that is rejected by the repository's documentation drift test.

No exported identifier, package import, runtime control flow, file format, command behavior, or generated artifact changes.

## Dependency direction

The intended dependency direction is:

```text
entrypoints -> command wiring -> internal domain/infrastructure packages
                               -> typed application graph
internal domain packages -> lower-level focused packages and external clients
```

Lower-level packages must not import command packages. The DI package may construct lower-level components but must not own domain behavior. Command code owns user interaction and orchestration, while reusable behavior belongs in a specifically named `internal` package only after its contract is stable.

## Validation

The documentation changes will be checked with:

- formatting for changed Go files;
- targeted tests for command documentation drift and the documented packages;
- the complete Go test suite;
- Go vet;
- standalone static analysis when installed;
- the repository lint configuration when available;
- a module tidy diff check; and
- a whitespace/error diff check.

Pre-existing lint findings will be reported separately from regressions introduced by this phase.

## Deferred work

Future extraction work should begin with characterization tests, not a package move. SOPS orchestration needs tests around prompts, output, key sources, replacement behavior, and file order. Drift orchestration needs lifecycle, callback, persistence, and provider-boundary tests. Atomic writes and retry classification should remain local until callers share the same policy rather than merely similar control flow.
