---
# mavu-llm-cli-5j6h
title: Session status filter bar in TUI
status: completed
type: feature
priority: normal
created_at: 2026-02-22T09:49:42Z
updated_at: 2026-02-22T10:07:31Z
---

Replace showArchived bool with general-purpose status system. Add toggleable status chips for archive/future/delete with number key toggles.

## Todo
- [x] Define sessionStatus type and known statuses slice
- [x] Replace showArchived bool with statusVisibility map
- [x] Refactor shouldSkipSessionTitle to use sessionTitleStatus helper
- [x] Update renderOpenCodeSessionsTUI with status filter bar
- [x] Update key handling (remove v, add 1-3 number keys)
- [x] Update non-TUI filtering (CLI + API)
- [x] Update filterOpenCodeSessionsForTUI and sessionVisibleInTUI
- [x] Update tests in main_test.go
- [x] Verify: go test, go build, manual TUI test

## Summary of Changes
- Added `sessionStatus` type and `knownSessionStatuses` slice (archive, future, delete)
- Added `sessionTitleStatus` helper (matches `name:` prefix convention)
- Added `defaultStatusVisibility` helper
- Simplified `shouldSkipSessionTitle` to only check explore subagent
- Replaced `showArchived bool` with `statusVisibility map[string]bool` throughout TUI
- Updated TUI render to show status filter bar with chips
- Replaced v key with 1-3 number keys for status toggling
- Updated `findOpenCodeSessions` with `skipKnownStatuses` param
- Updated API handler to construct status visibility from `include_archived` param
- Updated all tests
