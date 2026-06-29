$ErrorActionPreference = "Stop"
Set-Location $PSScriptRoot

New-Item -ItemType Directory -Path "data" -Force | Out-Null

Write-Host "==> Pulling latest image"
docker compose -f docker-compose.ghcr.yml pull
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host "==> Updating and starting service"
docker compose -f docker-compose.ghcr.yml up -d --remove-orphans
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host "==> Service status"
docker compose -f docker-compose.ghcr.yml ps
exit $LASTEXITCODE
