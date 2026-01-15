---
name: prod-eval
description: Evaluate Elixir code on the production server via dokku
---

# Production IEx Evaluation

You are evaluating Elixir code on the production server.

## Instructions

1. Take the user's Elixir code from the skill arguments (everything after `/prod-eval`)

2. If no code was provided, ask the user what code they want to evaluate

3. Execute the code on the production server:
   ```bash
   dokku enter web /bin/bash bin/iex.sh --eval "CODE_HERE"
   ```

4. Return the result to the user

## Important Notes

- This runs on PRODUCTION - be careful with destructive operations
- The code is executed via RPC on the running Phoenix node
- Complex code with quotes may need proper escaping
- For multi-line code, consider using parentheses to group expressions

## Examples

Simple evaluation:
```bash
dokku enter web /bin/bash bin/iex.sh --eval "1 + 1"
```

Query example:
```bash
dokku enter web /bin/bash bin/iex.sh --eval "MyApp.Repo.all(MyApp.User) |> length()"
```

Function call:
```bash
dokku enter web /bin/bash bin/iex.sh --eval "MyApp.SomeModule.some_function(:arg)"
```
