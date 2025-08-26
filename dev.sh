#!/usr/bin/env bash
set -euo pipefail

command -v docker >/dev/null 2>&1 || { echo >&2 "docker is required"; exit 1; }
command -v supabase >/dev/null 2>&1 || { echo >&2 "supabase CLI is required"; exit 1; }
command -v go >/dev/null 2>&1 || { echo >&2 "go is required"; exit 1; }
command -v npm >/dev/null 2>&1 || { echo >&2 "npm is required"; exit 1; }

root_dir="$(cd "$(dirname "$0")" && pwd)"

echo "==> Starting Supabase (Docker)"
pushd "$root_dir/supabase" >/dev/null
supabase start
popd >/dev/null

echo "==> Waiting for DB on 127.0.0.1:54322"
deadline=$((SECONDS + 120))
db_ready=false
while [ $SECONDS -lt $deadline ]; do
  if (echo > /dev/tcp/127.0.0.1/54322) >/dev/null 2>&1; then
    db_ready=true; break
  fi
  sleep 2
done
if [ "$db_ready" != true ]; then
  echo "Database port 54322 is not reachable. Aborting." >&2
  exit 1
fi

echo "==> Launching Backend (Go)"
(cd "$root_dir/backend" && go run main.go) &
backend_pid=$!

echo "==> Launching Frontend (Next.js)"
(cd "$root_dir/frontend" && [ -d node_modules ] || npm install && npm run dev) &
frontend_pid=$!

cleanup() {
  echo "\nStopping dev processes..."
  kill $backend_pid $frontend_pid 2>/dev/null || true
}
trap cleanup EXIT INT TERM

echo
echo "All set! Services are starting:"
echo "- Frontend:  http://localhost:3000"
echo "- Backend:   http://localhost:8080 (health: /health)"
echo "- Supabase:  http://127.0.0.1:54323 (Studio)"
echo
wait $backend_pid $frontend_pid

