# mavu-llm-cli

mavu-llm manages LLM setup files for a project (skills + agent docs).

## Install

Build the binary first:

```bash
go build -o mavu-llm
```

Option A: user-local symlink

```bash
mkdir -p ~/.local/bin
ln -sf /Users/manfred/Documents/www/mavu-llm-cli/mavu-llm ~/.local/bin/mavu-llm
```

Option B: Homebrew bin (system-wide)

```bash
ln -sf /Users/manfred/Documents/www/mavu-llm-cli/mavu-llm /opt/homebrew/bin/mavu-llm
```

## Usage

```bash
mavu-llm types
mavu-llm init --type <project-type> [--path <dir>]
mavu-llm update [--path <dir>]
mavu-llm template-paths
```

When a project type defines `mcps`, mavu-llm writes `opencode.json` and `.mcp.json`
into the target directory using the MCP templates.

## Templates

mavu-llm loads templates from disk (nothing is embedded). Set
`MAVU_LLM_TEMPLATES_DIR` to a directory that contains:

- `project_types/*.toml`
- `snippets/*.md`
- `skill_templates/<skill>/**`
- `command_templates/<command>.md`
- `mcp_templates/<mcp>.mcp.json`

If the env var is not set, mavu-llm searches the current working directory
and the directory containing the `mavu-llm` binary for `project_types/`.
