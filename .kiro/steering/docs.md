# Documentation Standards

## Table of Contents Requirement

**Rule:** All markdown files exceeding 100 lines MUST include a table of contents.

### Placement
- Position TOC immediately after the main title and any introductory paragraph
- Use `## Table of Contents` as the heading
- Leave one blank line before and after the TOC

### Format
```markdown
## Table of Contents

- [Section Name](#section-name)
  - [Subsection Name](#subsection-name)
- [Another Section](#another-section)
```

### Link Generation
- Convert headings to lowercase
- Replace spaces with hyphens
- Remove special characters except hyphens
- Example: `## My Section Name!` → `#my-section-name`

### Automation
When creating or updating large markdown files:
1. Count lines in the file
2. If > 100 lines, generate TOC from all `##` and `###` headings
3. Insert after main title
4. Update TOC when headings change

### Exceptions
- README.md files under 100 lines
- Auto-generated documentation (schemas, API references)
- Files explicitly marked with `<!-- no-toc -->` comment

## Documentation Structure

Follow Diátaxis framework:
- **Tutorials**: Learning-oriented, step-by-step guides
- **How-To Guides**: Task-oriented, problem-solving recipes
- **Reference**: Information-oriented, technical descriptions
- **Explanation**: Understanding-oriented, conceptual discussions

## Writing Style

- Use present tense
- Active voice preferred
- Code examples must be runnable
- Include expected output for commands
- Link to related documentation
- Keep paragraphs short (3-5 sentences)

## Code Examples

- Always use syntax highlighting with language identifier
- Include context (what the code does)
- Show both command and output when relevant
- Use realistic but sanitized data (no real credentials)

Example:
```bash
# Initialize a new cluster
mise run build
./bin/opencenter cluster init my-cluster

# Expected output:
# ✓ Created cluster configuration at ~/.config/opencenter/clusters/opencenter/.my-cluster-config.yaml
```

## File Naming

- Use kebab-case: `my-document.md`
- Be descriptive: `cluster-bootstrap-workflow.md` not `workflow.md`
- Group related docs in subdirectories
- Use README.md for directory overviews
