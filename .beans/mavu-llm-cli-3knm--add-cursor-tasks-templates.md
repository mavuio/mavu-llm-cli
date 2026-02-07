---
# mavu-llm-cli-3knm
title: Add Cursor tasks templates
status: completed
type: feature
priority: normal
created_at: 2026-02-04T04:02:31Z
updated_at: 2026-02-04T04:06:23Z
---

Add session_templates support (global + .mavu) to generate .vscode/tasks.json from project type, with overwrite behavior and Phoenix dev tasks template.

## Tasks
- [x] Add session_templates support + tasks.json generation
- [x] Add Phoenix tasks template + project type config
- [x] Update tests/template root for session templates

## Summary of Changes
- Added session_templates config support to write .vscode/tasks.json from templates.
- Added Phoenix tasks template and wired session config to project type.
- Included session_templates path in template roots and test template setup.
