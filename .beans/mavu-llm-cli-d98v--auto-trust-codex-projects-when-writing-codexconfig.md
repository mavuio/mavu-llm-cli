---
# mavu-llm-cli-d98v
title: Auto-trust Codex projects when writing .codex/config.toml
status: completed
type: task
priority: normal
created_at: 2026-02-07T21:28:18Z
updated_at: 2026-02-07T21:32:03Z
---

After writing .codex/config.toml into a project, also upsert a trust entry in ~/.codex/config.toml so Codex picks up the project-scoped config.

- [x] Add ensureCodexProjectTrusted() function in main.go
- [x] Call it from runSetup() after writeCodexMcpConfig()
- [x] Verify go build succeeds
- [x] Verify go test passes

## Summary of Changes

Added `ensureCodexProjectTrusted()` function in `main.go` that upserts a trust entry in `~/.codex/config.toml` after writing project-scoped `.codex/config.toml`. Uses read-modify-write pattern matching `writeCodexMcpConfig`. Called from `runSetup()` inside the MCP-writing block.
