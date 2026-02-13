# mavu-llm-cli

llm manages LLM setup files for a project (skills + agent docs).

## Install

Build the binary first:

```bash
go build -o llm
```

Option A: user-local symlink

```bash
mkdir -p ~/.local/bin
ln -sf /Users/manfred/Documents/www/mavu-llm-cli/llm ~/.local/bin/llm
```

Option B: Homebrew bin (system-wide)

```bash
ln -sf /Users/manfred/Documents/www/mavu-llm-cli/llm /opt/homebrew/bin/llm
```

## Usage

```bash
llm types
llm init --type <project-type> [--path <dir>]
llm update [--path <dir>]
llm template-paths
llm opencode-sessions|os [--path <dir>] [--exclude-prefix <prefix>] [--storage-path <dir>] [filter]
```

When a project type defines `mcps`, llm writes `opencode.json`, `.mcp.json`, and
`.codex/config.toml` into the target directory using the MCP templates.

## Local Project Configuration

Projects can override global templates using the `.mavu/` directory:

### Local Skills

Create custom skills in `.mavu/skill_templates/<skill-name>/`:
- Local skills are auto-discovered and added to both `.claude/skills/` and `.codex/skills/`
- Local skills take precedence over global templates with the same name

### Local MCPs

Create project-specific MCP configurations in `.mavu/mcp.json`.

**Flat format** (recommended):
```json
{
  "my-local-mcp": {
    "type": "stdio",
    "command": "node",
    "args": ["./server.js"]
  },
  "tidewave": {
    "type": "http",
    "url": "http://localhost:9999/custom-path"
  }
}
```

**Wrapped format** (also supported):
```json
{
  "mcpServers": {
    "tidewave": {
      "type": "http",
      "url": "https://example.com/mcp",
      "headers": {
        "Authorization": "Bearer token"
      }
    }
  }
}
```

- Local MCPs take precedence over global MCP templates
- Supports environment variable expansion: `${VAR_NAME}`
- Merged with global MCPs defined in project config
- Written to `.mcp.json` (Claude Desktop), `opencode.json` (OpenCode), and `.codex/config.toml` (Codex)
- Note: Codex only loads `.codex/config.toml` for trusted projects (configured in `~/.codex/config.toml`).

## Templates

llm loads templates from disk (nothing is embedded). Set
`MAVU_LLM_TEMPLATES_DIR` to a directory that contains:

- `project_types/*.toml`
- `snippets/*.md`
- `skill_templates/<skill>/**`
- `command_templates/<command>.md`
- `mcp_templates/<mcp>.mcp.json`

If the env var is not set, llm searches the current working directory
and the directory containing the `llm` binary for `project_types/`.
