$ErrorActionPreference = 'Stop'

function Require($cmd, $installHint) {
  if (-not (Get-Command $cmd -ErrorAction SilentlyContinue)) {
    Write-Error "Command not found: $cmd`n$installHint"
    exit 1
  }
}

Write-Host "==> Checking prerequisites" -ForegroundColor Cyan
Require 'docker' "Install Docker Desktop and ensure it's running."
Require 'supabase' "Install Supabase CLI: https://supabase.com/docs/guides/local-development"
Require 'go' "Install Go: https://go.dev/dl/"
Require 'npm' "Install Node.js: https://nodejs.org/en/download"

$root = Split-Path -Parent $MyInvocation.MyCommand.Path
$supabaseDir = Join-Path $root 'supabase'

Write-Host "==> Starting Supabase (Docker)" -ForegroundColor Cyan
Push-Location $supabaseDir
try {
  $dbUp = (Test-NetConnection 127.0.0.1 -Port 54322 -WarningAction SilentlyContinue).TcpTestSucceeded
  if ($dbUp) {
    Write-Host "DB already reachable on 54322; skipping 'supabase start'" -ForegroundColor Yellow
  } else {
    & supabase start
    $exitCode = $LASTEXITCODE
    if ($exitCode -ne 0) {
      Write-Warning "'supabase start' failed (exit $exitCode). Trying 'supabase stop' then retry..."
      & supabase stop
      & supabase start
      $exitCode = $LASTEXITCODE
      if ($exitCode -ne 0) {
        Write-Error @"
Supabase failed to start (exit $exitCode).
Common causes:
  - Docker container name conflict ("name is already in use").
  - Stale containers from previous runs.

Try:
  1) Run: 'supabase stop'
  2) If it persists, remove stale containers:
     docker ps -a --format "{{.ID}}`t{{.Names}}" | findstr kintai
     docker rm -f <ID or Name>
  3) Then rerun: npm run dev
"@
        exit 1
      }
    }
  }
} finally {
  Pop-Location
}

Write-Host "==> Waiting for DB on 127.0.0.1:54322" -ForegroundColor Cyan
$deadline = (Get-Date).AddMinutes(2)
do {
  $ready = (Test-NetConnection 127.0.0.1 -Port 54322 -WarningAction SilentlyContinue).TcpTestSucceeded
  if (-not $ready) { Start-Sleep -Seconds 2 }
} until ($ready -or (Get-Date) -gt $deadline)
if (-not $ready) {
  Write-Error "Database port 54322 is not reachable. Aborting."
  exit 1
}

Write-Host "==> Launching Backend (Go)" -ForegroundColor Cyan
$backendProc = Start-Process -FilePath "powershell" -ArgumentList @(
  '-NoExit',
  '-Command',
  "cd `"$root/backend`"; if (Test-Path .env) { Write-Host 'Using backend/.env' } else { Write-Warning 'backend/.env not found' }; go run main.go"
) -PassThru

Write-Host "==> Launching Frontend (Next.js)" -ForegroundColor Cyan
$frontendProc = Start-Process -FilePath "powershell" -ArgumentList @(
  '-NoExit',
  '-Command',
  "cd `"$root/frontend`"; if (-not (Test-Path node_modules)) { Write-Host 'Installing frontend deps...'; npm install }; npm run dev"
) -PassThru

Write-Host "" 
Write-Host "All set! Services are starting in separate terminals:" -ForegroundColor Green
Write-Host "- Frontend:  http://localhost:3000"
Write-Host "- Backend:   http://localhost:8080 (health: /health)"
Write-Host "- Supabase:  http://127.0.0.1:54323 (Studio)"
Write-Host "" 
Write-Host "Close this window or press Ctrl+C to stop Supabase after child windows exit." -ForegroundColor Yellow

# Ensure Supabase stops when this host exits unexpectedly
$null = Register-EngineEvent -SourceIdentifier 'PowerShell.Exiting' -Action ([ScriptBlock]::Create(@"
try {
  Push-Location '$supabaseDir'
  supabase stop | Out-Null
} catch {}
try { Pop-Location } catch {}
"@))

# Wait for child windows; on completion or interruption, stop Supabase
try {
  Wait-Process -Id $backendProc.Id, $frontendProc.Id
} finally {
  try {
    Push-Location $supabaseDir
    supabase stop | Out-Null
  } catch {}
  try { Pop-Location } catch {}
}
