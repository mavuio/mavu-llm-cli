---
name: my-board
description: Personal issue board using bw (beadwork). Use when the user wants to track, plan, or review their own work items — not for repo task lists or coding todos.
---

## CRITICAL: Track All Work With BW

**BEFORE starting any task:**

1. FIRST: Run `bw prime` to get current workflow context and repo state
2. Create an issue: `bw create "Title" -t <type> -d "Description..."`
3. Start work: `bw start <id>` (moves to in_progress, assigns to you)
4. Do the work
5. Close when done: `bw close <id>`
6. WHEN COMMITTING: Include both code changes AND bw issue file(s) in the commit

If you identify follow-up work, create a new issue instead of doing it immediately.

Committing, closing issues, and syncing are part of completing a task — not separate actions requiring additional permission.

## Issue Types
{{range .Types}}

* **{{.Name}}**{{if .Description}}: {{.Description}}{{end}}
  {{- end}}

## Statuses
open, in_progress, closed, deferred

## Priorities
P0 (highest) through P4 (lowest). Use `-p` flag with 0–4 or P0–P4.

## CLI Reference

### Listing issues
```shell
bw list                                       # Open and in-progress (default limit 10)
bw list --all                                 # All issues, no filters
bw list -s closed --limit 5                   # By status
bw list -t bug -p 1                           # By type and priority
bw list -g "search term"                      # Search title and description
bw list --parent <id>                         # Children of an issue
bw list --label <label>                       # By label
bw list --deferred                            # Deferred issues only
bw list --overdue                             # Overdue issues only
bw list --json                                # Machine-readable output
```

Key flags: `-s` status, `-t` type, `-p` priority, `-g` search, `--label`, `--json`, `--limit`, `--all`

### Viewing an issue
```shell
bw show <id>                                  # Full details with formatting
bw show <id> --json                           # Machine-readable output
bw show <id> --only summary                   # Compact one-line summary
bw show <id> --only description,comments      # Specific sections
bw show <id> --only blockedby,unblocks        # Dependency context
```

### Creating issues
```shell
bw create "Title" -t <type> -d "Description"
bw create "Fix login bug" -p 1 -t bug
bw create "Q3 planning" --defer 2027-07-01
bw create "Ship v2" --due 2027-09-01
bw create "Subtask" --parent <id>
bw create "Quick task" --silent               # Output bare ID for scripting
```

Always specify `-t` type. Add detailed descriptions with `-d` when possible.

### Starting work
```shell
bw start <id>                                 # Move to in_progress, assign to self
bw start <id> -a alice                        # Assign to someone else
```

Refuses to start blocked issues.

### Updating issues
```shell
bw update <id> --title "New title"            # Change title
bw update <id> -d "New description"           # Change description
bw update <id> -p 0                           # Change priority
bw update <id> -s in_progress                 # Change status directly
bw update <id> --parent <other-id>            # Set parent
bw update <id> --due 2027-09-01               # Set due date
bw update <id> --defer 2027-06-01             # Set defer date
```

### Closing issues
```shell
bw close <id>                                 # Close an issue
bw close <id> --reason duplicate              # Close with reason
```

### Other commands
```shell
bw comment <id> "Fixed in latest deploy"      # Add a comment
bw label <id> +bug +urgent                    # Add labels
bw label <id> -wontfix                        # Remove labels
bw defer <id> 2 weeks                         # Defer until later
bw defer <id> "next monday"                   # Natural language dates
bw undefer <id>                               # Restore deferred issue to open
bw reopen <id>                                # Reopen a closed issue
bw delete <id>                                # Delete an issue
bw history <id>                               # Show issue history
```

### Finding work
```shell
bw ready                                      # Unblocked issues ready for work
bw blocked                                    # Issues waiting on dependencies
```

### Dependencies
```shell
bw dep add <id> blocks <other-id>             # This issue blocks another
bw dep remove <id> blocks <other-id>          # Remove dependency
```

### Sync & workflow context
```shell
bw prime                                      # Print workflow context (run before starting work)
bw sync                                       # Fetch, rebase/replay, push
bw export                                     # Export issues as JSONL
bw import <file>                              # Import issues from JSONL
```

## Working on an issue
1. Run `bw prime` to understand current state
2. Pick work: `bw ready` to find unblocked issues
3. Start: `bw start <id>`
4. Follow instructions in the issue description
5. If it has a checklist, update items as you complete them
6. Close: `bw close <id>`
7. Commit code and issue files together

## Dependencies
* **Parent**: `bw create "Subtask" --parent <id>` or `bw update <id> --parent <other-id>`
* **Blocking**: `bw dep add <id> blocks <other-id>` (this issue blocks another)
* Use `bw show <id> --only blockedby,unblocks` to see dependency context
