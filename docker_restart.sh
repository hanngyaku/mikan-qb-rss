#!/bin/sh
set -eu

cd "$(dirname "$0")"

mkdir -p data

echo "==> 拉取最新镜像"
docker compose -f docker-compose.ghcr.yml pull

echo "==> 更新并启动服务"
docker compose -f docker-compose.ghcr.yml up -d --remove-orphans

echo "==> 当前服务状态"
docker compose -f docker-compose.ghcr.yml ps
