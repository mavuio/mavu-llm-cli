---
name: deploy
description: Run tests, bump version, build, commit with a clean release message, and push.
---

# Deploy Workflow

version: 1.0.0

Use this skill to ship `mavu-llm-cli` changes safely.

## Steps

1. Check repo state first:
   ```bash
   git status --short
   git diff
   git log -5 --oneline
   ```

2. Choose version bump type (`patch` by default unless user asks for `minor` or `major`).

3. Update the `version` constant in `main.go`.

4. Run formatting and verification:
   ```bash
   gofmt -w main.go main_test.go
   go test ./...
   go build -o mavu-llm .
   ```

5. Stage only relevant source/docs changes (never stage the built binary):
   ```bash
   git add main.go main_test.go README.md
   ```
   Only add files that actually changed.

6. Commit with a clear release-style message:
   - Title format: `release: bump version to vX.Y.Z`
   - Optional second sentence explaining the primary user-facing change.

7. Push current branch:
   - If upstream exists: `git push`
   - If not: `git push -u origin <branch>`

8. Report back:
   - New version
   - Commit hash and message
   - Branch pushed

## Guardrails

- Stop immediately if tests fail; report errors instead of pushing.
- Never use force push.
- Never include unrelated files in the release commit.
