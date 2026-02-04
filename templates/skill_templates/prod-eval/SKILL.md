---
name: prod-eval
description: Evaluate Elixir code on the production server via dokku
---

# Production IEx Evaluation

You are evaluating Elixir code on the production server.

## Instructions

1. Take the user's Elixir code from the skill arguments (everything after `/prod-eval`)

2. If no code was provided, ask the user what code they want to evaluate

3. Base64-encode the Elixir code and execute it on the production server using `--eval-b64`:
   ```bash
   dokku enter web /bin/bash bin/iex.sh --eval-b64 "BASE64_CODE_HERE"
   ```

4. Return the result to the user

## Important Notes

- This runs on PRODUCTION - be careful with destructive operations
- The code is executed via RPC on the running Phoenix node
- Always use `--eval-b64` (do not use `--eval` or `--eval-stdin`)
- Base64 encoding avoids quoting/escaping issues and supports multi-line code

## Examples

Simple evaluation (manual base64):
```bash
python - <<'PY'
import base64
code = "1 + 1"
print(base64.b64encode(code.encode()).decode())
PY

dokku enter web /bin/bash bin/iex.sh --eval-b64 "BASE64_CODE_HERE"
```

Query example:
```bash
python - <<'PY'
import base64
code = "MyApp.Repo.all(MyApp.User) |> length()"
print(base64.b64encode(code.encode()).decode())
PY

dokku enter web /bin/bash bin/iex.sh --eval-b64 "BASE64_CODE_HERE"
```

Function call:
```bash
python - <<'PY'
import base64
code = "MyApp.SomeModule.some_function(:arg)"
print(base64.b64encode(code.encode()).decode())
PY

dokku enter web /bin/bash bin/iex.sh --eval-b64 "BASE64_CODE_HERE"
```
