# Project Type Config

## Tool-specific overrides (Option A)

Use top-level defaults, with optional tool-specific overrides and additive lists.

Example:

```toml
name = "Example Project"
description = "Example project type"
skills = ["core-skill", "build-skill"]
commands = ["release-notes"]
mcps = ["demo"]
snippets = ["base_rules", "plan"]

[claude]
commands = ["release-notes"]
snippets = ["base_rules"]          # replaces defaults
snippets_append = ["extra"]        # appended after defaults (if no snippets override)
snippets_prepend = ["first"]       # prepended before defaults (if no snippets override)

[codex]
skills = ["codex-only-skill"]
commands = ["codex-release-notes"]
mcps = ["codex-only-mcp"]
snippets_append = ["codex_extra"]
```

Notes:
- Top-level `skills`, `commands`, `mcps`, and `snippets` define defaults for all tools.
- Tool-specific `skills`, `commands`, `mcps`, and `snippets` replace the defaults.
- Use `snippets_append` and `snippets_prepend` to add entries without replacing defaults.
