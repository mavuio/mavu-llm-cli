# Shared Notes Symlink on Setup

## Context

`mavu-llm-cli` manages setup files for target projects through `runSetup`, which is used by both `llm init` and `llm update`. The update flow should create a `shared_notes` symlink inside the target project so project-specific notes can live under `/www/mavunotes/projects/<project-name>`.

Symlinks may point to targets that do not exist yet, so setup must not require the notes directory to be present.

## Goal

For a target project directory such as:

```text
/www/my-project
```

ensure setup creates:

```text
/www/my-project/shared_notes -> /www/mavunotes/projects/my-project
```

The project name is derived dynamically from the target project directory basename. The project `.gitignore` must also include `shared_notes` in the managed mavu-llm block so the symlink is not committed accidentally.

## Non-goals

- Do not add TOML configuration for the notes path in this change.
- Do not create `/www/mavunotes/projects/<project-name>`.
- Do not delete or replace a real file or directory named `shared_notes`.

## Design

Add a small helper called by `runSetup` so both `llm init` and `llm update` receive the behavior consistently.

The helper will:

1. Clean the target root directory path.
2. Compute `projectName` with `filepath.Base(rootDir)`.
3. Build the symlink path as `<rootDir>/shared_notes`.
4. Build the target path as `/www/mavunotes/projects/<projectName>`.
5. Use `os.Lstat` so broken symlinks are detected as existing symlinks.
6. If `shared_notes` is already a symlink to the expected target, do nothing.
7. If `shared_notes` is a symlink to a different target, remove and recreate it.
8. If `shared_notes` exists and is not a symlink, return an error to avoid data loss.
9. If `shared_notes` does not exist, create it with `os.Symlink` even when the target is missing.

## Error Handling

The helper returns contextual errors for failed `lstat`, `readlink`, `remove`, or `symlink` operations. A non-symlink conflict is reported as an error that tells the user the path already exists and is not a symlink.

## Testing

Add unit tests that verify:

- `runUpdate` creates `shared_notes` pointing to `/www/mavunotes/projects/<project-basename>` even when the target does not exist.
- Re-running the helper is idempotent when the symlink is already correct.
- An existing wrong symlink is replaced.
- An existing real directory or file at `shared_notes` is not removed and causes an error.
- The managed `.gitignore` block includes `shared_notes`.

## Self-review

- No placeholders remain.
- The target path convention is explicit and dynamic.
- The behavior for missing targets and conflicts is unambiguous.
- The scope is small enough for one implementation pass.
