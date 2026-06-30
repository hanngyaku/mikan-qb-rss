$ErrorActionPreference = "Stop"

# Deployment settings
$ContainerName = "mikan-qb-rss"
$ImageName = "ghcr.io/hanngyaku/mikan-qb-rss:latest"
$HostPort = 18081
$DataDir = Join-Path $PSScriptRoot "data"

New-Item -ItemType Directory -Path $DataDir -Force | Out-Null
$DataDir = (Resolve-Path $DataDir).Path

Write-Host "==> Pulling latest image"
docker pull $ImageName
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host "==> Replacing container"
docker rm -f $ContainerName 2>$null | Out-Null

docker run -d `
    --name $ContainerName `
    --restart unless-stopped `
    -p "${HostPort}:8081" `
    -e "DB_PATH=/app/data/app.db" `
    -e "LOG_PATH=/app/data/app.log" `
    -e "LISTEN_ADDR=:8081" `
    -v "${DataDir}:/app/data" `
    $ImageName

if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
Write-Host "==> Running at http://localhost:$HostPort"
