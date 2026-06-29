$ErrorActionPreference = "Stop"
Set-Location $PSScriptRoot

New-Item -ItemType Directory -Path "data" -Force | Out-Null

Write-Host "==> 拉取最新镜像"
docker compose -f docker-compose.ghcr.yml pull
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host "==> 更新并启动服务"
docker compose -f docker-compose.ghcr.yml up -d --remove-orphans
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host "==> 当前服务状态"
docker compose -f docker-compose.ghcr.yml ps
exit $LASTEXITCODE
