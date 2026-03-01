# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

This is the **mobile H5 app** (红包/lucky money) for the BaseGoUni platform. It is a Vue 3 + TypeScript mobile-first SPA using Vant UI and UnoCSS, designed as a Telegram Mini-App / mobile web page. The backend API docs are in the parent directory.

## Commands

Requires **pnpm >=10** and **Node >=20.19.0**.

```bash
pnpm install        # Install dependencies
pnpm dev            # Dev server at http://localhost:3000 (with mock API)
pnpm build:dev      # Development build with type check → dist/
pnpm build:pro      # Production build with type check → dist/
pnpm preview        # Preview built output
pnpm typecheck      # Vue + TypeScript type check
pnpm lint           # ESLint check
pnpm lint:fix       # ESLint auto-fix
```

### Environment
- `.env` — shared: `VITE_APP_API_BASE_URL=/api`, `VITE_APP_VCONSOLE=false`
- `.env.development` — local overrides
- `.env.production` — production overrides
- Dev port: **3000**; `/api` requests are proxied to the backend (see `vite.config.ts` server.proxy)
- `VITE_APP_VCONSOLE=true` enables the vConsole mobile debug panel

## Architecture

### Request Flow
```
src/api/<module>.ts  →  axios (VITE_APP_API_BASE_URL)  →  /api/v1/app/...  (backend)
```
The `/api` prefix is proxied in dev. In production, Nginx rewrites `/api` to the backend host.

### File-based Routing
Pages live in `src/pages/` — `unplugin-vue-router` auto-generates routes from the file tree. No manual route registration is needed.

```
src/pages/
├── index.vue          →  /
├── [...all].vue       →  404 catch-all
├── login/             →  /login/...
├── recharge/          →  /recharge/...
└── withdraw/          →  /withdraw/...
```

Route guards are in `src/router/index.ts`: NProgress, keep-alive tracking, page title, and user-info loading on login.

### State (Pinia + persisted state)
| Store | Purpose |
|-------|---------|
| `useUserStore` | User info, auth token — persisted to localStorage |
| `useRouteCacheStore` | Keep-alive component name list |

All stores under `src/stores/modules/` are auto-persisted via `pinia-plugin-persistedstate`.

### CSS — UnoCSS (not Tailwind)
This project uses **UnoCSS** atomic classes (e.g., `flex`, `w-full`, `text-sm`). Do **not** use Tailwind class names — the purge rules differ. Component-scoped styles use **Less** (`src/styles/var.less` for variables).

### Vant UI Component Usage
Vant components are **auto-imported** via `unplugin-vue-components` — no manual `import` needed in templates. However, functional components (Toast, Dialog, Notify, ImagePreview) must be imported from `vant` in `<script>` when called programmatically; their CSS is imported globally in `main.ts`.

### Mock Server
During `pnpm dev`, `vite-plugin-mock-dev-server` intercepts requests matching `/api/**`. Mock definitions go in `mock/` directory. The mock server runs on port **8086**.

### Key Conventions
- **Composables** in `src/composables/` — shared logic extracted as Vue composables
- **Auth token** managed by `src/utils/auth.ts`
- **i18n** keys live in `src/locales/`; use `useI18n()` composable in components
- **Page title** set via `setPageTitle()` from `src/utils/set-page-title.ts`
- TypeScript route types are auto-generated to `src/types/typed-router.d.ts` — do not edit manually

### Adding a New Page
1. Create `src/pages/<feature>/index.vue` — the route `/feature` is created automatically
2. Add API calls in `src/api/` if needed
3. Add mock handler in `mock/` for development without backend
