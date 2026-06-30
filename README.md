# Mikan qBittorrent RSS

Vue 3 + TypeScript + Go 的 qBittorrent RSS 管理工具（MVP 第一阶段）。

## 本地运行

要求 Go 1.24+、Node.js 24+。

```bash
go run ./cmd/server
cd web
npm install
npm run dev
```

前端地址为 `http://localhost:5173`，后端为 `http://localhost:8081`，Swagger UI 为 `http://localhost:8081/swagger/index.html`。

## Swagger 与前端 API 类型

```bash
make swagger
```

也可以在 PowerShell 中分别运行：

```powershell
go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g cmd/server/main.go -o docs --parseInternal
Set-Location web
npm run generate:api
```

## Docker

```bash
docker compose up --build
```

访问 `http://localhost:18081`。可复制 `.env.example` 为 `.env` 修改对外端口。

前端热更新调试：

```bash
docker compose -f docker-compose.dev.yml up --build
```

调试前端为 `http://localhost:5173`，调试后端为 `http://localhost:8081`。SQLite 数据保存在 `app-data` volume。

## 从 GHCR 部署

推送到 GitHub `main` 分支后，Actions 会将前端编译进 Go 运行镜像，并发布：

- `ghcr.io/<owner>/mikan-qb-rss:latest`

```bash
docker compose -f docker-compose.ghcr.yml pull
docker compose -f docker-compose.ghcr.yml up -d
```

`data` 目录及其中的 `app.db`、`app.log` 会在首次启动时自动创建。

也可以使用更新脚本：

```bash
chmod +x docker_restart.sh
./docker_restart.sh
```

脚本完全独立，不依赖 Compose 或 `.env`。运行前可直接修改脚本顶部的 `HOST_PORT` 和 `DATA_DIR`；它会拉取 `ghcr.io/hanngyaku/mikan-qb-rss:latest` 并替换旧容器。

Windows PowerShell：

```powershell
.\docker_restart.ps1
```

可在脚本顶部修改 `$HostPort` 和 `$DataDir`，同样不依赖 Compose 或 `.env`。

## 当前功能

添加订阅会拉取并解析 RSS、清理目录名、确保 qBittorrent 分类存在、创建 RSS feed 和下载规则，再写入 SQLite。订阅支持编辑、删除和重新同步。

设置页可配置默认排除正则；系统会记录最近一次创建/编辑订阅使用的排除正则，并在下次新建时自动填入。排除 `720` 或集数范围可使用 `720|\d+-\d+`，该值写入 qBittorrent `mustNotContain`。

每个订阅可设置 Season（默认 1），保存路径为 `名称/Season N`。后端每分钟检查已完成任务，从 `[12]` 形式的文件名提取集数，并通过 qBittorrent API 重命名为 `名称 S01 E12.mp4`；无法识别集数的文件不会修改。

运行日志同时输出到控制台和 `app.log`；Docker 中保存在数据卷的 `/app/data/app.log`。前端“日志”页面可按行数读取最新日志。
