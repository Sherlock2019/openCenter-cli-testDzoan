# IDE Integration


## Table of Contents

- [Who This Is For](#who-this-is-for)
- [What You Get](#what-you-get)
- [Quick Setup](#quick-setup)
- [Visual Studio Code](#visual-studio-code)
- [JetBrains IDEs](#jetbrains-ides)
- [Vim/Neovim](#vimneovim)
- [Emacs](#emacs)
- [Schema Management](#schema-management)
- [Using the opencenter Config IDE Command](#using-the-opencenter-config-ide-command)
- [YAML Linting](#yaml-linting)
- [Troubleshooting](#troubleshooting)
- [Best Practices](#best-practices)
- [Related Documentation](#related-documentation)
- [External Resources](#external-resources)
**doc_type: how-to**

This guide shows you how to set up IDE integration for opencenter cluster configuration files with autocomplete, validation, and inline documentation.

## Who This Is For

Developers and operators who edit opencenter cluster configuration files and want IDE support for faster, error-free configuration.

## What You Get

- **Autocomplete**: Intelligent suggestions for configuration keys and values
- **Validation**: Real-time error detection against the JSON schema
- **Documentation**: Hover tooltips with field descriptions and constraints
- **Error Detection**: Immediate feedback on typos and invalid values

## Quick Setup

Run the IDE integration command:

```bash
opencenter config ide
```

This generates the JSON schema and configures VS Code automatically. For other IDEs, see the sections below.

## Visual Studio Code

VS Code has the best support through the YAML extension.

### Setup Steps

1. Install the YAML extension:
   ```bash
   code --install-extension redhat.vscode-yaml
   ```

2. Run the setup command:
   ```bash
   opencenter config ide --ide=vscode
   ```

3. Restart VS Code

4. Open any `*-config.yaml` file and start typing

### Features

- **Autocomplete**: Press `Ctrl+Space` (or `Cmd+Space` on macOS)
- **Validation**: Errors appear as you type
- **Hover Docs**: Hover over any key for descriptions
- **Format**: Press `Shift+Alt+F` to format YAML

### Manual Configuration

If the automatic setup doesn't work, add this to `.vscode/settings.json`:

```json
{
  "yaml.schemas": {
    "./schema/cluster.schema.json": [
      "**/clusters/**/*.yaml",
      "**/clusters/**/*-config.yaml",
      "**/.opencenter.yaml"
    ]
  },
  "yaml.validate": true,
  "yaml.completion": true,
  "yaml.format.enable": true
}
```

## JetBrains IDEs

IntelliJ IDEA, PyCharm, WebStorm, and other JetBrains IDEs have built-in JSON Schema support.

### Setup Steps

1. Generate the schema:
   ```bash
   opencenter cluster schema --out schema/cluster.schema.json
   ```

2. Open **Settings/Preferences** → **Languages & Frameworks** → **Schemas and DTDs** → **JSON Schema Mappings**

3. Click **+** to add a new mapping:
   - **Name**: opencenter Cluster Configuration
   - **Schema file or URL**: `schema/cluster.schema.json`
   - **Schema version**: JSON Schema version 7

4. Add file patterns (click **+** in the bottom section):
   - `**/clusters/**/*.yaml`
   - `**/clusters/**/*-config.yaml`
   - `**/.opencenter.yaml`

5. Click **Apply** and restart your IDE

### Features

- **Autocomplete**: Press `Ctrl+Space`
- **Quick Documentation**: Press `Ctrl+Q` (Windows/Linux) or `F1` (macOS)
- **Reformat**: Press `Ctrl+Alt+L` (Windows/Linux) or `Cmd+Option+L` (macOS)

## Vim/Neovim

Use the YAML language server with coc.nvim or nvim-lspconfig.

### Option 1: coc.nvim

1. Install [coc.nvim](https://github.com/neoclide/coc.nvim)

2. Install the YAML language server:
   ```vim
   :CocInstall coc-yaml
   ```

3. Add to `coc-settings.json` (`:CocConfig`):
   ```json
   {
     "yaml.schemas": {
       "./schema/cluster.schema.json": [
         "**/clusters/**/*.yaml",
         "**/clusters/**/*-config.yaml",
         "**/.opencenter.yaml"
       ]
     },
     "yaml.validate": true,
     "yaml.completion": true
   }
   ```

4. Generate the schema:
   ```bash
   opencenter cluster schema --out schema/cluster.schema.json
   ```

### Option 2: nvim-lspconfig

1. Install [nvim-lspconfig](https://github.com/neovim/nvim-lspconfig)

2. Install yaml-language-server:
   ```bash
   npm install -g yaml-language-server
   ```

3. Add to `init.lua`:
   ```lua
   require'lspconfig'.yamlls.setup{
     settings = {
       yaml = {
         schemas = {
           ["./schema/cluster.schema.json"] = {
             "**/clusters/**/*.yaml",
             "**/clusters/**/*-config.yaml",
             "**/.opencenter.yaml"
           }
         },
         validate = true,
         completion = true
       }
     }
   }
   ```

4. Generate the schema:
   ```bash
   opencenter cluster schema --out schema/cluster.schema.json
   ```

## Emacs

Use lsp-mode with yaml-language-server.

### Setup Steps

1. Install [lsp-mode](https://github.com/emacs-lsp/lsp-mode)

2. Install yaml-language-server:
   ```bash
   npm install -g yaml-language-server
   ```

3. Add to `init.el`:
   ```elisp
   (use-package lsp-mode
     :hook (yaml-mode . lsp)
     :config
     (setq lsp-yaml-schemas
           '(:cluster "./schema/cluster.schema.json")))
   
   (add-to-list 'auto-mode-alist '("\\*-config\\.yaml\\'" . yaml-mode))
   (add-to-list 'auto-mode-alist '("\\.opencenter\\.yaml\\'" . yaml-mode))
   ```

4. Generate the schema:
   ```bash
   opencenter cluster schema --out schema/cluster.schema.json
   ```

5. Restart Emacs

## Schema Management

The JSON schema is generated from Go struct definitions and includes type validation, pattern matching, and inline documentation.

### Schema Versions

opencenter supports multiple schema versions:

- **v1.0** (current): Production schema at `schema/cluster.schema.json`
- **v2.0** (development): Next-generation schema at `schema/cluster-v2.schema.json`

For complete schema documentation, see [JSON Schema Reference](../reference/json-schema.md).

### Generating the Schema

Generate or update the schema:

```bash
# Generate v1.0 schema (current)
opencenter cluster schema --out schema/cluster.schema.json

# Generate v2.0 schema (development)
opencenter cluster schema --version 2.0 --out schema/cluster-v2.schema.json

# Pretty-print (default)
opencenter cluster schema --out schema/cluster.schema.json --pretty

# Check schema version
opencenter cluster schema --version
```

The schema includes:
- Type validation for all fields
- Pattern validation for CIDR blocks, UUIDs, hostnames
- Range validation for numeric fields
- Enum validation for predefined options
- Required field markers
- Field descriptions and constraints

### When to Regenerate

Regenerate the schema after:
- Updating opencenter to a new version
- Adding custom service types
- Modifying configuration structs
- Changing validation rules

Commit the updated schema to version control so your team uses the same validation rules.

### Schema Versioning

Configurations can specify their schema version:

```yaml
# v1.0 (implicit, default)
opencenter:
  meta:
    name: my-cluster

# v2.0 (explicit)
schema_version: "2.0"
opencenter:
  meta:
    name: my-cluster
```

For migration between versions, see [Migration Guide](../cluster-config/migration-guide.md).

## Using the opencenter Config IDE Command

The `opencenter config ide` command automates IDE setup.

### Basic Usage

```bash
# Auto-detect IDE and configure
opencenter config ide

# Target specific IDE
opencenter config ide --ide=vscode
opencenter config ide --ide=jetbrains
opencenter config ide --ide=vim
opencenter config ide --ide=emacs

# Generate schema only
opencenter config ide --schema-only

# Show setup instructions
opencenter config ide --show-instructions
opencenter config ide --show-instructions --ide=vim
```

### What It Does

1. Generates the latest JSON schema
2. Creates IDE-specific configuration files (VS Code only)
3. Sets up schema associations
4. Configures YAML validation and formatting

For VS Code, it creates or updates `.vscode/settings.json` with:
- Schema associations for cluster config files
- YAML validation and completion settings
- SOPS custom tags support
- Format-on-save configuration

## YAML Linting

The project includes `.yamllint` for consistent formatting.

### Install yamllint

```bash
# macOS
brew install yamllint

# Linux
pip install yamllint

# Verify
yamllint --version
```

### Lint Your Configs

```bash
# Lint a specific file
yamllint ~/.config/opencenter/clusters/myorg/my-cluster/.my-cluster-config.yaml

# Lint all cluster configs
yamllint ~/.config/opencenter/clusters/

# Check from project root
yamllint testdata/
```

The `.yamllint` configuration enforces:
- 2-space indentation
- Line length limits
- Consistent key ordering
- Proper quoting rules

## Troubleshooting

### Schema Not Loading

**Symptom**: No autocomplete or validation in your IDE

**Solutions**:
1. Check the schema file exists:
   ```bash
   ls -la schema/cluster.schema.json
   ```

2. Regenerate the schema:
   ```bash
   opencenter cluster schema --out schema/cluster.schema.json
   ```

3. Restart your IDE or reload the window

4. Check IDE logs for schema loading errors:
   - VS Code: View → Output → YAML Support
   - JetBrains: Help → Show Log in Finder/Explorer

### Validation Errors on Valid Config

**Symptom**: IDE shows errors for configuration that passes `opencenter cluster validate`

**Solutions**:
1. Ensure schema version matches opencenter version:
   ```bash
   opencenter cluster schema --version
   opencenter version
   ```

2. Regenerate schema after updating opencenter:
   ```bash
   opencenter cluster schema --out schema/cluster.schema.json
   ```

3. Check file path matches schema patterns (must contain `clusters/` or end with `-config.yaml`)

### Autocomplete Not Working

**Symptom**: No suggestions when typing

**Solutions**:
1. Verify YAML language server is running:
   - VS Code: Check Output → YAML Support
   - Vim: Run `:CocInfo` or check LSP status

2. Confirm file extension is `.yaml` or `.yml`

3. Trigger autocomplete manually:
   - VS Code: `Ctrl+Space` (Windows/Linux) or `Cmd+Space` (macOS)
   - JetBrains: `Ctrl+Space`
   - Vim: `Ctrl+X Ctrl+O` or let coc.nvim auto-trigger

4. Check that the file path matches schema patterns

### Performance Issues

**Symptom**: IDE becomes slow when editing large configs

**Solutions**:
1. Disable validation temporarily for large files
2. Split large configurations into multiple files
3. Increase IDE memory:
   - VS Code: Add `"files.maxMemoryForLargeFilesMB": 4096` to settings
   - JetBrains: Help → Edit Custom VM Options → increase `-Xmx`

### SOPS Encrypted Values

**Symptom**: Validation errors on encrypted values

**Solution**: The schema supports SOPS encrypted values. Ensure your IDE recognizes custom YAML tags:

VS Code (`.vscode/settings.json`):
```json
{
  "yaml.customTags": [
    "!vault",
    "!encrypted/pkcs1-oaep"
  ]
}
```

## Best Practices

### Configuration Organization

1. Store configs in `~/.config/opencenter/clusters/<org>/<cluster>/`
2. Use version control for all configuration files
3. Encrypt secrets with SOPS before committing
4. Run `opencenter cluster validate` before committing changes

### Schema Maintenance

1. Regenerate schema after updating opencenter
2. Commit schema changes with configuration changes
3. Document breaking schema changes in commit messages
4. Test existing configs after schema updates

### IDE Configuration

1. Enable format-on-save for YAML files
2. Use 2-space indentation (matches opencenter defaults)
3. Keep real-time validation enabled
4. Create snippets for common configuration patterns

### Example VS Code Snippet

Add to `.vscode/opencenter.code-snippets`:

```json
{
  "OpenCenter Service": {
    "prefix": "oc-service",
    "body": [
      "${1:service-name}:",
      "  enabled: ${2:true}",
      "  namespace: ${3:$1}",
      "  $0"
    ],
    "description": "Add an opencenter service"
  }
}
```

## Related Documentation

- [JSON Schema Reference](../reference/json-schema.md) - Complete JSON schema documentation
- [Configuration Reference](../reference/configuration.md) - Complete configuration field reference
- [CLI Commands](../reference/cli-commands.md) - All opencenter commands
- [Adding Services](adding-services.md) - How to enable and configure services
- [Migration Guide](../cluster-config/migration-guide.md) - v1 to v2 schema migration

## External Resources

- [JSON Schema Documentation](https://json-schema.org/)
- [YAML Language Server](https://github.com/redhat-developer/yaml-language-server)
- [VS Code YAML Extension](https://marketplace.visualstudio.com/items?itemName=redhat.vscode-yaml)
- [opencenter GitHub](https://github.com/rackerlabs/opencenter-cli)
