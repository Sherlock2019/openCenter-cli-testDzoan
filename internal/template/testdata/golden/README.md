# Golden File Tests

This directory contains golden files for template rendering tests. Golden files are expected outputs that are compared against actual template rendering results to ensure consistency and correctness.

## What are Golden Files?

Golden files are reference outputs that represent the expected result of template rendering. They serve as a baseline for comparison in tests, ensuring that template changes don't inadvertently alter output in unexpected ways.

## Running Golden File Tests

### Normal Test Run

To run the golden file tests and compare against existing golden files:

```bash
go test -v ./internal/template -run TestTemplateRenderingGolden
```

### Updating Golden Files

When you intentionally change template behavior or add new test cases, you need to update the golden files:

```bash
go test -v ./internal/template -run TestTemplateRenderingGolden -update-golden
```

**Warning:** Only use `-update-golden` when you've verified that the new output is correct. This flag will overwrite existing golden files.

## Golden Files in This Directory

- `simple.golden` - Simple template with basic variable substitution
- `cluster-config.golden` - Kubernetes cluster configuration template
- `inventory.golden` - Ansible inventory template with loops
- `service-manifest.golden` - Kubernetes service manifest with complex data structures
- `custom-functions.golden` - Template using custom registered functions
- `empty-data.golden` - Template with empty data input
- `nil-values.golden` - Template handling nil values

## Adding New Golden File Tests

To add a new golden file test:

1. Add a new test case to `engine_golden_test.go` in the `TestTemplateRenderingGolden` function
2. Run the test with `-update-golden` to generate the golden file
3. Verify the generated golden file contains the expected output
4. Commit both the test code and the golden file

## Best Practices

1. **Keep golden files small and focused** - Each golden file should test a specific template feature
2. **Review golden file changes carefully** - When updating golden files, always review the diff to ensure changes are intentional
3. **Use descriptive names** - Golden file names should clearly indicate what they're testing
4. **Version control** - Always commit golden files to version control
5. **Document expected behavior** - Add comments in test code explaining what each golden file validates

## Troubleshooting

### Test fails with "does not match golden file"

This means the template output has changed. Either:
- The change is intentional: Run with `-update-golden` to update the golden file
- The change is a bug: Fix the template or code causing the unexpected output

### Golden file not found

Run the test with `-update-golden` to create the missing golden file.

### Differences in line endings

Ensure consistent line endings across platforms. The test framework handles this automatically, but if you manually edit golden files, use LF (Unix-style) line endings.
