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

## Review Migrations

Always show the generated migration files for user inspection before proceeding:

```bash
cat priv/repo/migrations/*_<migration_name>.exs
```

**Never run migrations automatically.** Wait for explicit user approval after they have reviewed the migration files.

## Run Migrations

Once the user has approved the migrations:

```bash
mix ash.migrate
```


