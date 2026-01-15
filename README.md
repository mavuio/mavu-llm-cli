# mavu-llm-cli

mavu-llm manages LLM setup files for a project (skills + agent docs).

## Install

Build the binary first:

```bash
go build -o mavu-llm
```

Option A: user-local symlink

```bash
mkdir -p ~/.local/bin
ln -sf /Users/manfred/Documents/www/mavu-llm-cli/mavu-llm ~/.local/bin/mavu-llm
```

Option B: Homebrew bin (system-wide)

```bash
ln -sf /Users/manfred/Documents/www/mavu-llm-cli/mavu-llm /opt/homebrew/bin/mavu-llm
```

## Usage

```bash
mavu-llm types
mavu-llm init --type <project-type> [--path <dir>]
mavu-llm update [--path <dir>]
```
