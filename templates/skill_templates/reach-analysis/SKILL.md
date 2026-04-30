---
name: reach-analysis
description: Use Reach for understanding module connections, function dependencies, change impact, and data flow. Use when exploring how modules relate, what a function depends on, what breaks if something changes, or tracing data through the system. Prefer targeted per-module/per-function commands over whole-codebase scans.
---

# Reach — Program Dependence Graph for Elixir

Reach builds a graph of what depends on what — data flow, control flow, side effects.
Use it for **targeted** analysis of specific modules and functions, not whole-codebase scans.

## When to use

- Understanding how a module connects to the rest of the system
- Checking what breaks before changing a function
- Tracing data flow from input to output
- Exploring callers/callees of a function
- Assessing coupling between modules

## Targeted commands (prefer these)

### Function dependencies — what calls this, what does it call?

```bash
mix reach.deps MyApp.SomeModule.some_function/2
mix reach.deps MyApp.SomeModule.some_function/2 --depth 2
mix reach.deps lib/my_app/some_module.ex:45
```

### Change impact — what breaks if I change this?

```bash
mix reach.impact MyApp.SomeModule.some_function/2
mix reach.impact MyApp.SomeModule.some_function/2 --depth 2
```

### Data flow — trace values through the system

```bash
# Where does user input end up?
mix reach.flow --from conn.params --to Repo
mix reach.flow --variable user --in MyApp.Accounts.create_user/1

# Taint: does untrusted input reach a dangerous sink?
mix reach.flow --from conn.params --to System.cmd
```

### Program slicing — minimum code affecting a value

```bash
# Backward slice: what affects the value at this line?
mix reach.slice lib/my_app/some_module.ex:45

# Forward slice: where does this value flow to?
mix reach.slice --forward lib/my_app/some_module.ex:30 --variable user

# Slice a specific variable in a function
mix reach.slice MyApp.SomeModule.create/1 --variable changeset
```

### Cross-function data flow

```bash
mix reach.xref --top 10
```

## Overview commands (use sparingly, heavy output)

### Module coupling metrics

```bash
# Full coupling table (Ca=afferent, Ce=efferent, I=instability)
mix reach.coupling --sort afferent

# Find orphan modules (nothing depends on them)
mix reach.coupling --orphans
```

### Module inventory

```bash
mix reach.modules --sort complexity
```

### Hotspots — high complexity × many callers

```bash
mix reach.hotspots
```

## Specialized analysis

### OTP patterns — GenServer state, ETS, missing handlers

```bash
mix reach.otp
```

### Code smells — redundant computations, fusible Enum chains

```bash
mix reach.smell
```

### Dead code

```bash
mix reach.dead_code
```

### Interactive HTML visualization

```bash
mix reach lib/my_app/some_module.ex
```

Generates a self-contained HTML with control flow, call graph, and data flow tabs.

### Terminal graph (requires boxart, already installed)

```bash
mix reach.graph MyApp.SomeModule.handle_call/3
mix reach.graph MyApp.SomeModule.handle_call/3 --call-graph
mix reach.deps MyApp.SomeModule.register/2 --graph
mix reach.impact MyApp.SomeModule.register/2 --graph
```

## JSON output

All commands support `--format json` for programmatic consumption:

```bash
mix reach.deps MyApp.SomeModule.func/2 --format json
mix reach.impact MyApp.SomeModule.func/2 --format json
```

## Guidance

- **Start targeted**: use `reach.deps` and `reach.impact` on the specific function you're working on
- **Escalate to overview** only when you need broader context (coupling, hotspots, modules)
- **Combine with tidewave**: Reach for static analysis, Tidewave for runtime state & live evaluation
