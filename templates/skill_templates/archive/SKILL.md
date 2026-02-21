---
name: archive
description: Archive the current opencode session by prefixing its title with "archive: ".
---

# Archive Session

Use this skill to archive the current opencode session title.

## What It Does

- Finds the most recently updated opencode session (current session in normal usage).
- Renames it by prefixing the title with `archive: `.
- Keeps it idempotent: if the title is already prefixed, it does not add it again.

## Preferred Command

```bash
# Archive current session title
./.codex/skills/archive/scripts/archive_current_session.sh

# Preview only
./.codex/skills/archive/scripts/archive_current_session.sh --dry-run
```

## Verify

```bash
opencode session list
```
