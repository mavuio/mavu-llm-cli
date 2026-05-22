# Stale Skill Cleanup Design

## Context

`mavu-llm-cli` writes skill templates into generated tool directories for target projects:

- `.codex/skills`
- `.claude/skills`
- `.agents/skills`

The current `createSkillDirs` flow detects existing child entries that are not in the resolved skill config. It warns that they will be removed, then asks whether to move them into `.mavu/skill_templates/` to keep them.

Example:

```text
Warning: 2 skill(s) not in config will be removed from agents: ash-framework, phoenix-framework
Move to .mavu/skill_templates/ to keep them? [y/N]:
```

This prompt is noisy and can block automated `init` or `update` runs.

## Goals

- Remove stale generated skill directories automatically during `llm init` and `llm update`.
- Remove the interactive prompt entirely.
- Keep support for `.mavu/skill_templates/` as the explicit preservation mechanism.
- Print a concise message when stale skills are removed.
- Keep the cleanup scoped to generated skill target directories only.

## Non-goals

- Do not add new CLI flags for stale-skill cleanup.
- Do not add a TOML policy setting.
- Do not move stale skills automatically into `.mavu/skill_templates/`.
- Do not change how desired skills are resolved from project type config plus local skill templates.

## Chosen Approach

Simplify `createSkillDirs` by removing the prompt and the optional move-to-local branch.

For each target (`codex`, `claude`, `agents`), the function will:

1. Resolve desired skills from the configured skills plus discovered local skills.
2. Read existing entries in the generated target skill directory.
3. Collect entries whose names are not desired.
4. If any stale entries exist, print a concise removal message.
5. Remove each stale entry with `os.RemoveAll`.
6. Copy desired skills from local or global templates as it does today.

## User Preservation Model

Users preserve custom skills by placing them in `.mavu/skill_templates/<skill-name>/` before running setup. Local skills are already discovered and merged into each target's desired skill set, so they will not be considered stale.

Generated target directories are treated as managed output. Anything in those directories that is not backed by config or `.mavu/skill_templates/` can be removed.

## Message Format

When stale skills are removed, print a single concise line per target, for example:

```text
Removed 2 stale skill(s) from agents: ash-framework, phoenix-framework
```

No stdin prompt should be emitted.

## Error Handling

- Directory read errors should continue to return errors as they do today.
- Removal errors from `os.RemoveAll` should return and stop setup.
- Missing target directories are still created before cleanup, as they are today.

## Tests

Add tests that verify:

- A stale skill directory under a generated target is removed during setup.
- A skill present in `.mavu/skill_templates/` is preserved and copied.
- The prompt text `Move to .mavu/skill_templates/ to keep them?` no longer appears in captured setup output.

Run `go test ./...` after implementation.
