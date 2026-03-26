# AGENTS.md

This file defines the working rules for coding agents in this repository.

## Project Summary

- Project type: mobile-first H5 SPA for BaseGoUni red packet / lucky money flows
- Stack: Vue 3, TypeScript, Vite, Vant, Pinia, vue-router, UnoCSS, Less
- Package manager: `pnpm`
- Runtime requirement: Node `>=20.19.0`, pnpm `>=10`
- Main app mode: Telegram Mini-App / mobile web page

## Core Commands

```bash
pnpm install
pnpm dev
pnpm build:dev
pnpm build:pro
pnpm preview
pnpm typecheck
pnpm lint
pnpm lint:fix
```

## Environment Notes

- Shared env: `.env`
- Dev overrides: `.env.development`
- Production overrides: `.env.production`
- Dev server is started with `pnpm dev`
- API base URL comes from `VITE_APP_API_BASE_URL`
- Mobile debug panel is controlled by `VITE_APP_VCONSOLE`
- WebSocket config is controlled by `VITE_WS_URL` / `VITE_APP_WS_URL` and `VITE_APP_WS_UID`

## Architecture

### Routing

- This project uses file-based routing via `unplugin-vue-router`
- Pages live under `src/pages/`
- Do not manually register normal page routes unless the codebase already requires it
- Router guards live in [src/router/index.ts](/c:/code/iiii/red/BaseGoUni/RedPocketH5/src/router/index.ts)
- Public routes currently include `/`, `/login`, `/register`, `/resetpwd`, and `/luckyDetail`

### State

- Stores live under `src/stores/modules/`
- Pinia persisted state is enabled
- Auth and current user state are managed through `useUserStore`
- Route keep-alive state is managed through `useRouteCacheStore`

### API Layer

- API types and request functions are centralized in [src/api/user.ts](/c:/code/iiii/red/BaseGoUni/RedPocketH5/src/api/user.ts)
- Before creating a new API file, prefer extending `src/api/user.ts` to match the current codebase convention
- Axios wrapper lives in [src/utils/request.ts](/c:/code/iiii/red/BaseGoUni/RedPocketH5/src/utils/request.ts)
- The request layer automatically:
  - injects `Authorization` from local storage
  - handles HTTP errors with Vant `showToast`
  - treats business responses with `success === false` or `code` not in `[0, 200]` as failures
- Callers should not duplicate generic toast-based business error handling unless there is a page-specific need

### WebSocket

- WebSocket code lives under `src/plugins/websocket/`
- Reuse the existing singleton client instead of creating parallel WebSocket implementations

## UI and Styling Rules

- This project uses UnoCSS, not Tailwind
- Prefer existing UnoCSS utility patterns already used in nearby files
- Component-scoped styles should use Less when scoped styling is needed
- Theme tokens are defined in [src/styles/themes/app-theme.css](/c:/code/iiii/red/BaseGoUni/RedPocketH5/src/styles/themes/app-theme.css)
- Do not introduce bare hex colors into page or component scoped styles when an existing theme token can be used
- Reuse existing color, radius, spacing, and shadow variables before adding new ones
- Keep the UI mobile-first and aligned with Vant interaction patterns

## Auth and Navigation Rules

- Token helpers are in `src/utils/auth.ts`
- Unauthenticated users should be redirected to login for protected pages
- If you add a page that should be public, update the public route logic in [src/router/index.ts](/c:/code/iiii/red/BaseGoUni/RedPocketH5/src/router/index.ts)
- Root tab-like routes are configured in [src/config/routes.ts](/c:/code/iiii/red/BaseGoUni/RedPocketH5/src/config/routes.ts)
- Current root route names are `Home`, `History`, `Team`, `Wallet`, `Profile`

## i18n and Content

- Locale files live in `src/locales/`
- Reuse existing i18n keys where possible
- When adding user-facing text on shared or major pages, update the locale files instead of hardcoding copy unless the surrounding page already hardcodes text

## File Handling Rules

- Treat generated type files under `src/types/` as generated unless the file is clearly meant for manual maintenance
- In particular, do not manually edit `src/types/typed-router.d.ts`
- Preserve existing naming, folder layout, and feature placement
- Prefer small targeted edits over broad refactors unless the task explicitly asks for restructuring

## Change Guidelines

- Match existing code style in the touched area
- Reuse existing utilities and components before adding new abstractions
- Avoid introducing a second pattern when the repo already has a clear primary pattern
- Keep API contracts, route names, and storage keys backward compatible unless the task requires a breaking change
- For new pages:
  1. add `src/pages/<feature>/index.vue`
  2. add or extend API functions in `src/api/user.ts` if needed
  3. add mock handlers in `mock/` if local development depends on backend simulation
  4. update `src/config/routes.ts` if the page is a root tab page

## Validation Checklist

Run the smallest relevant checks after changes:

- `pnpm typecheck` for TypeScript or Vue changes
- `pnpm lint` for lint validation
- `pnpm build:dev` or `pnpm build:pro` for integration-sensitive changes

If you cannot run validation, state that clearly in the handoff.

## Commit and Hook Notes

- Pre-commit runs `lint-staged`
- Commit messages are checked by `commitlint`
- Do not change hook behavior unless the task is explicitly about tooling

## Agent Behavior

- Read nearby files before editing
- Prefer code-backed truth over stale documentation when they differ
- Keep documentation updates synchronized with behavior changes when touching architecture, routes, env usage, or developer workflow
- Do not rewrite unrelated files
- Surface assumptions clearly if a task depends on backend behavior not present in this repository
