# AGENTS.md

This file provides guidance to Codex and other coding agents when working in this repository.

> This file covers the `RedH5V2` frontend only. The parent repository has a backend-focused `AGENTS.md`; do not apply backend-only Go instructions to this Vue H5 project.

## Commands

```bash
# Install dependencies
pnpm install

# Start local development server
pnpm dev

# Type check only
pnpm typecheck

# Unit tests
pnpm test:unit

# Production build
pnpm build

# Staging / production mode builds
pnpm build:staging
pnpm build:production

# Validate dist artifacts after build
pnpm check:dist

# Preview built dist
pnpm preview
```

The dev server defaults to port `5173` and proxies `/api` to `VITE_API_PROXY_TARGET` or `http://127.0.0.1:8080`.

## Tech Stack

- Vue 3 with `<script setup>` and Composition API
- Vite
- TypeScript
- pnpm
- Vant UI
- Tailwind CSS v4 through `@tailwindcss/vite`
- Pinia
- `vue-router` file-based routing through `vue-router/vite`
- `unplugin-auto-import` and `unplugin-vue-components`
- `vue-i18n`
- `vite-plugin-pwa` and Workbox
- Axios HTTP client
- Vitest for focused unit tests

## Project Structure

```text
src/
  api/                 API modules
  assets/styles/       global CSS, layout CSS, and theme CSS
  components/          shared Vue components
  composables/         reusable Composition API utilities
  locales/             i18n messages and setup data
  pages/               file-based route pages
  plugins/             app plugin setup
  router/              router guards and page tracking
  stores/              Pinia stores
  stores/modules/      optional auto-imported store modules
  types/               shared TS types and generated declarations
  utils/               HTTP, runtime config, storage, tracking, money helpers
scripts/
  check-dist.mjs       build artifact validation
public/
  offline.html         PWA offline fallback
```

## Routing

Routes are generated from `src/pages` by `vue-router/vite`.

Use `definePage()` in page components for route metadata:

```ts
definePage({
  name: 'home',
  meta: {
    title: '红包大厅',
    keepAlive: true,
    tabbar: true,
    requiresAuth: false,
  },
})
```

Route meta fields are declared in `src/types/router-meta.d.ts`:

- `title`: page title used by `AppShell`
- `requiresAuth`: guarded by `src/router/index.ts`
- `keepAlive`: included by the root `KeepAlive`
- `tabbar`: shows `AppTabbar` on mobile only

Keep page names stable when they are listed in `src/App.vue` `keepAliveNames`.

## Layout And Theme

The app supports automatic mobile/PC layout switching:

- `src/composables/useResponsiveLayout.ts` uses `matchMedia('(min-width: 1024px)')`
- `<1024px` uses mobile H5 layout
- `>=1024px` uses PC layout
- `src/components/AppShell.vue` owns PC topbar, context bar, content shell, and mobile tabbar placement

Styles are intentionally split:

- `src/assets/styles/theme.css`
  - PP.PE inspired visual theme
  - dark obsidian background
  - red/gold brand tokens
  - glass surfaces
  - Vant CSS variable overrides
  - visual-only utilities such as `glass-panel`, `soft-badge`, `eyebrow`
- `src/assets/styles/main.css`
  - Tailwind import
  - global box sizing
  - app shell layout
  - page layout
  - grid and responsive breakpoints

When adding styles, keep visual tokens and color treatment in `theme.css`; keep dimensions, layout, and media queries in `main.css` unless a component has a clearly local scoped style.

Do not add new generic gradients, decorative blobs, or unrelated palettes. The current default visual direction is dark, premium, red/gold, glassmorphism, with `8px` radius.

## Auto Imports

Vite config auto-imports:

- Vue APIs
- Vue Router APIs
- Pinia APIs
- VueUse APIs
- `useI18n`
- composables under `src/composables`
- stores under `src/stores`
- module stores under `src/stores/modules`

Vant components are auto-resolved. Prefer using Vant components directly in templates without manual imports unless a local pattern requires otherwise.

Generated files:

- `src/types/auto-imports.d.ts`
- `src/types/components.d.ts`
- `src/types/typed-router.d.ts`

These may change after dev/build/typecheck tooling runs. Do not hand-edit generated declarations unless repairing generator output is the explicit task.

## State Management

Use Pinia setup stores. Existing stores:

- `src/stores/app.ts`
- `src/stores/user.ts`

New stores can be added under `src/stores` or `src/stores/modules`; both are configured for auto import.

For auth-sensitive changes, keep token storage behavior consistent between the user store and `src/utils/http.ts`. HTTP injects `Authorization: Bearer <token>` when request meta `auth` is enabled.

## HTTP And API Rules

Use `src/utils/http.ts` for API calls. It provides:

- `request<T>()`
- `get<T>()`
- `post<T>()`
- `put<T>()`
- `patch<T>()`
- `del<T>()`
- `upload<T>()`
- `download()`
- `HttpError`

Default behavior:

- unwraps backend envelopes and returns `data`
- shows Vant error toast unless `meta.showError === false`
- injects auth unless `meta.auth === false`
- supports request dedupe, retry, envelope/response return modes
- tracks API errors through `trackApiError()`
- clears auth and redirects on `401`

Do not create a second Axios instance unless there is a concrete isolation reason. Add API wrappers in `src/api`.

API failure policy:

- Do not silently fall back to static, mock, or hardcoded business data when an API request fails.
- Failed or empty API-backed views must render the project state/empty/error UI, such as `AppState`, with a retry path where useful.
- Static data is acceptable only for intentional local content such as navigation labels, theme assets, demo-only tests, or explicit design placeholders; it must not mask production API failure.

## Commercial H5 Utilities

Current reusable business infrastructure:

- Money formatting: `src/utils/money.ts`
- TTL local storage: `src/utils/storage.ts`, `src/composables/useTtlLocalStorage.ts`
- Submit lock: `src/composables/useSubmitLock.ts`
- Order polling: `src/composables/useOrderPolling.ts`
- Page state: `src/composables/usePageState.ts`
- Runtime config and force update: `src/utils/runtime.ts`
- Tracking queue and event reporting: `src/utils/tracker.ts`
- PWA registration and install/update flow: `src/pwa.ts`

Use these before adding new ad hoc implementations.

## PWA And Runtime Config

PWA is configured in `vite.config.ts` with `VitePWA`.

Important rules:

- API routes are denied from navigation fallback with `/^\/api\//`
- runtime cache is limited to static resource destinations
- `public/offline.html` is the offline fallback
- PWA assets live under `public`

Runtime commercial controls are read from `VITE_FORCE_UPDATE_URL` through `src/utils/runtime.ts`. The root `App.vue` displays maintenance or force-update states before rendering app routes.

## Environment Variables

Supported env variables include:

- `VITE_API_BASE_URL`
- `VITE_API_PROXY_TARGET`
- `VITE_CDN_BASE_URL`
- `VITE_APP_BASE`
- `VITE_TRACKING_ENDPOINT`
- `VITE_RELEASE_VERSION`
- `VITE_FORCE_UPDATE_URL`

`vite.config.ts` uses `VITE_CDN_BASE_URL || VITE_APP_BASE || '/'` as build base.

After production builds, run:

```bash
pnpm check:dist
```

## Testing And Verification

Before claiming frontend changes are complete, run the narrowest relevant checks and include the result in the response.

For most changes:

```bash
pnpm typecheck
pnpm build
```

For utility/composable behavior:

```bash
pnpm test:unit
```

For release/PWA/base-path changes:

```bash
pnpm build
pnpm check:dist
```

For visual layout changes, verify both:

- mobile viewport around `390x844`
- desktop viewport around `1440x900`

Use browser screenshots or DOM measurements when available. Confirm no horizontal overflow and that mobile-only UI such as `AppTabbar` does not appear on PC.

## Coding Conventions

- Prefer Composition API and `<script setup>`.
- Keep TypeScript types in `src/types/business.ts` when shared across modules.
- Keep API wrapper functions small and typed.
- Do not duplicate page loading/error/empty logic; use `usePageState()`.
- Do not duplicate request protection; use `useSubmitLock()`.
- Do not implement polling loops manually; use `useOrderPolling()`.
- Keep To C money and transaction display conservative: front-end protects interaction and display, backend remains source of truth.
- Keep cards at `8px` radius unless updating the theme intentionally.
- Avoid nested cards and marketing-only hero sections for operational H5 views.

## Git And Workspace Notes

This project may live inside a larger monorepo with unrelated sibling changes. When editing `RedH5V2`, only touch files inside this directory unless the user explicitly asks otherwise.

Do not revert unrelated changes in sibling projects such as `RedPocketH5`, `RedTenantAdmin`, or `pure-admin-thin`.
