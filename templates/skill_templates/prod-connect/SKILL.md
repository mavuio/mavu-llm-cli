---
name: prod-connect
description: Connect to the production server via dokku or ssh
---

## Connect via dokku

if remote dokku is present:

use `dokku enter web`
to connect to dokku

## Connect via ssh

use ssh, get credentials from

.mwproject

e.g:

MW_TARGET=root@www.example.com:/www/foo/
