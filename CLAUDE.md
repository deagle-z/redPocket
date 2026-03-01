# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

> **This file covers the Go backend only.** Frontend sub-projects each have their own CLAUDE.md:
> - [pure-admin-thin/CLAUDE.md](pure-admin-thin/CLAUDE.md) — Superadmin dashboard (Element Plus)
> - [RedTenantAdmin/CLAUDE.md](RedTenantAdmin/CLAUDE.md) — Tenant management panel (Element Plus)
> - [RedPocketH5/CLAUDE.md](RedPocketH5/CLAUDE.md) — Mobile H5 app (Vant)

## Commands

```bash
# Set up Go proxy (for Linux/China servers)
export GOPROXY=https://goproxy.cn,direct
export GONOSUMDB=*

# Initialize vendor dependencies
go mod vendor

# Install and generate Swagger docs
go install github.com/swaggo/swag/cmd/swag@latest
swag init

# Run the application (configs default to core.yaml, cs.yaml, sc.yaml)
go run main.go
go run main.go -cc core.yaml -cs cs.yaml -sc sc.yaml

# Run tests
go test -v ./core/services -run TestGenerateThunderIndexes
go test -v ./core/utils/...

# Build and run with Docker
docker rm -f bgu-1 && docker rmi bgu-1 && docker build --build-arg BUILDKIT_INLINE_CACHE=1 --memory 1GB -t bgu-1 . && docker run -e TZ=Asia/Shanghai -p 9001:8080 --name bgu-1 --restart always -d bgu-1 && docker logs -t -f bgu-1
```

The app runs on port `8080` by default (mapped to host port `9001` in Docker). Swagger UI is available at `/swagger/index.html`.

## Architecture

### Layered Structure
```
API handler → Repository → GORM (MySQL)
                         → Redis (cache/locks)
                         → RabbitMQ (async events)
```

- **`core/`** — Framework-level code (reusable across tenants/hosts)
- **`app/`** — Application-specific extensions (Telegram bot, app-layer services)
- **`tenant/`** — Tenant-specific API handlers (mirrors core/api for tenant isolation)

### Multi-Tenancy via Table Prefix
Every request goes through `hostInfoMiddleware` which:
1. Resolves the incoming host to a `HostInfo` record (cached in Redis)
2. Creates a `*gorm.DB` scoped to that host's `TablePrefix` via `utils.NewPrefixDb(prefix)`
3. Stores `hostInfo` and `db` in the Gin context (`c.Get("hostInfo")`, `c.Get("db")`)

All repository functions receive this prefixed DB. Tenants are fully isolated by table prefix — the same MySQL instance serves all tenants.

### Route Permission Levels
Defined in [core/common/web_routes.go](core/common/web_routes.go):

| Prefix | Auth | Roles |
|--------|------|-------|
| `/api/v1/` | None | Public |
| `/api/v1/outside` | JWT | 1,2,3,4 (all authenticated users) |
| `/api/v1/manager` | JWT | 1,2 (managers + superadmin) |
| `/api/v1/admin` | JWT | 1 (superadmin only) |
| `/api/v1/tenant` | JWT | Tenant users only |
| `/api/v1/app` | None | Mobile app endpoints |

### Configuration Files
- **`core.yaml`** — Infrastructure: MySQL master/slave, Redis, RabbitMQ, Aliyun OSS, Cloudflare R2, Telegram bot token
- **`cs.yaml`** — Seed data: default host, admin credentials, roles, menus, invite codes
- **`sc.yaml`** — Scheduler config; merged into `utils.CsConfig` at startup

Global config is accessed via `utils.GlobalConfig` (type `base.CoreConfig`) and `utils.CsConfig`.

### Key Globals (`core/utils/`)
- `utils.Db` — Base `*gorm.DB` (no prefix); use `utils.NewPrefixDb(prefix)` for tenant-scoped queries
- `utils.RD` — Redis client
- `utils.GlobalConfig` — Loaded from `core.yaml`
- `utils.CsConfig` — Loaded from `cs.yaml` / `sc.yaml`

### Middleware Execution Order
1. `hostInfoMiddleware` — resolves host, creates prefixed DB, sets `hostInfo`/`db` in context
2. CORS middleware
3. `authMiddleware` / `tenantAuthMiddleware` — JWT validation, role check
4. `manageLog` (on write routes) — captures request/response, publishes audit log via RabbitMQ

### Telegram Integration
The Telegram bot is initialized in `main.go` if `GlobalConfig.Telegram.Enabled` is true. Bot handlers live in `app/services/` and use polling. Telegram Mini-App authentication (Web App login) is handled by `core/api/tg_auth_api.go` + `core/utils/tg_auth_utils.go`, which validates the `web_app_data` HMAC signature from Telegram.

### Red Packet (Lucky Money) Feature
Core business logic in `core/services/lucky_money_service.go`. Key mechanics:
- Balance deducted atomically using Redis distributed locks (prevents race conditions on grab)
- "Thunder" (`雷`) system: a configurable losing index causes the grabber to lose their amount
- Distribution algorithm in `core/utils/lucky_money_utils.go`

### Adding a New API Endpoint
1. Add POJO/model in `core/pojo/` if needed
2. Add repository function in `core/repository/`
3. Add handler in `core/api/`
4. Register route in `core/common/web_routes.go` under the appropriate permission group
5. Run `swag init` to regenerate Swagger docs
