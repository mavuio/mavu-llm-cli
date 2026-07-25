---
name: phoenix-server
description: Use this skill to start and stop Phoenix development servers, and to create Git worktrees with mwt, including their reserved ports and provisioning.
---

<objective>
Run Phoenix development servers without port collisions and keep each project's main checkout and worktrees inside its reserved ten-port block.
</objective>

<essential_rules>
- Run `mpx prompt` for current `mpx` usage before managing a server.
- Never guess the main project port. Read `DEV_PORT` from `dc/.env` and verify the effective port with `tidewave-cli --port`.
- Project base ports are allocated ten apart and end in `0` (for example, `1120`).
- The main checkout owns the base port.
- Git worktrees for that project own the next nine ports in the same block (`1121` through `1129`).
- Never allocate a worktree from another project's block. For example, a worktree of the `1120` project must not use `1140`.
- `mpx newport` is for allocating a new project's ten-port block, not for allocating a worktree port.
- Create new worktrees with `mwt` (see below) — it names the branch, assigns the port, and provisions the worktree automatically. Only fall back to the manual workflow for worktrees not created with `mwt`.
</essential_rules>

<worktree_creation_with_mwt>
`mwt` is a zsh function (`~/.zsh/custom/mwt.zsh`) wrapping worktrunk (`wt`). It is the standard way to create and switch worktrees:

```text
mwt                         list worktrees (also: mwt list, mwt ls)
mwt <slug|PORT|wtPORT…>     switch to an existing worktree
mwt -n                      create wt<PORT> on the next free project port
mwt -n <slug>               create wt<PORT>-<slug> on the next free port
mwt -n <PORT>[-slug]        create wt<PORT>[-slug] (mwt -n 1125-be → wt1125-be)
mwt -n --db …               create with its own database (wt-own-db later)
mwt rm [<branch>]           remove worktree (delegates to wt remove)
```

Naming and ports: worktree branches are named `wt<PORT>[-slug]` (e.g. `wt1121`, `wt1122-redesign`). `PORT` is the full local `DEV_PORT` and must be in the project's worktree block (base `1120` → `1121` through `1129`). Short offset names such as `wt1-redesign` are not supported. Branches not matching `wt<PORT>` get a random free port when provisioned outside `mwt`.

Provisioning: on creation, worktrunk runs `~/.config/worktrunk/hooks/post-create.sh` (wired via the `pre-start` hook in `~/.config/worktrunk/config.toml`), which makes the worktree runnable immediately:

- copies gitignored secrets (`config/dev.secret.exs`, all `dc/.env*` files)
- symlinks `node_modules` (root, `assets/`, `be_assets/`) and `priv/data` to the main checkout
- rsyncs `_build` and `deps` so nothing recompiles from scratch
- writes `dc/.env.worktree` with the computed `DEV_PORT` and a `MAVU_WORKTREE=<branch>` marker
- logs to `/tmp/worktree-hook.log`

Database: worktrees share the main checkout's database by default. Create with `mwt -n --db <slug>` (or run `wt-own-db` inside an existing worktree) to get a separate database: this writes `MAVU_WORKTREE_DB=1`, which makes the project's env hooks derive a `-<worktree>`-suffixed `MAVU_PG_NAME`/`DATABASE_URL`, and a worktrunk `post-start` hook runs `mix ecto.setup` for it.

After creation, start the server as usual (`mpx start wt<PORT> -- mvsrv`) and verify the port with `tidewave-cli --port`.
</worktree_creation_with_mwt>

<worktree_port_workflow>
Manual fallback for worktrees not created with `mwt`. For a project whose main `DEV_PORT` is `1120`:

1. Treat `1120` as reserved for the main checkout.
2. Check `1121` through `1129` in ascending order and choose the first free port:

   ```bash
   lsof -nP -iTCP:1121 -sTCP:LISTEN
   ```

   No output means the port is currently free. Also account for other known worktrees whose servers may be stopped temporarily.

3. In the worktree, create or update `dc/.env.worktree` without changing the shared `dc/.env`:

   ```text
   DEV_PORT=1121
   ```

4. Load the project environment and verify the effective port:

   ```bash
   . mvp -q
   tidewave-cli --port
   ```

5. Start the server with `mpx`. Use a worktree-specific job/session name when needed to avoid ambiguity:

   ```bash
   mpx start wt1121 -- mvsrv
   ```

6. Verify the local listener and application before reporting the URL.
</worktree_port_workflow>

<public_url_mapping>
The secure development URL port is the local `DEV_PORT` plus `40000`:

- `1120` → `https://vm.mavu.io:41120`
- `1121` → `https://vm.mavu.io:41121`
- `1129` → `https://vm.mavu.io:41129`

Always report the URL corresponding to the worktree's effective port, not the main checkout's URL.
</public_url_mapping>

<cleanup>
- Stop the exact worktree server/session that was started.
- Keep `dc/.env.worktree` while the worktree actively uses that assigned port so `mvp`, `tidewave-cli`, and reload commands target the correct server.
- Do not commit `dc/.env.worktree` unless the project explicitly tracks worktree environment files.
</cleanup>

<success_criteria>
- Main checkout remains on the project's base port.
- Every running worktree uses a unique port from base + 1 through base + 9.
- `tidewave-cli --port` reports the assigned worktree port.
- The reported secure URL maps to that same port and responds successfully.
</success_criteria>
