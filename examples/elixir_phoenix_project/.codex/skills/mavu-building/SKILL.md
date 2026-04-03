---
name: mavu-building
description: Mavu UI/component building patterns — LiveView directory conventions, file naming, _mavubit shared libraries, reference codebase. Use when creating or modifying LiveViews, components, or UI code.
---

# Mavu Building Patterns

Use this skill when creating or modifying LiveViews, components, or UI code in a Mavu codebase.

## When to use this skill

- Creating new LiveViews or components
- Modifying existing UI code structure
- Working with `_mavubit` shared libraries
- Setting up new CRUD list/edit views
- Looking up reference implementations

## Reference codebase

When creating new resources, **always** look for existing patterns in the reference codebase at `_mavu_codesamples/samples/` first. Search for similar implementations, naming conventions, and structural patterns before writing new code from scratch. Stick as closely as possible to the style and structure found there.

## `lib/_mavubit/` — shared reusable libraries

`lib/_mavubit/` contains reusable libraries managed by [bit.dev](https://bit.dev) (a component management solution similar to Git subtree but more convenient). These components are shared across all mavu projects.

You **may** modify files in `lib/_mavubit/`, but changes **must** be generic and valuable for all projects using these components — not project-specific. When in doubt, ask whether the change should live in `_mavubit` or in the project's own code.

## LiveView directory conventions

Related LiveView files are co-located in a single directory. The main LiveView module lives **inside** that directory rather than as a sibling `.ex` file, even though the module name doesn't reflect the extra nesting.

```
lib/my_app_be/vocabulary_entries_live/
  00_vocabulary_entries_live.ex         ← main LiveView (module: MyAppBe.VocabularyEntriesLive)
  10_vocabulary_entry_list_lc.html.ex   ← list component
  100_vocabulary_entry_edit_component_lc.html.ex
```

### Numeric ordering prefix

Files are prefixed with a number (`00_`, `10_`, `100_`, …) to control sort order within the directory. The main entry-point file gets `00_`; supporting components follow in ascending order.

### `.html.ex` double extension

Files that contain HTML/template markup use the `.html.ex` extension instead of plain `.ex`, so Tailwind can find them. Plain `.ex` is reserved for files with no rendered markup.

### Module naming

The module name does **not** include the directory nesting. Example:
- File: `lib/my_app_be/films_live/00_films_live.ex`
- Module: `MyAppBe.FilmsLive` (not `MyAppBe.FilmsLive.FilmsLive`)

### Component suffix conventions

- `*_lc` — LiveComponent (e.g., `PersonListLc`, `PersonEditComponentLc`)
- `*_c` — Function component module (e.g., `SidebarComponents`)

## CRUD list/edit pattern

Each resource typically has:

1. **Main LiveView** (`00_*_live.ex`) — handles routing, params, top-level events
2. **List component** (`10_*_list_lc.html.ex`) — uses `MavuListAsh`/`MavuDatagrid` for data tables with filtering, sorting, pagination, bookmarks
3. **Edit component** (`100_*_edit_component_lc.html.ex`) — form with `MavuInputs.ash_input/1` for Ash-aware fields

### List component pattern

```elixir
def listconf(assigns) do
  %{
    resource: MyApp.SomeResource,
    action: :list_action,
    # column config, filters, etc.
  }
end

def load_items(assigns) do
  # query via MavuListAsh
end
```

### Edit component pattern

```elixir
def load_item(id, assigns) do
  MyApp.SomeResource.resolve!(id)
end
```

## Key UI modules

- **`MavuDatagrid`** — configurable data table component
- **`MavuInputs`** — Ash-aware form inputs (`ash_input/1`, `custom_input/1`)
- **`MavuListAsh`** — list query helpers (columns from Ash resource, pagination, streaming)
- **`MavuListAshfilter`** — filter form generation from Ash resource attributes
- **`MavuNav.SidebarComponents`** — sidebar navigation with pin/expand
- **`MavuNav.MenuComponents`** — menu rendering
- **`MavuSubforms`** — embedded/nested form handling for list-type attributes
