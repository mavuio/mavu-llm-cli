---
name: beans
description: Use when managing tasks and issues with beans
---


## CRITICAL: Track All Work With Beans
**BEFORE starting any task:**

1. FIRST: Create a bean with `beans create "Title" -t <type> -d "Description..." -s in-progress`
2. THEN: Do the work
3. FINALLY: Mark completed with `beans update <bean-id> -s completed`
   
   * NEVER mark a bean "completed" if it has unchecked todo items!
4. WHEN COMMITTING: Include both code changes AND bean file(s) in the commit

If you identify follow-up work, create a new bean instead of doing it immediately.

## Issue Types
{{range .Types}}

* **{{.Name}}**{{if .Description}}: {{.Description}}{{end}}
  {{- end}}

## Statuses
{{range .Statuses}}

* **{{.Name}}**{{if .Description}}: {{.Description}}{{end}}
  {{- end}}

## Priorities
{{range .Priorities}}

* **{{.Name}}**{{if .Description}}: {{.Description}}{{end}}
  {{- end}}

## CLI Reference
### Listing beans
```shell
beans list                                    # All beans
beans list -s todo -s in-progress             # By status
beans list -t bug -p critical -p high         # By type and priority
beans list -S "search term"                   # Full-text search
beans list --is-blocked                       # Show blocked beans
beans list --parent <id>                      # Children of a bean
```

Key flags: `-s` status, `-t` type, `-p` priority, `-S` search, `--tag`, `--json`, `-q` (IDs only)

### Viewing a bean
```shell
beans show <id>                               # Full details with formatting
beans show <id> --json                        # Machine-readable output
```

### Creating beans
```shell
beans create "Title" -t <type> -d "Description" -s <status>
beans create "Fix bug" -t bug -d "Users cannot log in" -s todo
```

Always specify `-t` type. Add detailed descriptions with checklists when possible.

### Updating beans
```shell
beans update <id> -s in-progress              # Change status
beans update <id> --title "New title"         # Change title
beans update <id> --parent <other-id>         # Set parent
beans update <id> --blocking <other-id>       # Add blocking relationship
beans update <id> --remove-parent             # Remove parent
```

### Other commands
* `beans delete <id>` - Delete a bean
* `beans archive` - Delete all completed/scrapped beans (only when user asks)
* `beans roadmap` - Generate markdown roadmap from milestones

## Working on a bean
1. Read the bean: `beans show <id>`
2. Mark in-progress: `beans update <id> -s in-progress`
3. Follow instructions in the bean body
4. If it has a checklist, update items as you complete them (`- [ ]` → `- [x]`)
5. Mark completed: `beans update <id> -s completed`

## Relationships
* **Parent**: `beans update <id> --parent <other-id>` (hierarchy: milestone → epic → feature → task/bug)
* **Blocking**: `beans update <id> --blocking <other-id>` (this bean blocks another)

## GraphQL (Advanced)
For complex queries involving relationships, use GraphQL:

```shell
# Get bean with relationships
beans query '{ bean(id: "abc") { title parent { title } children { title status } blockedBy { title } } }'

# Find actionable beans (not blocked, not done)
beans query '{ beans(filter: { excludeStatus: ["completed", "scrapped"], isBlocked: false }) { id title } }'

# Full schema reference
beans query --schema
```
