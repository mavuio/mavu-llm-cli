---
# mavu-llm-cli-7p85
title: Session templates as single files
status: completed
type: task
priority: normal
created_at: 2026-02-04T04:10:49Z
updated_at: 2026-02-04T04:11:39Z
---

Stop merging multiple session templates; copy a single session template file to .vscode/tasks.json (one template per project type) like commands/snippets.

## Tasks
- [x] Update session template handling to copy a single file
- [x] Remove merge logic + parsing helpers

## Summary of Changes
- Switched session template handling to copy a single tasks template file directly to .vscode/tasks.json.
- Removed session template merge/parsing helpers and error on multiple session entries.
