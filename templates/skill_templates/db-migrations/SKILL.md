---
name: db-migrations
description: Generate and run Ash/Phoenix database migrations with JSONB optimization
---

# Database Migrations

## Generate Migrations

Use the appropriate mix task to generate migrations:

```bash
mix ash.codegen <migration_name>
```

## Fix JSONB Types

After generating migrations, convert array of maps to JSONB for better performance:

```bash
rpl '{:array, :map}' ':jsonb' priv/repo/migrations/*
```

JSONB is preferred over PostgreSQL arrays for storing JSON objects because it supports better indexing and query performance.

## Run Migrations

```bash
mix ash.migrate
```


