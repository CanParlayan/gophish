#!/usr/bin/env bash
# Deploy the tracked GoPhish branch to this server's existing production installation.
set -Eeuo pipefail

SRC="/home/pentester/gophish-src-v0.12.1"
PROD="/home/pentester/gophish"
BACKUPS="/home/pentester/gophish-prod-backups"
SERVICE="gophish"
REMOTE="fork"
BRANCH="feature/scenario-tracking"
DB="$PROD/gophish.db"

cd "$SRC"
[[ "$(git branch --show-current)" == "$BRANCH" ]] || { echo "ERROR: expected branch $BRANCH"; exit 1; }
[[ -z "$(git status --porcelain --untracked-files=no)" ]] || { echo "ERROR: tracked local changes exist; resolve before deploy"; exit 1; }
[[ -f "$PROD/config.json" && -f "$DB" && -x "$PROD/gophish" ]] || { echo "ERROR: production installation incomplete"; exit 1; }
command -v sqlite3 >/dev/null
command -v go >/dev/null
command -v sudo >/dev/null

echo "== Fetch and fast-forward =="
git fetch "$REMOTE" "$BRANCH"
git merge --ff-only "$REMOTE/$BRANCH"
echo "Deploying $(git rev-parse --short HEAD)"

# This server uses SQLite. Never deploy code requiring an unapplied schema change.
latest_sql=$(find "$SRC/db/db_sqlite3/migrations" -maxdepth 1 -type f -name '*.sql' -printf '%f\n' | sed -n 's/^\([0-9]\+\)_.*/\1/p' | sort -n | tail -1)
applied_sql=$(sudo sqlite3 "$DB" "SELECT COALESCE(MAX(version_id),0) FROM goose_db_version WHERE is_applied=1;")
if [[ -n "$latest_sql" ]] && (( 10#$latest_sql > 10#$applied_sql )); then
  echo "ERROR: pending SQLite migration $latest_sql (DB at $applied_sql). Review/apply migration separately."
  exit 1
fi

echo "== Test and build =="
go test ./models ./controllers
build_dir=$(mktemp -d)
trap 'rm -rf "$build_dir"' EXIT
go build -o "$build_dir/gophish" .

stamp=$(date -u +%Y%m%d-%H%M%S)
backup="$BACKUPS/deploy-$stamp"
sudo mkdir -p "$backup"
echo "== Backing up to $backup =="
sudo cp -a "$PROD/gophish" "$backup/gophish"
sudo cp -a "$PROD/templates" "$backup/templates"
sudo cp -a "$PROD/static" "$backup/static"
sudo sqlite3 "$DB" ".backup '$backup/gophish.db'"
[[ "$(sudo sqlite3 "$backup/gophish.db" 'PRAGMA integrity_check;')" == "ok" ]] || { echo "ERROR: invalid database backup"; exit 1; }

rollback() {
  echo "ERROR: deployment failed. Restoring binary and frontend from $backup"
  sudo systemctl stop "$SERVICE" || true
  sudo cp -a "$backup/gophish" "$PROD/gophish"
  sudo rm -rf "$PROD/templates" "$PROD/static"
  sudo cp -a "$backup/templates" "$PROD/templates"
  sudo cp -a "$backup/static" "$PROD/static"
  sudo systemctl start "$SERVICE" || true
  echo "Database left untouched; SQLite backup available at $backup/gophish.db"
}
trap rollback ERR

echo "== Install backend and frontend =="
sudo systemctl stop "$SERVICE"
sudo install -m 755 "$build_dir/gophish" "$PROD/gophish"
sudo cp -a "$SRC/templates/." "$PROD/templates/"
sudo cp -a "$SRC/static/." "$PROD/static/"
sudo systemctl start "$SERVICE"
sleep 3
sudo systemctl is-active --quiet "$SERVICE"
trap - ERR
echo "SUCCESS: $(git rev-parse --short HEAD) is running; backup: $backup"
sudo systemctl --no-pager --full status "$SERVICE"
