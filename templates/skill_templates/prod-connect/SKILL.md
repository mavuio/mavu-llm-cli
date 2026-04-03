---
name: prod-connect
description: Connect to the production server via dokku or ssh
---


## Connect via dokku

if remote dokku is present:

use  `dokku enter web` 
to connect to dokku 

## Connect via ssh

use ssh, get credentials from

.mwproject

e.g:

MW_TARGET=root@www.example.com:/www/foo/

## Sync changed files to prod

Use `mavu_scp` from inside the project repo. It reads `MAVU_TARGET` from the environment and, if needed, tries to source `mvp/mvp` automatically.

Typical usage:

- Explicit files:
  - `mavu_scp -- lib/foo.ex assets/js/app.js`
- Tracked dirty files:
  - `mavu_scp --dirty-files`
- Watch all tracked files and sync on change:
  - `mavu_scp_auto`
  - or `mavu_scp --watch`

Optional filters:

- `--match <regex>` to include only matching files
- `--skip <regex>` to skip matching files

Notes:

- Target comes from `MAVU_TARGET`; you do not pass it manually.
- After a successful copy, the tool also runs `recompile_on_server/recompile_on_server.sh` and plays the pop sound.
- Built-in skips include `mix-manifest.json` and `.idea`.
- `mavu_scp_auto` is an alternate entrypoint to the same Go binary, defaulting to watch mode.
