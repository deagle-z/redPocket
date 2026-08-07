# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

> **This file covers the Go backend only.** Frontend sub-projects each have their own CLAUDE.md — read that file instead when working inside one of these directories:
> - [pure-admin-thin/CLAUDE.md](pure-admin-thin/CLAUDE.md) — Superadmin dashboard (Vue 3 + Element Plus)
> - [RedTenantAdmin/CLAUDE.md](RedTenantAdmin/CLAUDE.md) — Tenant management panel (Vue 3 + Element Plus)
> - [RedH5V2/CLAUDE.md](RedH5V2/CLAUDE.md) — Current mobile H5 app (Vue 3 + Vant + pnpm)
> - [RedPocketH5/](RedPocketH5/) — Older H5 app; has only an `AGENTS.md`, no CLAUDE.md
>
> [AGENTS.md](AGENTS.md) is the Codex-facing sibling of this file. When you change backend architecture facts here, mirror them there.

## Commands

Go 1.24. Dependencies are resolved from the module cache — there is no checked-in `vendor/` directory (`go mod vendor` in the README is optional, for offline/Linux builds).

```bash
# Set up Go proxy (for Linux/China servers)
export GOPROXY=https://goproxy.cn,direct
export GONOSUMDB=*

# Build / vet
go build ./...
go vet ./...

# Install and generate Swagger docs
go install github.com/swaggo/swag/cmd/swag@latest
swag init

# Run the application (configs default to core.yaml, cs.yaml, sc.yaml)
go run main.go
go run main.go -cc core.yaml -cs cs.yaml -sc sc.yaml

# Skip GORM AutoMigrate on startup (useful when schema is already up-to-date)
BGU_SKIP_AUTO_MIGRATE=1 go run main.go

# Tests — whole package, or a single test by name
go test -v ./core/utils/...
go test -v ./core/services -run TestGenerateThunderIndexes
go test -v ./core/pay/vcpaypen -run TestSign

# Build and run with Docker
docker rm -f bgu-1 && docker rmi bgu-1 && docker build --build-arg BUILDKIT_INLINE_CACHE=1 --memory 1GB -t bgu-1 . && docker run -e TZ=Asia/Shanghai -p 9001:8080 --name bgu-1 --restart always -d bgu-1 && docker logs -t -f bgu-1
```

The app runs on port `8080` by default (mapped to host port `9001` in Docker). Swagger UI is available at `/swagger/index.html`. The `dist/` directory is served as static files at `/`.

**Verification convention:** after modifying Go backend code, do not run the Go test suite unless the user explicitly asks. Prefer `go build ./...` to confirm the change compiles.

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
3. `authMiddleware` / `tenantAuthMiddleware` / `appAuthMiddle` — JWT validation, role check
4. `manageLog` (on write routes) — captures request/response, publishes audit log via RabbitMQ

### Async Task Queue (Asynq)
`core/services/lucky_expire_task.go` initializes an Asynq client+server backed by the same Redis instance. Registered task types:
- `lucky:expire` — refund expired lucky money packets
- `lucky:bot_grab` — simulate bot grabbing a packet
- `trial_lucky:bot_grab` — trial bot grabs
- `telegram:welcome_message` / `telegram:delete_message` — Telegram bot messages
- `recharge:first_gift_installment` — first-recharge gift payout

Enqueue helpers follow the pattern `EnqueueXxxTask(tablePrefix, id, processAt)`. The server runs with concurrency 5 on a single `default` queue.

### Payment Provider Plugin System
`core/pay/provider.go` defines `Provider` (payin) and `PayoutProvider` (payout) interfaces, plus the shared `PayRequest`/`PayResponse`/`PayoutRequest`/`PayoutResponse` DTOs. `Provider.Name()` must return the channel code **in uppercase**, matching `sys_pay_channel.channel_code` in the DB — that string is how a configured channel row resolves to code.

Existing channels: `gctpk`, `gctpkBRL`, `gctpkmxn`, `vcpaymxn`, `vcpaypen` (each its own subpackage), plus `YoyopayProvider` in `core/pay/yoyopay_provider.go` — a stub test channel that returns `AutoSuccess: true` so orders settle immediately without a real gateway call.

Wiring a new channel touches **three** places, and missing any one fails silently rather than at compile time:
1. `pay.Register(&Provider{})` inside the subpackage's `init()`
2. A blank import in `main.go` (e.g. `_ "BaseGoUni/core/pay/gctpk"`) — without this the `init()` never runs and the channel is simply absent from the registry
3. Callback routes in `core/common/web_routes.go` under the `/api/v1/pay` group: `POST /<channel>/notify` (payin) and `POST /<channel>/payoutNotify` (payout), handled by `Xxx PayinCallback` / `XxxPayoutCallback` in `core/api/`

Amount handling: providers receive both `Amount` and `ProviderAmount`; use the `pay.ResolveOrderAmount` / `pay.ResolvePayoutAmount` helpers rather than reading the fields directly. Most gateways want minor units (cents) as an integer.

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
Translation files live in `core/locales/` (en, pt-BR, es-PE, id). Initialized via `utils.InitI18n()` at startup.

### Adding a New API Endpoint
1. Add POJO/model in `core/pojo/` if needed
2. Add repository function in `core/repository/` — takes `db *gorm.DB` as its first arg, never touches `utils.Db` directly
3. Add handler in `core/api/` (or `tenant/api/` for tenant-scoped endpoints)
4. Register route in `core/common/web_routes.go` under the appropriate permission group
5. Run `swag init` to regenerate Swagger docs

Handlers follow a consistent shape — match it rather than inventing a new one:

```go
// GetPayChannels godoc
//
//	@Summary	获取支付通道列表
//	@Tags		支付通道
//	@Param		data body	pojo.PayChannelSearch	true	"查询条件"
//	@Success	200	{object}	pojo.PayChannelResp
//	@Router		/api/v1/admin/payChannel/list [post]
func GetPayChannels(ctx *gin.Context) {
	var search pojo.PayChannelSearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)          // tenant-prefixed DB from hostInfoMiddleware
	utils.SuccessObjBack(ctx, repository.GetPayChannels(db, search))
}
```

**Gin context keys** set by middleware: `db` (`*gorm.DB`, prefixed — used by essentially every handler), `userId`, `hostInfo`, `tenantId`, `token`.

**Response helpers** (`core/utils/common_utils.go`) — always return through these, never `ctx.JSON` directly: `SuccessObjBack(ctx, data)`, `SuccessBack(ctx, msg)`, `ErrorBack(ctx, msg)`, `ErrorObjBack(ctx, data, msg)`, `ErrorMsgBack(ctx, msg)`.

Swagger godoc comment blocks above each handler are the source for `swag init`; keep the `@Router` path in sync with the actual registration in `web_routes.go`.
