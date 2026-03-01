# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

This is the **superadmin dashboard** for the BaseGoUni platform. It is a Vue 3 + TypeScript SPA using Element Plus, targeting role-1 (superadmin) and role-2 (manager) users. The backend API docs are in the parent directory.

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
- Dev port: **8848**; backend proxy is not used — `VITE_BASE_URL` is injected directly into the HTTP client

## Architecture

### Request Flow
```
src/api/<module>.ts  →  src/utils/http (axios wrapper)  →  VITE_BASE_URL + /api/v1/...
```
Each API module exports typed functions. The HTTP client reads `VITE_BASE_URL` at build time. All endpoints require a `Authorization: Bearer <token>` header, managed by the HTTP interceptors.

### Routing
Routes live in `src/router/modules/` as individual `.ts` files — Vite auto-imports all of them. The router flattens nested trees to max 2 levels for the sidebar menu. Non-menu routes (login, 404, redirect) are in `src/router/remaining.ts`.

Dynamic menus are built from the permission store (`src/store/modules/permission.ts`) which fetches `/api/v1/outside/routes` after login.

### State (Pinia)
| Store | Purpose |
|-------|---------|
| `useUserStore` | Current user info, login/logout |
| `usePermissionStore` | Dynamic routes, sidebar menus |
| `useMultiTagsStore` | Open tabs (multi-tab navigation) |

### Key Conventions
- **API modules** follow `src/api/<domain>.ts` — each exports typed `get*`, `set*`, `del*` functions matching backend REST conventions
- **Permission components** `<ReAuth>` / `<RePerms>` gate UI elements by role; import from `@/components/ReAuth` and `@/components/RePerms`
- **Icons** use Iconify via `<IconifyIconOnline>` / `<IconifyIconOffline>` wrapped in `<ReIcon>`
- **Tables** use `@pureadmin/table` wrapper around Element Plus `el-table`
- Global config (site title, logo, etc.) is loaded at runtime from `public/platform-config.json` — edit that file, not `src/config/`

### Adding a New Page
1. Create `src/views/<feature>/index.vue`
2. Add a route file `src/router/modules/<feature>.ts` with `meta.roles` set to allowed role numbers
3. Add an API module `src/api/<feature>.ts`
4. The sidebar menu picks up the new route automatically on next login
