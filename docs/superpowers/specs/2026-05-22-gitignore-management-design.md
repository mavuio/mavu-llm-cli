# Gitignore Management Design

## Context

`mavu-llm-cli` is a Go CLI that manages LLM setup files for target projects. The `init` and `update` commands both flow through `runSetup`, which writes skills, commands, session tasks, root docs, and `.mavu/config.toml` for the target project.

The CLI should also ensure target projects ignore internal/local agent artifacts that should not be committed.

## Goals

- During both `llm init` and `llm update`, ensure the target project's `.gitignore` contains a CLI-managed section.
- Start with these ignore entries exactly:
  - `.pi`
  - `.dexter.*`
  - `.bit`
- Create `.gitignore` if it does not exist.
- Preserve all user-authored `.gitignore` content outside the managed section.
- Make repeated runs idempotent.

## Non-goals

- Do not add project-type-specific gitignore configuration yet.
- Do not rewrite or sort the user's existing `.gitignore` entries.
- Do not remove matching entries outside the managed section.
- Do not add a separate CLI command for gitignore management yet.

## Chosen Approach

Add a focused gitignore helper called from `runSetup`.

The helper will:

1. Read `<target>/.gitignore` if present, or treat it as empty if missing.
2. Build a stable managed block using start/end markers and the configured default entries.
3. Replace the existing managed block when both markers are present.
4. Append the managed block when no managed block exists.
5. Write the resulting file back to `.gitignore`.

This approach keeps the feature centralized in the setup path without adding configuration schema or an extra command before either is needed.

## Managed Block Format

Use an explicit marker pair so the CLI only owns its own section:

```gitignore
# BEGIN mavu-llm managed
.pi
.dexter.*
.bit
# END mavu-llm managed
```

The formatter should ensure:

- The file ends with a newline.
- The managed block has exactly the entries listed above, one per line.
- There is a blank line between pre-existing content and an appended managed block when the existing file is non-empty.
- Replacing an existing managed block does not modify content before or after the block except for normalizing the block itself.

## Integration Point

Call `ensureManagedGitignore(rootDir)` inside `runSetup` after legacy MCP cleanup and before root docs/config writes.

That placement means:

- Both `init` and `update` get the behavior automatically.
- The operation is independent from template rendering.
- Any filesystem error stops setup consistently with other setup write failures.

## Error Handling

- Missing `.gitignore`: create it with the managed block.
- Read/write errors: return the underlying error and stop setup.
- Malformed managed block, such as a start marker without an end marker: append a fresh managed block rather than attempting to delete ambiguous user content. This avoids data loss while still ensuring required ignores exist.

## Tests

Add unit tests for the helper:

- Creates `.gitignore` when missing.
- Appends the managed block to an existing user `.gitignore`.
- Replaces a stale managed block with the current entries.
- Preserves user content outside the managed block.
- Is idempotent across repeated calls.

Add integration coverage for setup:

- `runInit` creates or updates `.gitignore` in the target project.
- `runUpdate` keeps the managed block current.

Note: the current test suite already has unrelated failing tests because the PHP project template now lists `my-board` while older tests still expect `beans`. The gitignore tests should be written independently, and the unrelated fixture mismatch can be addressed separately if needed.
