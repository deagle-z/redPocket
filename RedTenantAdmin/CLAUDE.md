# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

This is the **tenant management panel** for the BaseGoUni platform. It is a Vue 3 + TypeScript SPA using Element Plus, targeting tenant users authenticated via `/api/v1/tenant/login`. It shares the same base framework (pure-admin v5.9.0) as `pure-admin-thin` but exposes tenant-scoped operations only (no superadmin features). The backend API docs are in the parent directory.

## Commands

Requires **pnpm >=9** and **Node ^18.18.0 || ^20.9.0 || >=22.0.0**.

```bash
pnpm install          # Install dependencies
pnpm dev              # Dev server at http://localhost:8848
pnpm build            # Production build → dist/
pnpm build:staging    # Staging build
pnpm typecheck        # TypeScript type check (no emit)
pnpm lint             # ESLint + Prettier + Stylelint
pnpm clean:cache      # Delete .cache, node_modules, dist
```

### Environment
- `.env` — shared defaults; `VITE_BASE_URL` points to the backend
- `.env.development` — overrides for local dev (`VITE_BASE_URL=http://127.0.0.1:8080`)
- `VITE_ROUTER_HISTORY` — `"hash"` (default) or `"h5"`
- Dev port: **8848** (same as pure-admin-thin — run only one at a time, or change `VITE_PORT`)

## Architecture

This project is a **fork of pure-admin-thin** scoped to tenant operations. All architectural patterns are identical — refer to [pure-admin-thin/CLAUDE.md](../pure-admin-thin/CLAUDE.md) for the shared patterns (routing, Pinia stores, HTTP client, permission components, icon system).

### Key Differences from pure-admin-thin

| Aspect | pure-admin-thin | RedTenantAdmin |
|--------|----------------|----------------|
| Login endpoint | `/api/v1/user/login` | `/api/v1/tenant/login` |
| API group | `/api/v1/admin/*` | `/api/v1/tenant/*` |
| Auth middleware | `authMiddleware` (roles 1–4) | `tenantAuthMiddleware` |
| Unique views | — | `src/views/cash_history/` |

### Request Flow
```
src/api/<module>.ts  →  src/utils/http  →  VITE_BASE_URL + /api/v1/tenant/...
```

### Tenant-specific API Modules (`src/api/`)
All 14 API modules exist here but call `/api/v1/tenant/*` endpoints instead of `/api/v1/admin/*`. The tenant token encodes the tenant user identity — the backend resolves the correct table prefix automatically.

### Adding a New Page
1. Create `src/views/<feature>/index.vue`
2. Add a route file `src/router/modules/<feature>.ts`
3. Add an API module `src/api/<feature>.ts` calling `/api/v1/tenant/<feature>`
4. Ensure the corresponding backend route exists in `tenantGroup` in `core/common/web_routes.go`
