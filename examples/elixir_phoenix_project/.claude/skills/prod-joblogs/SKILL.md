---
name: prod-joblogs
description: Query joblogs in production via Flightcontrol LogEngine
---

# Production Joblogs Query

You are querying joblogs (Flightcontrol logs) on the production server.

## Instructions

1. Take the user's arguments from the skill invocation (everything after `/prod-joblogs`).
2. If no arguments were provided, ask for a day key (mon/tue/wed/thu/fri/sat/sun) or a full tid.
3. Build the tid:
   - If the argument starts with "joblog_", use it as-is.
   - Otherwise, treat it as a day key and use `"joblog_#{day_key}"`.
4. Choose logtype:
   - Default to `:default` unless the user specifies a logtype.
5. Execute the query on production:
   ```bash
   dokku enter web /bin/bash bin/iex.sh --eval "CODE_HERE"
   ```
6. Return the result to the user.

## Useful query patterns

- Latest N items (newest first):
  ```elixir
  MyApp.Flightcontrol.LogEngine.get_log_items({"joblog_sun", :default})
  |> Enum.take(10)
  ```

- Latest error payload:
  ```elixir
  items = MyApp.Flightcontrol.LogEngine.get_log_items({"joblog_sun", :default})
  error = Enum.find(items, &(&1[:level] == :error))
  error[:data]
  ```

- Count errors and show most recent:
  ```elixir
  items = MyApp.Flightcontrol.LogEngine.get_log_items({"joblog_sun", :default})
  errors = Enum.filter(items, &(&1[:level] == :error))
  %{error_count: length(errors), latest_error: List.first(errors)}
  ```

## Important Notes

- This runs on PRODUCTION. Avoid destructive operations.
- Joblogs are stored in Flightcontrol buckets, not in an Ecto table.
- Log items are newest-first because `LogEngine.add_to_log/4` prepends items.
