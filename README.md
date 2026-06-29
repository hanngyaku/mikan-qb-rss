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

访问 `http://localhost:8080`。可复制 `.env.example` 为 `.env` 修改对外端口。

前端热更新调试：

```bash
docker compose -f docker-compose.dev.yml up --build
```

调试前端为 `http://localhost:5173`，调试后端为 `http://localhost:8081`。SQLite 数据保存在 `app-data` volume。

## 第一阶段边界

添加订阅会拉取并解析 RSS、清理目录名并写入 SQLite；qBittorrent RSS feed/rule 写入、订阅编辑/删除/同步留到下一阶段。
