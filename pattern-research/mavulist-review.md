# MavuList Package Review

Analysis of the MavuList hex package (v1.0.24) and its `_bit` Ash integration layer (`MavuListAsh`, `MavuListAshfilter`).

**Source:** `_bit/mavu_list_ashfilter/` (~2,729 lines) — the active codebase. The old hex package (`deps/mavu_list/`, ~1,086 lines) is **retired** and no longer maintained.

---

## Architecture Overview

```
┌─────────────────────────────┐
│  _bit/mavu_list_ashfilter/  │  Core Ash integration: filter forms, resource columns, presets,
│  ~2,729 lines               │  bulk edit, bookmarks, filter UI components
│  (MyAppBe.MavuListAsh)      │
│  (MyAppBe.MavuListAshfilter)│
└──────────────┬──────────────┘
               │
┌──────────────▼──────────────┐
│  _mavubit/be1/be/mavu_list/ │  UI components: pagination, searchbox, column chooser,
│  (shared across projects)   │  multi-item handler, tag filter, label, export
└─────────────────────────────┘
```

> **Note:** The old `deps/mavu_list` hex package is retired. All active code lives in the `_bit` and `_mavubit` layers.

---

## Flaws & Improvement Opportunities

### 🔴 P0 — Real Bugs / Security

#### 1. `String.to_atom/1` from user input in `get_preset_action_atom/1`

**File:** `_bit/mavu_list_ashfilter/mavu_list_ash.ex:167`

```elixir
def get_preset_action_atom(str) do
  try do
    preset_atom = String.to_atom("mavu_list_#{str}")  # ← unbounded atom creation
    {:ok, preset_atom}
  rescue
    ArgumentError -> {:error, "invalid preset: #{str}"}
  end
end
```

`str` comes from URL params (`context.params["preset"]`). `String.to_atom/1` never raises `ArgumentError` — it always succeeds, creating atoms that are never garbage-collected. An attacker sending random preset values will exhaust the atom table and crash the BEAM.

**Fix:** Use `String.to_existing_atom/1` (which *does* raise) or maintain an explicit allowlist of valid presets.

#### 2. `do_get_ids_for_query/2` loads up to 100 million records into memory

**File:** `_bit/mavu_list_ashfilter/mavu_list_ash.ex:185`

```elixir
|> Ash.Query.limit(100_000_000)
|> Ash.read!(page: false)
|> Enum.map(fn a -> a[:id] end)
```

When "select all pages" is used in bulk actions, this loads up to 100M records into memory at once. For large tables, this will OOM the server.

**Fix:** Use `Ash.stream!/2` (which is already available — see `get_stream_for_query/3` right below) or at least cap at a sane limit (e.g., 10,000) with a warning.

---

### 🟡 P1 — Significant Design Issues

#### 3. Two separate query executions per list load (N+1 for counts)

In `handle_data/2`, the flow is:
1. `apply_filter → apply_ash_filterform → autoload_fields → apply_sort` → builds filtered query
2. `apply_paging` → executes `Ash.read!(page: [limit: N, offset: M])` → **Query 1**
3. `update_metadata` → calls `get_length(source, conf)` → `Ash.read!(page: [limit: 1, count: true])` → **Query 2**

But Query 2 runs against the *original unfiltered source*, not the filtered query. This means:
- The `total_count` in metadata is **wrong** — it reflects the unfiltered count, not the count matching current filters
- Two round-trips to the DB per list load

**Fix:** Use Ash's built-in `count: true` option in the paging query itself, or run the count against `filtered_source` instead of `source`.

---

### 🟠 P2 — Code Quality / Maintainability

#### 6. Zero typespecs, zero tests, placeholder README

- No `@spec` or `@type` definitions anywhere in the package
- No test files at all (`test/` directory is empty)
- README still says "Eigenart — TODO: Add description"

Makes the package fragile for refactoring and hard for new contributors (or AI agents) to understand contracts.

#### 7. Dead/unused code

- `sort_by/3` (line 234) — takes `data`, `conf`, ignores third arg, returns `data` unchanged. Never called from project code.
- `handle_event_in_state/3` (3-arity catch-all, line 633) — only `IO.inspect`s unknown events, no error handling or logging

#### 8. Commented-out debug lines throughout

~18 commented-out debug lines across the codebase (`# |> IO.inspect(label: "mwuits-debug ...")`, `# |> MavuUtils.die(...)`). These are noise and make it harder to read the actual logic.

#### 9. `MavuList.Ash.get_filterform_module/1` ignores its argument

```elixir
def get_filterform_module(_listconf) do
  MyAppBe.MavuListAshfilter.FilterForm
end
```

Always returns the same hardcoded module regardless of `listconf`. The parameter exists but is never used — suggests an unfinished abstraction.

#### 10. Column visibility logic uses stringly-typed `hidden` field

```elixir
a when a in ["no", "false", false, :no] -> %{visible: true, editable: true}
a when a in ["yes", "true", true, :yes] -> %{visible: false, editable: true}
"never" -> %{visible: true, editable: false}
"always" -> %{visible: false, editable: false}
```

Mixes strings, booleans, and atoms for the same concept. No exhaustive match — if `hidden` is set to any other value, the function crashes. Should use a clean enum or boolean.

---

### 🔵 P3 — Minor / Nice-to-Have

#### 13. `@default_per_page` is hardcoded to 20

No way to configure a global default from `config.exs`. Projects must set `per_page` in every `listconf`.

#### 14. `default_sort/1` always returns `:asc`

Ignores the state entirely. Could read from `conf[:default_sort]` to allow per-resource defaults.

#### 15. URL tweaks encoding uses JSON strings in query params

Tweaks are JSON-encoded into URL params (`?items_filtered_tweaks={...}`). This makes URLs long and ugly. A more compact encoding (e.g., base64 or structured params) would be cleaner.

#### 16. `get_stream_for_query/3` exists but sorting is commented out

```elixir
# |> MavuList.apply_sort(state.conf, MavuList.get_sort_tweaks(state))
```

Streaming without sort may return results in unpredictable order.

---

## Summary

| Priority | # | Issue |
|----------|---|-------|
| 🔴 P0 | 1 | `String.to_atom/1` from user input — atom exhaustion DoS |
| 🔴 P0 | 2 | 100M record limit in `do_get_ids_for_query` — potential OOM |
| 🟡 P1 | 3 | Count query runs against unfiltered source (wrong totals) |
| 🟠 P2 | 4 | No typespecs, no tests |
| 🟠 P2 | 5 | Dead/unused code (`sort_by/3`, catch-all event handler) |
| 🟠 P2 | 6 | Commented-out debug lines (~18) |
| 🟠 P2 | 7 | `get_filterform_module/1` ignores its argument |
| 🟠 P2 | 8 | Stringly-typed `hidden` column visibility |
| 🔵 P3 | 9 | Hardcoded `@default_per_page` |
| 🔵 P3 | 10 | `default_sort/1` ignores state |
| 🔵 P3 | 11 | JSON-in-URL-params for tweaks |
| 🔵 P3 | 12 | Streaming without sort |
