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

# Skip GORM AutoMigrate on startup (useful when schema is already up-to-date)
BGU_SKIP_AUTO_MIGRATE=1 go run main.go
```

The app runs on port `8080` by default (mapped to host port `9001` in Docker). Swagger UI is available at `/swagger/index.html`. The `dist/` directory is served as static files at `/`.

## Architecture

### Layered Structure
```
API handler → Repository → GORM (MySQL, master/slave via dbresolver)
                         → Redis (cache/locks/online tracking)
                         → Asynq (deferred tasks, backed by Redis)
```

- **`core/`** — Framework-level code (reusable across tenants/hosts)
- **`app/`** — Application-specific extensions (Telegram bot, app-layer services, DB init)
- **`tenant/`** — Tenant-specific API handlers (mirrors core/api for tenant isolation)

### Multi-Tenancy via Table Prefix
Every request goes through `hostInfoMiddleware` which:
1. Resolves the incoming host to a `HostInfo` record (cached in Redis)
2. Creates a `*gorm.DB` scoped to that host's `TablePrefix` via `utils.NewPrefixDb(prefix)`
3. Stores `hostInfo` and `db` in the Gin context (`c.Get("hostInfo")`, `c.Get("db")`)

All repository functions receive this prefixed DB. Tenants are fully isolated by table prefix — the same MySQL instance serves all tenants.

Database initialization (table auto-migration + sharding setup) lives in `app/utils/db_utils.go:InitDb()`, not `core/utils`. Set `BGU_SKIP_AUTO_MIGRATE=1` to skip `AutoMigrate` on startup.

### Route Permission Levels
Defined in [core/common/web_routes.go](core/common/web_routes.go):

| Prefix | Auth | Roles |
|--------|------|-------|
| `/api/v1/` | None | Public |
| `/api/v1/outside` | JWT (`authMiddleware`) | 1,2,3,4 (all admin users) |
| `/api/v1/manager` | JWT (`authMiddleware`) | 1,2 (managers + superadmin) |
| `/api/v1/admin` | JWT (`authMiddleware`) | 1 (superadmin only) |
| `/api/v1/tenant` | JWT (`tenantAuthMiddleware`) | Tenant users only |
| `/api/v1/app` | None or `appAuthMiddle` | Mobile app endpoints |

Admin/manager JWT uses `CsConfig.DefaultHost.AccessSecret`. App (mobile) JWT uses `HostInfo.AccessSecret` (per-host secret).

### Configuration Files
- **`core.yaml`** — Infrastructure: MySQL master/slave, Redis, Aliyun OSS, Cloudflare R2, Telegram bot token
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
3. `authMiddleware` / `tenantAuthMiddleware` / `appAuthMiddle` — JWT validation, role check
4. `manageLog` (on write routes) — captures request/response and writes the audit log to the tenant database

### Async Task Queue (Asynq)
`core/services/lucky_expire_task.go` initializes an Asynq client+server backed by the same Redis instance. Registered task types:
- `lucky:expire` — refund expired lucky money packets
- `lucky:bot_grab` — simulate bot grabbing a packet
- `trial_lucky:bot_grab` — trial bot grabs
- `telegram:welcome_message` / `telegram:delete_message` — Telegram bot messages
- `recharge:first_gift_installment` — first-recharge gift payout

Enqueue helpers follow the pattern `EnqueueXxxTask(tablePrefix, id, processAt)`. The server runs with concurrency 5 on a single `default` queue.

### Payment Provider Plugin System
`core/pay/provider.go` defines `Provider` (payin) and `PayoutProvider` (payout) interfaces. Each channel registers itself via `pay.Register(p)` in its package `init()`. Import the channel package with a blank import in `main.go` to activate it (e.g., `_ "BaseGoUni/core/pay/gctpk"`). Payment callbacks arrive at `/api/v1/pay/<channel>/notify` and `/api/v1/pay/<channel>/payoutNotify`.

### Third-Party Game Wallet API
`core/game/hgGame.go` implements the wallet-mode game integration. Two public (no JWT) endpoints in `web_routes.go` handle balance queries and transfer calls initiated by the game platform:
- `POST /Cash/Get` — query player balance
- `POST /Cash/TransferInOut` — credit/debit player balance

These use HMAC signature validation via `X-Sign`, `X-Request-Id`, and `X-Appid` headers.

### CashHistory Sharding
`CashHistory` records are sharded across `N` tables (`cash_history_0` … `cash_history_N-1`) by `user_id` hash. A GORM callback registered in `app/utils/shard_db_utils.go:InitShardingHook()` routes reads/writes automatically. A `all_cash_history` SQL view unions all shards for cross-shard queries.

### Scheduled Tasks
`core/common/common_scheduler.go` uses `robfig/cron` with distributed Redis locks to prevent duplicate execution across instances:
- Every 10 s — flush cached `HostInfo`
- Every 60 s — sweep expired lucky packets across all hosts
- Every 60 s (offset 30 s) — ensure minimum active trial bot packets

Scheduler only starts when `CsConfig.RunScheduler = true`.

### WebSocket
`core/common/ws_notify.go` starts a hub (`startWsHub()`). `core/utils/ws_notify_hook.go` exposes a hook-injection pattern so the `core` package can send WS messages without importing `common` (avoiding import cycles): call `utils.RegisterWsNotify`, `utils.RegisterWsBroadcast`, `utils.RegisterWsNotifyUser` at startup to wire up the hub.

### Telegram Integration
The Telegram bot is initialized in `main.go` if `GlobalConfig.Telegram.Enabled` is true. Bot handlers live in `app/services/` and use polling. Telegram Mini-App authentication (Web App login) is handled by `core/api/tg_auth_api.go` + `core/utils/tg_auth_utils.go`, which validates the `web_app_data` HMAC signature from Telegram.

### Red Packet (Lucky Money) Feature
Core business logic in `core/services/lucky_money_service.go`. Key mechanics:
- Balance deducted atomically using Redis distributed locks (prevents race conditions on grab)
- "Thunder" (`雷`) system: a configurable losing index causes the grabber to lose their amount
- Distribution algorithm in `core/utils/lucky_money_utils.go`
- Expiry handled by Asynq (`lucky:expire` task) + cron sweep fallback

### i18n
Translation files live in `core/locales/` (en, pt-BR, es-MX, id). Initialized via `utils.InitI18n()` at startup.

### Adding a New API Endpoint
1. Add POJO/model in `core/pojo/` if needed
2. Add repository function in `core/repository/`
3. Add handler in `core/api/`
4. Register route in `core/common/web_routes.go` under the appropriate permission group
5. Run `swag init` to regenerate Swagger docs
