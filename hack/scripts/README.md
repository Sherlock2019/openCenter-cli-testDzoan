# hack/scripts

Maintenance utilities for the openCenter CLI repository. Reusable
across the openCenter ecosystem.

## Available scripts

| Script | Purpose |
|---|---|
| `audit_doc_frontmatter.py` | Verify every `docs/**/*.md` page has the required Diátaxis frontmatter (`id`, `title`, `sidebar_label`, `description`, `doc_type`, `audience`, `tags`) and that `doc_type` is one of `tutorial`, `how-to`, `reference`, `explanation`. Exits non-zero on failure for CI use. |
| `add_purpose_line.py` | Insert the steering-rule-mandated `**Purpose:**` line on pages that are missing one. Skips auto-generated Cobra reference pages. Idempotent. |
| `openstack-reset.sh` | Tear down OpenStack project resources between cluster lifecycle tests. Not documentation-related. |
| `tf2yaml.py` | Lightweight Terraform-to-YAML converter for `locals` and `module` blocks. Used by the GitOps templating helpers. |

## Usage

Run from the repository root:

```bash
# Recommended order for a doc refresh:
python3 hack/scripts/add_purpose_line.py
python3 hack/scripts/audit_doc_frontmatter.py
```

The doc-related scripts assume Markdown sources under
`docs/<category>/<name>.md`.

## When to use these

- Pre-merge validation that every page has Diátaxis frontmatter and a
  Purpose line (Documentation steering rule §2).
- Adding fresh pages: copy frontmatter from a sibling page and run
  `audit_doc_frontmatter.py` before commit.
