---
name: elixir-code-analysis
description: "Static code analysis for Elixir: module dependencies, change impact, data flow tracing, OTP inspection (Reach), code duplication detection (ExDNA), and AST pattern search/replace (ExAST). Use when exploring how modules relate, checking what breaks before a change, finding duplicated code, searching by AST pattern, or doing pattern-based refactoring."
---

# Elixir Code Analysis

Three complementary static analysis tools. No running server needed — they work on source files directly.

- **Reach** — dependency graphs, impact analysis, data flow, OTP inspection, dead code, smells
- **ExDNA** — code duplication / clone detection
- **ExAST** — AST-level search, replace, and diff

## Reach

### Inspect a target (start here)

```bash
# Function, module, file, or line — shows deps, impact, slices, context
mix reach.inspect MyApp.Accounts.create_user/2
mix reach.inspect MyApp.Accounts.create_user/2 --deps
mix reach.inspect MyApp.Accounts.create_user/2 --impact
mix reach.inspect MyApp.Accounts.create_user/2 --graph
```

### Project map (overview — heavier output)

```bash
mix reach.map                    # full project structure
mix reach.map --modules          # module inventory
mix reach.map --coupling         # coupling metrics (Ca, Ce, instability)
mix reach.map --hotspots         # high complexity × many callers
mix reach.map --depth            # module depth
mix reach.map --effects          # side-effect map
mix reach.map --boundaries       # boundary analysis
mix reach.map --data             # data-flow summaries
```

### Structural checks

```bash
mix reach.check                  # all checks
mix reach.check --dead-code      # unused pure expressions
mix reach.check --smells         # graph/effect/data-flow smells
mix reach.check --candidates     # refactoring candidates
mix reach.check --changed --base main  # changed functions since branch point
```

### Data flow tracing

```bash
# Taint: does untrusted input reach a sink?
mix reach.trace --from params --to write!
mix reach.trace --from input --to System.cmd

# Variable tracing
mix reach.trace --variable user --in MyApp.Accounts.create/1

# Slicing
mix reach.trace --backward lib/my_app/accounts.ex:45
mix reach.trace --forward lib/my_app/accounts.ex:45
```

### OTP analysis

```bash
mix reach.otp                    # all GenServers
mix reach.otp UserWorker         # specific module
mix reach.otp --concurrency      # Task/monitor/spawn topology
```

### HTML report

```bash
mix reach                        # interactive HTML with control flow, call graph, data flow
```

### Output formats

All reach commands support `--format json` or `--format text`.

## ExDNA — Duplication Detection

### Find clones

```bash
mix ex_dna                                    # whole project
mix ex_dna lib/my_app/accounts                # specific directory
mix ex_dna --min-mass 20                      # lower threshold (default: 30)
mix ex_dna --min-similarity 0.85              # near-miss detection (default: 1.0 = exact)
mix ex_dna --literal-mode abstract            # Type-II: ignore variable names
mix ex_dna --normalize-pipes                  # treat x |> f() same as f(x)
mix ex_dna --exclude-macro schema             # skip macro-generated code
mix ex_dna --format json                      # json, html, sarif also available
```

### Explain a specific clone

```bash
mix ex_dna.explain 1              # clone number from mix ex_dna output
mix ex_dna.explain 1 --min-mass 10
```

Shows anti-unification result, common structure, divergence points, and refactoring suggestion.

## ExAST — AST Search & Replace

### Search by pattern

```bash
mix ex_ast.search 'IO.inspect(_)'
mix ex_ast.search 'IO.inspect(_)' lib/
mix ex_ast.search '{:error, reason}' lib/ test/
mix ex_ast.search --count 'dbg(_)'                          # just count matches
mix ex_ast.search --limit 10 'Repo.get!(_, _)'              # cap results
```

#### Contextual filters

```bash
# Only inside specific constructs
mix ex_ast.search --inside 'def handle_call(_, _, _) do _ end' 'Repo.get!(_, _)'
mix ex_ast.search --not-inside 'test _ do _ end' 'IO.inspect(_)'

# Parent/ancestor/child filters
mix ex_ast.search 'IO.inspect(_)' --parent 'def _ do ... end'
mix ex_ast.search 'def name do ... end' --contains 'Repo.transaction(_)'

# Sibling ordering
mix ex_ast.search 'Repo.delete(record)' --follows 'record = Repo.get!(_, _)'

# Comment filtering
mix ex_ast.search 'def name do ... end' --comment-inside TODO
mix ex_ast.search 'def name do ... end' --comment-inside '/TODO|FIXME/'
```

### Replace by pattern

```bash
mix ex_ast.replace 'IO.inspect(expr, _)' 'expr' lib/
mix ex_ast.replace 'dbg(expr)' 'expr'
mix ex_ast.replace --dry-run 'IO.inspect(expr)' 'Logger.debug(inspect(expr))' lib/

# With context filters (same as search)
mix ex_ast.replace --not-inside 'test _ do _ end' 'IO.inspect(expr)' 'expr'
```

Always use `--dry-run` first to preview changes.

### AST diff

```bash
mix ex_ast.diff lib/old.ex lib/new.ex
mix ex_ast.diff --summary lib/old.ex lib/new.ex
```

### Pattern syntax

- Variables (`name`, `expr`) — capture any node (used in replacement)
- `_` or `_name` — wildcard (match, don't capture)
- Structs/maps — partial match (only listed keys must be present)
- Pipes normalized — `data |> Enum.map(f)` matches `Enum.map(data, f)`

## Guidance

- **Start targeted**: `reach.inspect` on the specific function you're working on
- **Escalate to overview** only when broader context is needed (`reach.map`, `reach.check`)
- **Combine tools**: reach finds the hotspot → ex_dna confirms duplication → ex_ast refactors it
- **Combine with Tidewave**: these tools do static analysis; use Tidewave for runtime state
