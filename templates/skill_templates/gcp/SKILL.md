---
name: gcp
description: Git add, commit with appropriate message, and push (git add .; git commit; git push)
---

# Git Commit and Push

You are performing a quick git add, commit, and push operation.

## Instructions

1. First, run these commands in parallel to understand what will be committed:
   - `git status` to see all staged and unstaged changes
   - `git diff --cached` to see staged changes
   - `git diff` to see unstaged changes
   - `git log -3 --oneline` to see recent commit message style

2. Stage all changes:
   ```bash
   git add .
   ```

3. Analyze all changes and create a concise, appropriate commit message that:
   - Summarizes the nature of the changes (new feature, bug fix, refactor, docs, etc.)
   - Uses imperative mood (e.g., "Add feature" not "Added feature")
   - Is concise (1-2 sentences max)
   - Follows the repository's existing commit style if visible

4. Create the commit using a HEREDOC for proper formatting:
   ```bash
   git commit -m "$(cat <<'EOF'
   Your commit message here.

   Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
   EOF
   )"
   ```

5. Push to the remote:
   ```bash
   git push
   ```

6. Report the result to the user, including:
   - What was committed (summary of changes)
   - The commit message used
   - Push status

## Important Notes

- Do NOT push to main/master with --force
- If the push fails due to remote changes, inform the user and suggest `git pull --rebase` first
- Do NOT commit files that likely contain secrets (.env, credentials.json, etc.) - warn the user if such files are staged
- If there are no changes to commit, inform the user instead of creating an empty commit
