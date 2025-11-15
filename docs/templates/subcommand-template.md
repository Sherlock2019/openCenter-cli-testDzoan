# `command subcommand` - Brief description

## Synopsis
```
command subcommand [OPTIONS] [ARGUMENTS]
```

## Description

A detailed explanation of what this subcommand does, its purpose, and when to use it. Include any important context or background information that helps users understand its functionality.

## Arguments

### `<argument-name>`
- **Required/Optional**: Required
- **Description**: What this argument does
- **Example**: `example-value`

### `[optional-argument]`
- **Required/Optional**: Optional
- **Description**: What this optional argument does
- **Default**: `default-value` (if applicable)

## Options

### `-f, --flag`
- **Description**: What this flag does
- **Type**: Boolean
- **Default**: `false`

### `-o, --option <value>`
- **Description**: What this option does
- **Type**: string/number/etc.
- **Default**: `default-value`
- **Example**: `--option "example"`

### `-h, --help`
- **Description**: Display help information for this subcommand

## Examples

### Basic usage
```bash
command subcommand argument
```
Description of what this example does.

### With options
```bash
command subcommand --option value argument
```
Description of what this example does.

### Advanced usage
```bash
command subcommand -f --option value argument1 argument2
```
Description of what this more complex example accomplishes.

## Output

Description of what the command outputs, including:
- Success messages
- Output format (JSON, plain text, etc.)
- Exit codes

## Notes

- Any additional notes, warnings, or tips
- Common pitfalls or gotchas
- Related subcommands or workflows

## See Also

- `command other-subcommand` - Related subcommand
- `command --help` - Main command help