---
name: mavu
description: Tools and conventions for general tooling in a mavu-style phoenix codebase 
---

# Mavu Code Patterns

Use this skill when you need to recognize or apply Mavu-specific conventions. Examples below are pulled from the current codebase.

## Reference codebase

When creating new code, **always** look for existing patterns in the reference codebase at `/www/mavu_codesamples/samples` first. Search for similar implementations, naming conventions, and structural patterns before writing new code from scratch.

## `lib/_mavubit/` — shared reusable libraries

`lib/_mavubit/` contains reusable libraries managed by [bit.dev](https://bit.dev) (a component management solution similar to Git subtree but more convenient). These components are shared across all mavu projects.

You **may** modify files in `lib/_mavubit/`, but changes **must** be generic and valuable for all projects using these components — not project-specific. When in doubt, ask whether the change should live in `_mavubit` or in the project's own code.

## When to use this skill

- Adding or updating Mavu resources/tools that should follow established patterns
- Reviewing code that relies on Mavu utilities like resolve helpers, buckets, or log helpers
- Answering questions about common Mavu utilities and macros

## Entity resolution

### MavuEntities.resolve!/2

Resolve any entity by its web ID (human-friendly ID with prefix).

Example:

```elixir
foo = MyApp.MavuEntities.resolve!("ag SNTX0")
```

Used in:

- lib/my_app/tools/tool_execute_trade.ex
- scripts/test_trigger_evaluation.exs

### webid format

A webid is a URL-safe identifier: `prefix_base62UUID` (e.g., `ag_032E9nHv1zpA2byKZd18qt`).

- **prefix**: entity type (`ag` for Foo, `tsk` for Task, etc.)
- **base62UUID**: 22-char encoded UUID

Defined in: lib/my_app/_bit/mavu_entities/

### Resource resolve/resolve!/resolve0

Each Ash resource exposes resolve helpers to normalize IDs or structs.

Example (Task resource):

```elixir
def resolve!(any, opts \\ []), do: resolve(any, opts) |> MavuUtils.unwrap!()
def resolve0(any, opts \\ []), do: resolve(any, opts) |> MavuUtils.unwrap0()
```

Resolution can be done by id, webid, and any identity of the resource.
if you pass a resource-struct, it will be returned as is.

### Resource resolve/resolve!/resolve0

Each Ash resource exposes resolve helpers to normalize IDs or structs.

**Resolution accepts:**
- UUID (`"019bd03a-2094-739b-b5bf-db2b08ea34b8"`)
- WebID (`"ag_032CLtNxtjfzxBQtv1dbya"`)
- Any identity defined on the resource (e.g., `name`, `ticker`)
- The resource struct itself (returned as-is)

**Prefer resource-specific resolve over MavuEntities.resolve!** when you know the resource type.

Example:

```elixir
# By name identity
MyApp.Foo.resolve!("Demo Ledger Foo")

# By ticker identity  
MyApp.Foo.resolve!("DEMO_LEDGER")

# By UUID
MyApp.Foo.resolve!("019bd03a-2094-739b-b5bf-db2b08ea34b8")

# By webid
MyApp.Foo.resolve!("foo_032CLtNxtjfzxBQtv1dbya")
Use resolve/1 for {:ok, result} tuples, resolve!/1 to unwrap or raise, resolve0/1 to unwrap or return nil.


```


## Result tuple handling

### MavuUtils.unwrap0/1

Unwraps `{:ok, value}` to `value`, returns nil on error.

Example:

```elixir
def resolve0(any, opts \\ []), do: resolve(any, opts) |> MavuUtils.unwrap0()
```

## Map/struct access

### MavuUtils.any_get0/2

Flexible getter that works with maps, structs, keyword lists, and atom/string keys. Also loads lazy Ash attributes (calculations, aggregates, relationships) automatically using `Ash.load` internally.

Example:

```elixir
MavuBuckets.get_value("foo_job:#{foo.id}", "state", %{}, opts)
|> MavuUtils.any_get0(key)
```

Example (loading lazy Ash attribute):

```elixir
# Loads :webid calculation on the fly
foo |> MavuUtils.any_get0(:webid)
```

## Shorthand m/1 macro

`m/1` (from `Shorthand`) builds or pattern-matches maps using same-name keys. It is commonly used in Reactor steps.

Example:

```elixir
run fn m(foo), _context ->
  MyApp.Foos.get_public_key_for_foo(foo)
end
```


## assign_new with m/1

Use `assign_new/3` with `m/1` to pull prior assigns without repeating key names.

Example:

```elixir
assigns
|> assign_new(:foo, fn m(context) -> foo_data(context) end)
|> assign_new(:performance_snapshot, fn m(foo) -> latest_performance_snapshot(foo) end)
```

## happy_with Macro

Syntactic sugar for Elixir's `with` that keeps left-arrow chains tidy. The key feature is **tagged pattern matching** using `@tag` - when a step fails, you know exactly which one.

### Tagged Pattern Matching

```elixir
happy_with do
  @run_action {:ok, results} <- run_action(agent_id, payload)
  @log_result _ <- results |> MavuUtils.log("results", :info)
  {:ok, results}
else
  {error_tag, error_context} ->
    # error_tag is :run_action or :log_result - identifies which step failed
    {:error, {error_tag, msg, error_context}}
end
```

### Key Benefits

- **Tags identify failed step** - Errors come as `{tag, original_error}` in the `else` clause
- **Cleaner syntax** - Last expression is the return value automatically
- **Guard support** - `@tag {:ok, x} when is_binary(x) <- expr`

## MavuBuckets

Key-value store backed by the database with PubSub support for LiveView.
Garbage collection is handled automatically by the database.

Example:

```elixir
MavuBuckets.set_value("foo:#{id}", "iteration", 0, persistence_level: 100)
MavuBuckets.subscribe_live_view("foo_detail:" <> assigns.foo.id)
```


Note: `MavuBuckets.bucket_cache/4` appears only in docs at the moment and is not referenced in app code.

## Logging

### MavuUtils.log/3

**Always** use MavuUtils.log instead of Logger.* It is a wrapper around Logger.log/3.

Example:

```elixir
params
|> MavuUtils.log("#clgreen CodeReloaderController", :info)
```

