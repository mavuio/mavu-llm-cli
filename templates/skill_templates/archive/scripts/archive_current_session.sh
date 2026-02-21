#!/usr/bin/env bash

set -euo pipefail

usage() {
  cat <<'EOF'
Archive the current opencode session title.

Usage:
  archive_current_session.sh [--dry-run]

Options:
  -n, --dry-run  Print what would change without writing.
  -h, --help     Show this help.
EOF
}

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    printf "Error: required command not found: %s\n" "$1" >&2
    exit 1
  fi
}

DRY_RUN=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    -n|--dry-run)
      DRY_RUN=1
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    -*)
      printf "Error: unknown option: %s\n" "$1" >&2
      usage >&2
      exit 1
      ;;
    *)
      printf "Error: unexpected argument: %s\n" "$1" >&2
      usage >&2
      exit 1
      ;;
  esac
done

require_cmd opencode
require_cmd python3

DB_PATH="$(opencode db path)"

if [[ -z "$DB_PATH" || ! -f "$DB_PATH" ]]; then
  printf "Error: opencode database not found: %s\n" "$DB_PATH" >&2
  exit 1
fi

SESSION_ID="$(python3 - "$DB_PATH" <<'PY'
import sqlite3
import sys

db_path = sys.argv[1]
conn = sqlite3.connect(db_path)
row = conn.execute("SELECT id FROM session ORDER BY time_updated DESC LIMIT 1").fetchone()
print(row[0] if row else "")
PY
)"

OLD_TITLE="$(python3 - "$DB_PATH" <<'PY'
import sqlite3
import sys

db_path = sys.argv[1]
row = sqlite3.connect(db_path).execute(
    "SELECT title FROM session ORDER BY time_updated DESC LIMIT 1"
).fetchone()
print(row[0] if row and row[0] is not None else "")
PY
)"
PREFIX="archive: "

if [[ -z "$SESSION_ID" ]]; then
  printf "Error: no session found to archive\n" >&2
  exit 1
fi

if [[ "$OLD_TITLE" == "$PREFIX"* ]]; then
  NEW_TITLE="$OLD_TITLE"
else
  NEW_TITLE="$PREFIX$OLD_TITLE"
fi

if [[ "$DRY_RUN" -eq 1 ]]; then
  printf "[dry-run] session: %s\n" "$SESSION_ID"
  printf "[dry-run] old title: %s\n" "$OLD_TITLE"
  printf "[dry-run] new title: %s\n" "$NEW_TITLE"
  exit 0
fi

if [[ "$NEW_TITLE" == "$OLD_TITLE" ]]; then
  printf "Session already archived: %s\n" "$SESSION_ID"
  printf "Title: %s\n" "$OLD_TITLE"
  exit 0
fi

python3 - "$DB_PATH" "$SESSION_ID" "$NEW_TITLE" <<'PY'
import sqlite3
import sys

db_path, session_id, new_title = sys.argv[1], sys.argv[2], sys.argv[3]

conn = sqlite3.connect(db_path)
cur = conn.execute(
    """
    UPDATE session
    SET title = ?,
        time_updated = CAST(strftime('%s','now') AS INTEGER) * 1000
    WHERE id = ?
    """,
    (new_title, session_id),
)
conn.commit()

if cur.rowcount != 1:
    print(f"Error: update affected {cur.rowcount} rows", file=sys.stderr)
    sys.exit(1)
PY

printf "Archived session: %s\n" "$SESSION_ID"
printf "Old title: %s\n" "$OLD_TITLE"
printf "New title: %s\n" "$NEW_TITLE"
